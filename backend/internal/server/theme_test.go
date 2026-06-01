package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIsValidTheme(t *testing.T) {
	for _, ok := range []string{"light", "dark", "midnight", "sepia", "contrast"} {
		if !isValidTheme(ok) {
			t.Errorf("isValidTheme(%q) = false, want true", ok)
		}
	}
	for _, bad := range []string{"", "Light", "neon", "light ", "dark;drop"} {
		if isValidTheme(bad) {
			t.Errorf("isValidTheme(%q) = true, want false", bad)
		}
	}
}

func TestDefaultThemeIsValid(t *testing.T) {
	if !isValidTheme(DefaultTheme) {
		t.Errorf("DefaultTheme %q is not in the allowed set", DefaultTheme)
	}
}

func TestHandleMe_IncludesTheme(t *testing.T) {
	a := &App{}
	user := &User{ID: "u1", Email: "a@b.com", Username: "alice", Theme: "dark"}
	r := httptest.NewRequest("GET", "/api/me", nil)
	r = r.WithContext(context.WithValue(r.Context(), ctxUserKey, user))
	rec := httptest.NewRecorder()
	a.handleMe(rec, r)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"theme":"dark"`) {
		t.Errorf("body = %q, want theme=dark in payload", rec.Body.String())
	}
}

func TestHandleSetTheme_RejectsInvalidJSON(t *testing.T) {
	a := &App{}
	r := httptest.NewRequest("PUT", "/api/me/theme", strings.NewReader("{not json"))
	rec := httptest.NewRecorder()
	a.handleSetTheme(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestHandleSetTheme_RejectsUnknownTheme(t *testing.T) {
	// Validation happens before any DB access, so a nil-DB App is fine here:
	// the handler must 400 without ever reaching the UPDATE.
	a := &App{}
	r := httptest.NewRequest("PUT", "/api/me/theme", strings.NewReader(`{"theme":"neon"}`))
	rec := httptest.NewRecorder()
	a.handleSetTheme(rec, r)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "unknown theme") {
		t.Errorf("body = %q, want 'unknown theme'", rec.Body.String())
	}
}
