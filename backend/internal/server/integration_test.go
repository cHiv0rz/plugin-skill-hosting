package server

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"marketplace/internal/config"
	"marketplace/internal/db"
)

// requireTestDB opens the disposable Postgres named by TEST_DATABASE_URL,
// applies all migrations, and returns a ready *sql.DB. The whole DB-backed
// integration layer is gated on this variable: with it unset (e.g. a plain
// `go test ./...` on a laptop) every integration test skips, so the default
// suite stays hermetic. CI sets it to a throwaway service container, so these
// run on every backend change.
func requireTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set TEST_DATABASE_URL to run DB integration tests")
	}
	pool, err := db.Open(dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.Migrate(pool); err != nil {
		pool.Close()
		t.Fatalf("migrate: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

// newIntegrationApp wires an App around a real DB in password-auth mode with a
// 32-byte API-token key (some handlers derive display tokens from it). DataDir
// points at a per-test temp dir so the git materialize step on plugin mutations
// writes its bare repos and work trees there and leaves the checkout clean.
func newIntegrationApp(t *testing.T, pool *sql.DB) *App {
	t.Helper()
	key := make([]byte, 32)
	return &App{
		DB: pool,
		Cfg: config.Config{
			APITokenKey:     key,
			AuthMode:        "password",
			MarketplaceName: "integration-test",
			DefaultLicense:  "MIT",
			DataDir:         t.TempDir(),
		},
	}
}

// seedUser inserts an approved user and returns it. Rows are namespaced by the
// caller-supplied username and torn down via t.Cleanup, so concurrent packages
// and reruns don't collide. password_hash is the only NOT NULL column without a
// default, so a placeholder is enough — these tests never exercise the password
// flow itself.
func seedUser(t *testing.T, pool *sql.DB, username string, admin bool) *User {
	t.Helper()
	ctx := context.Background()
	email := username + "@integration.test"
	u := &User{}
	err := pool.QueryRowContext(ctx, `
		INSERT INTO users (email, username, password_hash, status, is_admin)
		VALUES ($1, $2, 'x', 'approved', $3)
		RETURNING id, email, username, status, is_admin, token_version
	`, email, username, admin).Scan(&u.ID, &u.Email, &u.Username, &u.Status, &u.IsAdmin, &u.TokenVersion)
	if err != nil {
		t.Fatalf("seed user %q: %v", username, err)
	}
	t.Cleanup(func() {
		// plugins/skills cascade off the owner row.
		_, _ = pool.ExecContext(context.Background(), `DELETE FROM users WHERE id = $1`, u.ID)
	})
	return u
}

// authedReq builds a request carrying an authenticated user and chi URL params,
// matching what the router's middleware + mux would have populated. urlParams
// are flattened key/value pairs: authedReq(..., "name", "my-plugin").
func authedReq(method, target, body string, user *User, urlParams ...string) *http.Request {
	var r *http.Request
	if body == "" {
		r = httptest.NewRequest(method, target, nil)
	} else {
		r = httptest.NewRequest(method, target, strings.NewReader(body))
	}
	ctx := context.WithValue(r.Context(), ctxUserKey, user)
	if len(urlParams) > 0 {
		rctx := chi.NewRouteContext()
		for i := 0; i+1 < len(urlParams); i += 2 {
			rctx.URLParams.Add(urlParams[i], urlParams[i+1])
		}
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	}
	return r.WithContext(ctx)
}

// readBody returns the recorder's body as a string for assertion messages.
func readBody(rec *httptest.ResponseRecorder) string {
	b, _ := io.ReadAll(rec.Result().Body)
	return string(b)
}
