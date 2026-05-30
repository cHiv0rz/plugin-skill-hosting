package server

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"marketplace/internal/config"
)

// oauthTestApp returns an App with OAuth enabled but no DB. Every assertion in
// this file deliberately exercises a code path that returns before any DB
// access, so a nil DB is fine — the happy-path code exchange and refresh
// rotation are the only paths that touch Postgres and are out of scope here.
func oauthTestApp() *App {
	return &App{Cfg: config.Config{
		JWTSecret:            "test-secret-do-not-use",
		PublicBaseURL:        "https://mp.example.com",
		AuthMode:             "password",
		MCPOAuthClientID:     "test-client",
		MCPOAuthClientSecret: "test-client-secret",
		MCPOAuthRedirectURIs: []string{"https://claude.ai/api/mcp/auth_callback"},
	}}
}

// pkceChallenge returns BASE64URL(SHA256(verifier)) — the S256 transform.
func pkceChallenge(verifier string) string {
	h := sha256.Sum256([]byte(verifier))
	return base64.RawURLEncoding.EncodeToString(h[:])
}

// --- PKCE -------------------------------------------------------------------

func TestVerifyCodeChallenge(t *testing.T) {
	const verifier = "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	challenge := pkceChallenge(verifier)
	if !verifyCodeChallenge(verifier, challenge) {
		t.Error("verifyCodeChallenge rejected a matching verifier/challenge pair")
	}
	if verifyCodeChallenge("not-the-verifier", challenge) {
		t.Error("verifyCodeChallenge accepted a mismatched verifier")
	}
	if verifyCodeChallenge(verifier, "") {
		t.Error("verifyCodeChallenge accepted an empty challenge")
	}
}

// --- redirect_uri allowlist -------------------------------------------------

func TestValidRedirectURI(t *testing.T) {
	a := oauthTestApp()
	if !a.validRedirectURI("https://claude.ai/api/mcp/auth_callback") {
		t.Error("validRedirectURI rejected the configured URI")
	}
	// Exact match only — no prefix or substring matching.
	for _, bad := range []string{
		"https://claude.ai/api/mcp/auth_callback/extra",
		"https://claude.ai/api/mcp/auth_callback?x=1",
		"https://claude.ai/api/auth/callback", // trimmed from defaults
		"https://evil.example.com/cb",
		"",
	} {
		if a.validRedirectURI(bad) {
			t.Errorf("validRedirectURI accepted disallowed URI %q", bad)
		}
	}
}

// --- discovery metadata -----------------------------------------------------

