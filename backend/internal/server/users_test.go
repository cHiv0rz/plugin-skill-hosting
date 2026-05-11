package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"marketplace/internal/config"
)

func TestUserSummary_OmitsSecrets(t *testing.T) {
	// Belt-and-braces: confirm the public DTO has no password/api_token/oidc
	// fields. If anyone adds one by mistake, this test fails loudly.
	u := UserSummary{ID: "id", Username: "alice", Email: "a@b.com"}
	buf, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(buf)
	for _, forbidden := range []string{"password", "api_token", "apiToken", "oidc"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Errorf("UserSummary JSON %q contains forbidden field %q", body, forbidden)
		}
	}
}

func TestListUsers_RequiresAuth(t *testing.T) {
	// Build the real router so we exercise the route-table wiring, not just
	// the handler in isolation. With no Authorization header the request must
	// be rejected by authMiddleware before the (nil) DB is touched.
	app := &App{Cfg: config.Config{
		AuthMode:  "password",
		JWTSecret: "x",
	}}
	h := NewRouter(app)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest("GET", "/api/users", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
