package server

import (
	"net/http/httptest"
	"strings"
	"testing"

	"marketplace/internal/config"
)

// Every backend response must carry the defense-in-depth security headers,
// regardless of route or status code.
func TestSecurityHeadersPresent(t *testing.T) {
	app := &App{Cfg: config.Config{AuthMode: "password", JWTSecret: "x"}}
	h := NewRouter(app)

	want := map[string]string{
		"X-Content-Type-Options": "nosniff",
		"X-Frame-Options":        "DENY",
		"Referrer-Policy":        "strict-origin-when-cross-origin",
	}

	// /healthz (200) and a 404 both flow through the middleware stack.
	for _, path := range []string{"/healthz", "/does-not-exist"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest("GET", path, nil))
		for k, v := range want {
			if got := rec.Header().Get(k); got != v {
				t.Errorf("%s: header %s = %q, want %q", path, k, got, v)
			}
		}
		csp := rec.Header().Get("Content-Security-Policy")
		if !strings.Contains(csp, "default-src 'none'") || !strings.Contains(csp, "frame-ancestors 'none'") {
			t.Errorf("%s: CSP = %q, want default-src/frame-ancestors none", path, csp)
		}
	}
}