func TestHandleOAuthMeta(t *testing.T) {
	a := oauthTestApp()
	rec := httptest.NewRecorder()
	a.handleOAuthMeta(rec, httptest.NewRequest("GET", "/.well-known/oauth-authorization-server", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if body["issuer"] != "https://mp.example.com" {
		t.Errorf("issuer = %v", body["issuer"])
	}
	if body["authorization_endpoint"] != "https://mp.example.com/oauth/authorize" {
		t.Errorf("authorization_endpoint = %v", body["authorization_endpoint"])
	}
	if body["token_endpoint"] != "https://mp.example.com/oauth/token" {
		t.Errorf("token_endpoint = %v", body["token_endpoint"])
	}
}

func TestHandleOAuthMeta_DisabledIs404(t *testing.T) {
	a := &App{Cfg: config.Config{PublicBaseURL: "https://mp.example.com"}} // no client id
	rec := httptest.NewRecorder()
	a.handleOAuthMeta(rec, httptest.NewRequest("GET", "/.well-known/oauth-authorization-server", nil))
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404 when OAuth disabled", rec.Code)
	}
}

// handleOAuthProtectedResource (the RFC 9728 document) and the /mcp challenge's
// resource_metadata pointer are already covered in auth_test.go.

// --- /mcp 401 challenge -----------------------------------------------------

func TestMCPAuthChallenge(t *testing.T) {
	a := oauthTestApp()
	got := a.mcpAuthChallenge()
	// Must point clients at the RFC 9728 resource-metadata document so they can
	// discover the authorization server.
	if !strings.Contains(got, `resource_metadata="https://mp.example.com/.well-known/oauth-protected-resource/mcp"`) {
		t.Errorf("challenge missing resource_metadata pointer: %q", got)
	}
	if !strings.HasPrefix(got, `Bearer realm="plugin-marketplace"`) {
		t.Errorf("challenge missing Bearer realm: %q", got)
	}
}

func TestMCPAuthChallenge_DisabledIsBareBearer(t *testing.T) {
	a := &App{Cfg: config.Config{PublicBaseURL: "https://mp.example.com"}}
	if got := a.mcpAuthChallenge(); got != `Bearer realm="plugin-marketplace"` {
		t.Errorf("challenge = %q, want bare Bearer realm when OAuth disabled", got)
	}
}

// --- authorize validation (GET) ---------------------------------------------

func authorizeQuery(overrides map[string]string) string {
	q := url.Values{
		"client_id":             {"test-client"},
		"redirect_uri":          {"https://claude.ai/api/mcp/auth_callback"},
		"state":                 {"xyz"},
		"code_challenge":        {pkceChallenge("verifier")},
		"code_challenge_method": {"S256"},
		"response_type":         {"code"},
	}
	for k, v := range overrides {
		if v == "" {
			q.Del(k)
		} else {
			q.Set(k, v)
		}
	}
	return q.Encode()
}

func TestHandleOAuthAuthorize_InvalidClientID(t *testing.T) {
	a := oauthTestApp()
	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/oauth/authorize?"+authorizeQuery(map[string]string{"client_id": "wrong"}), nil)
	a.handleOAuthAuthorize(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	// A bad client_id must NOT redirect — that would be an open redirect vector.
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Errorf("unexpected redirect on bad client_id: %q", loc)
	}
}

func TestHandleOAuthAuthorize_InvalidRedirectURI(t *testing.T) {
	a := oauthTestApp()
	rec := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/oauth/authorize?"+authorizeQuery(map[string]string{"redirect_uri": "https://evil.example.com/cb"}), nil)
	a.handleOAuthAuthorize(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "" {
		t.Errorf("unexpected redirect on bad redirect_uri: %q", loc)
	}
}

// redirectErrorCode runs authorize and returns the "error" query param of the
// resulting redirect (or "" if it didn't redirect).
func redirectErrorCode(t *testing.T, a *App, query string) string {
	t.Helper()
	rec := httptest.NewRecorder()
	a.handleOAuthAuthorize(rec, httptest.NewRequest("GET", "/oauth/authorize?"+query, nil))
	if rec.Code != http.StatusFound {
		t.Fatalf("status = %d, want 302 redirect", rec.Code)
	}
	loc := rec.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse Location %q: %v", loc, err)
	}
	if got := u.Scheme + "://" + u.Host + u.Path; got != "https://claude.ai/api/mcp/auth_callback" {
		t.Errorf("redirect target = %q, want the registered redirect_uri", got)
	}
	if st := u.Query().Get("state"); st != "xyz" {
		t.Errorf("redirect dropped state, got %q", st)
	}
	return u.Query().Get("error")
}

func TestHandleOAuthAuthorize_UnsupportedResponseType(t *testing.T) {
	a := oauthTestApp()
	if code := redirectErrorCode(t, a, authorizeQuery(map[string]string{"response_type": "token"})); code != "unsupported_response_type" {
		t.Errorf("error = %q, want unsupported_response_type", code)
	}
}

func TestHandleOAuthAuthorize_MissingCodeChallenge(t *testing.T) {
	a := oauthTestApp()
	if code := redirectErrorCode(t, a, authorizeQuery(map[string]string{"code_challenge": ""})); code != "invalid_request" {
		t.Errorf("error = %q, want invalid_request", code)
	}
}

func TestHandleOAuthAuthorize_NonS256(t *testing.T) {
	a := oauthTestApp()
	if code := redirectErrorCode(t, a, authorizeQuery(map[string]string{"code_challenge_method": "plain"})); code != "invalid_request" {
		t.Errorf("error = %q, want invalid_request", code)
	}
}

func TestHandleOAuthAuthorize_ValidRendersLoginForm(t *testing.T) {
	a := oauthTestApp()
	rec := httptest.NewRecorder()
	a.handleOAuthAuthorize(rec, httptest.NewRequest("GET", "/oauth/authorize?"+authorizeQuery(nil), nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Errorf("content-type = %q, want text/html", ct)
	}
	if !strings.Contains(rec.Body.String(), "Sign in") {
		t.Error("login form body missing expected content")
	}
}

// --- token endpoint ---------------------------------------------------------

func postToken(a *App, form url.Values, basicID, basicSecret string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "/oauth/token", strings.NewReader(form.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if basicID != "" || basicSecret != "" {
		r.SetBasicAuth(basicID, basicSecret)
	}
	rec := httptest.NewRecorder()
	a.handleOAuthToken(rec, r)
	return rec
}

func tokenErrorCode(t *testing.T, rec *httptest.ResponseRecorder) string {
	t.Helper()
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode token error: %v (body %q)", err, rec.Body.String())
	}
	return body.Error
}

func TestHandleOAuthToken_InvalidClient(t *testing.T) {
	a := oauthTestApp()
	rec := postToken(a, url.Values{"grant_type": {"authorization_code"}}, "test-client", "wrong-secret")
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", rec.Code)
	}
	if got := rec.Header().Get("WWW-Authenticate"); !strings.HasPrefix(got, "Basic") {
		t.Errorf("WWW-Authenticate = %q, want Basic challenge", got)
	}
	if code := tokenErrorCode(t, rec); code != "invalid_client" {
		t.Errorf("error = %q, want invalid_client", code)
	}
}

func TestHandleOAuthToken_UnsupportedGrant(t *testing.T) {
	a := oauthTestApp()
	rec := postToken(a, url.Values{"grant_type": {"client_credentials"}}, "test-client", "test-client-secret")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if code := tokenErrorCode(t, rec); code != "unsupported_grant_type" {
		t.Errorf("error = %q, want unsupported_grant_type", code)
	}
	// Token responses must not be cached (RFC 6749 §5.1).
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", cc)
	}
}

