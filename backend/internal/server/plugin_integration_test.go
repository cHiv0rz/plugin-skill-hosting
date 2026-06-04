package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestPluginLifecycle_Integration drives a plugin through its full CRUD arc
// against a real Postgres: create, fetch, list, update, soft-delete, confirm it
// drops out of the active set, then restore. This is the path that the
// no-DB handler unit tests cannot reach — it exercises the actual SQL
// (INSERT ... RETURNING, the deleted_at filters, the owner-version logic) and
// the soft-delete semantics end to end.
func TestPluginLifecycle_Integration(t *testing.T) {
	pool := requireTestDB(t)
	app := newIntegrationApp(t, pool)
	owner := seedUser(t, pool, "lifecycle-owner", false)

	const name = "lifecycle-plugin"

	// --- create ---
	rec := httptest.NewRecorder()
	app.handleCreatePlugin(rec, authedReq(http.MethodPost, "/api/plugins",
		`{"name":"`+name+`","description":"first","authorName":"Ann","license":"MIT"}`, owner))
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d, want 200; body=%s", rec.Code, readBody(rec))
	}
	var created Plugin
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.Name != name || created.OwnerID != owner.ID {
		t.Fatalf("created = %+v, want name=%s owner=%s", created, name, owner.ID)
	}
	// First plugin for this owner gets the 0.1.0 initial version.
	if created.Version != "0.1.0" {
		t.Errorf("version = %q, want 0.1.0", created.Version)
	}

	// --- duplicate name is a 409, not a 500 ---
	rec = httptest.NewRecorder()
	app.handleCreatePlugin(rec, authedReq(http.MethodPost, "/api/plugins",
		`{"name":"`+name+`","description":"dup"}`, owner))
	if rec.Code != http.StatusConflict {
		t.Errorf("duplicate create status = %d, want 409; body=%s", rec.Code, readBody(rec))
	}

	// --- get ---
	rec = httptest.NewRecorder()
	app.handleGetPlugin(rec, authedReq(http.MethodGet, "/api/plugins/"+name, "", owner, "name", name))
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want 200; body=%s", rec.Code, readBody(rec))
	}
	var got Plugin
	_ = json.Unmarshal(rec.Body.Bytes(), &got)
	if got.Description != "first" {
		t.Errorf("description = %q, want first", got.Description)
	}

	// --- update ---
	rec = httptest.NewRecorder()
	app.handleUpdatePlugin(rec, authedReq(http.MethodPut, "/api/plugins/"+name,
		`{"description":"second","authorName":"Ann"}`, owner, "name", name))
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want 200; body=%s", rec.Code, readBody(rec))
	}
	var updated Plugin
	_ = json.Unmarshal(rec.Body.Bytes(), &updated)
	if updated.Description != "second" {
		t.Errorf("updated description = %q, want second", updated.Description)
	}

	// --- list shows exactly the one active plugin for this owner ---
	if n := app.countOwnerPlugins(t, owner.ID); n != 1 {
		t.Errorf("active plugin count = %d, want 1", n)
	}

	// --- soft-delete ---
	rec = httptest.NewRecorder()
	app.handleDeletePlugin(rec, authedReq(http.MethodDelete, "/api/plugins/"+name, "", owner, "name", name))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want 204; body=%s", rec.Code, readBody(rec))
	}

	// active fetch now 404s; the row still exists (soft delete).
	rec = httptest.NewRecorder()
	app.handleGetPlugin(rec, authedReq(http.MethodGet, "/api/plugins/"+name, "", owner, "name", name))
	if rec.Code != http.StatusNotFound {
		t.Errorf("get after delete status = %d, want 404", rec.Code)
	}
	if n := app.countOwnerPlugins(t, owner.ID); n != 0 {
		t.Errorf("active plugin count after delete = %d, want 0", n)
	}

	// --- restore ---
	rec = httptest.NewRecorder()
	app.handleRestorePlugin(rec, authedReq(http.MethodPost, "/api/plugins/"+name+"/restore", "", owner, "name", name))
	if rec.Code != http.StatusOK {
		t.Fatalf("restore status = %d, want 200; body=%s", rec.Code, readBody(rec))
	}
	if n := app.countOwnerPlugins(t, owner.ID); n != 1 {
		t.Errorf("active plugin count after restore = %d, want 1", n)
	}
}

// TestUpdatePlugin_NotOwner_Integration confirms ownership is enforced at the
// data layer: a second user cannot edit someone else's plugin.
func TestUpdatePlugin_NotOwner_Integration(t *testing.T) {
	pool := requireTestDB(t)
	app := newIntegrationApp(t, pool)
	owner := seedUser(t, pool, "owns-it", false)
	intruder := seedUser(t, pool, "wants-it", false)

	const name = "guarded-plugin"
	rec := httptest.NewRecorder()
	app.handleCreatePlugin(rec, authedReq(http.MethodPost, "/api/plugins",
		`{"name":"`+name+`","description":"mine"}`, owner))
	if rec.Code != http.StatusOK {
		t.Fatalf("create status = %d, want 200; body=%s", rec.Code, readBody(rec))
	}

	rec = httptest.NewRecorder()
	app.handleUpdatePlugin(rec, authedReq(http.MethodPut, "/api/plugins/"+name,
		`{"description":"hijacked"}`, intruder, "name", name))
	if rec.Code != http.StatusForbidden {
		t.Errorf("intruder update status = %d, want 403; body=%s", rec.Code, readBody(rec))
	}
}

// countOwnerPlugins returns the number of non-deleted plugins owned by ownerID.
func (a *App) countOwnerPlugins(t *testing.T, ownerID string) int {
	t.Helper()
	var n int
	if err := a.DB.QueryRowContext(context.Background(),
		`SELECT COUNT(*) FROM plugins WHERE owner_id = $1 AND deleted_at IS NULL`, ownerID).Scan(&n); err != nil {
		t.Fatalf("count plugins: %v", err)
	}
	return n
}
