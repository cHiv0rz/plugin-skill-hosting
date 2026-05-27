package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"marketplace/internal/metrics"
)

func (a *App) issueToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(30 * 24 * time.Hour).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(a.Cfg.JWTSecret))
}

func (a *App) parseToken(tok string) (string, error) {
	parsed, err := jwt.Parse(tok, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.Cfg.JWTSecret), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || !parsed.Valid {
		return "", errors.New("invalid token")
	}
	sub, _ := claims["sub"].(string)
	if sub == "" {
		return "", errors.New("missing sub")
	}
	return sub, nil
}

// authenticateRequest accepts:
//   - Authorization: Bearer <jwt> — JWTs (browser sessions) — recognised by 2 dots
//   - Authorization: Bearer <api_token> — opaque per-user API token
//   - HTTP Basic Auth — password = api token (username ignored)
//
// Returns (user, "", nil) on success; (nil, msg, nil) on credential failure
// (msg is the 401 reason, "" when no credential was presented at all); and
// (nil, "", err) on an unexpected backend error (DB outage, etc.) so callers
// can map that to 500 instead of silently 401-ing legitimate clients.
func (a *App) authenticateRequest(r *http.Request) (*User, string, error) {
	if h := r.Header.Get("Authorization"); h != "" {
		if strings.HasPrefix(h, "Bearer ") {
			tok := strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
			if tok == "" {
				return nil, "empty bearer token", nil
			}
			return a.resolveToken(r.Context(), tok)
		}
		if strings.HasPrefix(h, "Basic ") {
			if _, pass, ok := r.BasicAuth(); ok && pass != "" {
				return a.resolveToken(r.Context(), pass)
			}
			return nil, "invalid basic auth", nil
		}
	}
	return nil, "", nil
}

// resolveToken resolves either a JWT or a raw API token to a user. Distinguishes
// "credential is bad" (msg set, err nil) from "DB lookup failed" (msg empty,
// err set) so an intermittent backend hiccup is not reported back to the
// caller as a 401.
func (a *App) resolveToken(ctx context.Context, tok string) (*User, string, error) {
	if strings.Count(tok, ".") == 2 {
		userID, err := a.parseToken(tok)
		if err != nil {
			return nil, "invalid token", nil
		}
		u, err := a.userByID(ctx, userID)
		if err == sql.ErrNoRows {
			return nil, "unknown user", nil
		}
		if err != nil {
			return nil, "", err
		}
		return u, "", nil
	}
	u, err := a.userByAPIToken(ctx, tok)
	if err == sql.ErrNoRows {
		return nil, "invalid token", nil
	}
	if err != nil {
		return nil, "", err
	}
	return u, "", nil
}