func TestHandleCodeExchange_MissingParams(t *testing.T) {
	a := oauthTestApp()
	// Valid client, authorization_code grant, but no code/verifier/redirect_uri.
	rec := postToken(a, url.Values{"grant_type": {"authorization_code"}}, "test-client", "test-client-secret")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if code := tokenErrorCode(t, rec); code != "invalid_request" {
		t.Errorf("error = %q, want invalid_request", code)
	}
}

func TestHandleRefreshToken_MissingToken(t *testing.T) {
	a := oauthTestApp()
	rec := postToken(a, url.Values{"grant_type": {"refresh_token"}}, "test-client", "test-client-secret")
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", rec.Code)
	}
	if code := tokenErrorCode(t, rec); code != "invalid_request" {
		t.Errorf("error = %q, want invalid_request", code)
	}
}

// --- access-token scope restriction -----------------------------------------

func TestMCPAccessTokenCarriesScopeClaim(t *testing.T) {
	a := oauthTestApp()
	mcpTok, err := a.issueMCPAccessToken("user-1")
	if err != nil {
		t.Fatalf("issueMCPAccessToken: %v", err)
	}
	_, typ, err := a.parseToken(mcpTok)
	if err != nil {
		t.Fatalf("parseToken: %v", err)
	}
	if typ != tokenTypeMCPAccess {
		t.Errorf("typ = %q, want %q", typ, tokenTypeMCPAccess)
	}
	// Ordinary session tokens stay full-access (no typ claim).
	sessTok, err := a.issueToken("user-1")
	if err != nil {
		t.Fatalf("issueToken: %v", err)
	}
	if _, typ, _ := a.parseToken(sessTok); typ != "" {
		t.Errorf("session token typ = %q, want empty", typ)
	}
}

func TestResolveTokenRejectsMCPScopeOutsideMCP(t *testing.T) {
	a := oauthTestApp()
	mcpTok, err := a.issueMCPAccessToken("user-1")
	if err != nil {
		t.Fatalf("issueMCPAccessToken: %v", err)
	}
	// allowMCPScope=false (the /api and /git gates) must reject the token. The
	// rejection happens before any DB lookup, so a nil DB never gets touched.
	u, msg, err := a.resolveToken(context.Background(), mcpTok, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if u != nil {
		t.Errorf("expected nil user, got %+v", u)
	}
	if msg != "token not valid for this endpoint" {
		t.Errorf("msg = %q, want token-not-valid rejection", msg)
	}
}

func TestAuthMiddlewareRejectsMCPAccessToken(t *testing.T) {
	a := oauthTestApp()
	mcpTok, err := a.issueMCPAccessToken("user-1")
	if err != nil {
		t.Fatalf("issueMCPAccessToken: %v", err)
	}
	called := false
	h := a.authMiddleware(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { called = true }))
	r := httptest.NewRequest("GET", "/api/me/token/regenerate", nil)
	r.Header.Set("Authorization", "Bearer "+mcpTok)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, r)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 — an mcp_access token must not reach /api", rec.Code)
	}
	if called {
		t.Error("downstream /api handler ran with an mcp_access token")
	}
}

// --- discovery path routing -------------------------------------------------

func TestDiscoveryRoutesResolveToBackend(t *testing.T) {
	router := NewRouter(oauthTestApp())
	for _, path := range []string{
		"/.well-known/oauth-authorization-server",
		"/.well-known/oauth-protected-resource",
		"/.well-known/oauth-protected-resource/mcp",
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200 JSON metadata", path, rec.Code)
		}
		if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
			t.Errorf("GET %s content-type = %q, want JSON", path, ct)
		}
	}
}

func TestDiscoveryRoutesAre404WhenOAuthDisabled(t *testing.T) {
	router := NewRouter(&App{Cfg: config.Config{
		PublicBaseURL: "https://mp.example.com",
		AuthMode:      "password",
	}})
	for _, path := range []string{
		"/.well-known/oauth-authorization-server",
		"/.well-known/oauth-protected-resource",
		"/.well-known/oauth-protected-resource/mcp",
	} {
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		if rec.Code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404 when OAuth disabled", path, rec.Code)
		}
	}
}