func (a *App) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, errMsg, err := a.authenticateRequest(r)
		if err != nil {
			serverErr(w, r, err, "auth lookup error")
			return
		}
		if u == nil {
			if errMsg == "" {
				errMsg = "missing bearer token"
			}
			writeErr(w, http.StatusUnauthorized, errMsg)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// requireApprovedMiddleware refuses requests from users whose account is
// pending approval or has been rejected. Must run AFTER authMiddleware (or
// any other middleware that puts a *User into the request context).
//
// /api/me deliberately bypasses this so the frontend can fetch the user's
// own status and route them to the "pending" page.
func (a *App) requireApprovedMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := currentUser(r)
		if u == nil {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		if u.Status != UserStatusApproved {
			writeErr(w, http.StatusForbidden, "account "+u.Status)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireAdminMiddleware refuses requests from non-admin users. Must run AFTER
// authMiddleware + requireApprovedMiddleware (the admin set is a subset of
// approved users). User-management endpoints are gated to admins; everything
// else stays open to any approved user.
func (a *App) requireAdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := currentUser(r)
		if u == nil {
			writeErr(w, http.StatusUnauthorized, "missing bearer token")
			return
		}
		if !u.IsAdmin {
			writeErr(w, http.StatusForbidden, "admin only")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// tokenGateMiddleware authenticates marketplace.json, /git/*, and the read-only
// plugin endpoints. On failure it sends WWW-Authenticate so `git clone` and
// curl prompt for credentials.
func (a *App) tokenGateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, _, err := a.authenticateRequest(r)
		if err != nil {
			serverErr(w, r, err, "auth lookup error")
			return
		}
		if u == nil {
			w.Header().Set("WWW-Authenticate", `Basic realm="plugin-marketplace"`)
			writeErr(w, http.StatusUnauthorized, "authentication required")
			return
		}
		if u.Status != UserStatusApproved {
			writeErr(w, http.StatusForbidden, "account "+u.Status)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// mcpTokenGateMiddleware is the /mcp variant: same Bearer/Basic acceptance as
// the regular gate, but the 401 challenge advertises Bearer rather than Basic.
// MCP clients use the WWW-Authenticate scheme to decide their auth UX, and a
// Basic challenge here pushes them into the OAuth-fallback path.
func (a *App) mcpTokenGateMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, errMsg, err := a.authenticateRequest(r)
		if err != nil {
			serverErr(w, r, err, "auth lookup error")
			return
		}
		if u == nil {
			w.Header().Set("WWW-Authenticate", `Bearer realm="plugin-marketplace"`)
			if errMsg == "" {
				errMsg = "authentication required"
			}
			writeErr(w, http.StatusUnauthorized, errMsg)
			return
		}
		if u.Status != UserStatusApproved {
			writeErr(w, http.StatusForbidden, "account "+u.Status)
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserKey, u)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func generateAPIToken() (string, error) {
	return randHex(32)
}

func (a *App) userByAPIToken(ctx context.Context, token string) (*User, error) {
	u := &User{}
	err := a.DB.QueryRowContext(ctx,
		`SELECT id, email, username, api_token, status, is_admin, created_at FROM users WHERE api_token = $1`, token).
		Scan(&u.ID, &u.Email, &u.Username, &u.APIToken, &u.Status, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (a *App) userByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := a.DB.QueryRowContext(ctx,
		`SELECT id, email, username, api_token, status, is_admin, created_at FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.Email, &u.Username, &u.APIToken, &u.Status, &u.IsAdmin, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

type registerReq struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (a *App) handleRegister(w http.ResponseWriter, r *http.Request) {
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))
	req.Username = strings.TrimSpace(req.Username)
	if !strings.Contains(req.Email, "@") {
		writeErr(w, http.StatusBadRequest, "invalid email")
		return
	}
	if !usernameRe.MatchString(req.Username) {
		writeErr(w, http.StatusBadRequest, "username must be 3-32 chars, alphanumeric/_/-")
		return
	}
	if len(req.Password) < 8 {
		writeErr(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		serverErr(w, r, err, "hash error")
		return
	}

	apiTok, err := generateAPIToken()
	if err != nil {
		serverErr(w, r, err, "token error")
		return
	}

	// is_admin is computed in SQL so the empty-DB bootstrap is decided
	// atomically with the INSERT (matches the OIDC create path's
	// approved-bootstrap pattern). The first ever user lands as admin so the
	// /users page is operable immediately on a fresh deployment.
	var (
		id      string
		isAdmin bool
	)
	err = a.DB.QueryRowContext(r.Context(),
		`INSERT INTO users (email, username, password_hash, api_token, is_admin)
		 VALUES ($1, $2, $3, $4, NOT EXISTS (SELECT 1 FROM users))
		 RETURNING id, is_admin`,
		req.Email, req.Username, string(hash), apiTok).Scan(&id, &isAdmin)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			writeErr(w, http.StatusConflict, "email or username already in use")
			return
		}
		serverErr(w, r, err, "db error")
		return
	}

	tok, err := a.issueToken(id)
	if err != nil {
		serverErr(w, r, err, "token error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": tok,
		"user": User{
			ID:       id,
			Email:    req.Email,
			Username: req.Username,
			APIToken: apiTok,
			Status:   UserStatusApproved,
			IsAdmin:  isAdmin,
		},
	})
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (a *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	var (
		id, username, hash, apiTok, status string
		isAdmin                            bool
	)
	err := a.DB.QueryRowContext(r.Context(),
		`SELECT id, username, password_hash, api_token, status, is_admin FROM users WHERE email = $1`, req.Email).
		Scan(&id, &username, &hash, &apiTok, &status, &isAdmin)
	if err == sql.ErrNoRows {
		metrics.LoginsTotal.WithLabelValues("password", "failure").Inc()
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		metrics.LoginsTotal.WithLabelValues("password", "failure").Inc()
		serverErr(w, r, err, "db error")
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		metrics.LoginsTotal.WithLabelValues("password", "failure").Inc()
		writeErr(w, http.StatusUnauthorized, "invalid credentials")
		return
	}
	metrics.LoginsTotal.WithLabelValues("password", "success").Inc()

	tok, err := a.issueToken(id)
	if err != nil {
		serverErr(w, r, err, "token error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": tok,
		"user": User{
			ID:       id,
			Email:    req.Email,
			Username: username,
			APIToken: apiTok,
			Status:   status,
			IsAdmin:  isAdmin,
		},
	})
}

func (a *App) handleMe(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, currentUser(r))
}

func (a *App) handleRegenerateAPIToken(w http.ResponseWriter, r *http.Request) {
	user := currentUser(r)
	newTok, err := generateAPIToken()
	if err != nil {
		serverErr(w, r, err, "token error")
		return
	}
	if _, err := a.DB.ExecContext(r.Context(),
		`UPDATE users SET api_token = $1 WHERE id = $2`, newTok, user.ID); err != nil {
		serverErr(w, r, err, "db error")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"apiToken": newTok})
}

type authConfigResp struct {
	Mode                 string `json:"mode"`
	MarketplaceName      string `json:"marketplaceName"`
	DefaultLicense       string `json:"defaultLicense"`
	UserApprovalRequired bool   `json:"userApprovalRequired"`
}

func (a *App) handleAuthConfig(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, authConfigResp{
		Mode:                 a.Cfg.AuthMode,
		MarketplaceName:      a.Cfg.MarketplaceName,
		DefaultLicense:       a.Cfg.DefaultLicense,
		UserApprovalRequired: a.Cfg.RequiresUserApproval(),
	})
}
