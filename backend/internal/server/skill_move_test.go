package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// createPluginForMove makes an active plugin owned by user and returns it.
func createPluginForMove(t *testing.T, app *App, owner *User, name string) {
	t.Helper()
	rec := httptest.NewRecorder()
	app.handleCreatePlugin(rec, authedReq(http.MethodPost, "/api/plugins",
		`{"name":"`+name+`","description":"d","authorName":"A","license":"MIT"}`, owner))
	if rec.Code != http.StatusOK {
		t.Fatalf("create plugin %s: status %d; body=%s", name, rec.Code, readBody(rec))
	}
}

// createSkillForMove adds a skill to a plugin and returns its decoded form.
func createSkillForMove(t *testing.T, app *App, owner *User, plugin, skill string) Skill {
	t.Helper()
	rec := httptest.NewRecorder()
	app.handleCreateSkill(rec, authedReq(http.MethodPost, "/api/plugins/"+plugin+"/skills",
		`{"name":"`+skill+`","description":"does a thing","body":"## Instructions\nstep."}`,
		owner, "name", plugin))
	if rec.Code != http.StatusOK {
		t.Fatalf("create skill %s: status %d; body=%s", skill, rec.Code, readBody(rec))
	}
	var s Skill
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil {
		t.Fatalf("decode skill: %v", err)
	}
	return s
}

// TestMoveSkill_Integration moves a skill between two plugins and verifies the
// skill (keeping its id) lands in the target, leaves the source, and that the
// guard rails — unknown target, same plugin, name collision — respond correctly.
func TestMoveSkill_Integration(t *testing.T) {
	pool := requireTestDB(t)
	app := newIntegrationApp(t, pool)
	owner := seedUser(t, pool, "move-owner", false)

	createPluginForMove(t, app, owner, "move-src")
	createPluginForMove(t, app, owner, "move-dst")
	original := createSkillForMove(t, app, owner, "move-src", "widget")

	move := func(src, skill, target string) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		app.handleMoveSkill(rec, authedReq(http.MethodPost,
			"/api/plugins/"+src+"/skills/"+skill+"/move",
			`{"targetPlugin":"`+target+`"}`, owner, "name", src, "skill", skill))
		return rec
	}

	// --- moving onto the same plugin is a 400 ---
	if rec := move("move-src", "widget", "move-src"); rec.Code != http.StatusBadRequest {
		t.Errorf("same-plugin move status = %d, want 400; body=%s", rec.Code, readBody(rec))
	}

	// --- unknown target is a 404 ---
	if rec := move("move-src", "widget", "no-such-plugin"); rec.Code != http.StatusNotFound {
		t.Errorf("unknown-target move status = %d, want 404; body=%s", rec.Code, readBody(rec))
	}

	// --- happy path ---
	rec := move("move-src", "widget", "move-dst")
	if rec.Code != http.StatusOK {
		t.Fatalf("move status = %d, want 200; body=%s", rec.Code, readBody(rec))
	}
	var moved Skill
	if err := json.Unmarshal(rec.Body.Bytes(), &moved); err != nil {
		t.Fatalf("decode moved: %v", err)
	}
	if moved.ID != original.ID {
		t.Errorf("moved id = %s, want unchanged %s", moved.ID, original.ID)
	}

	// --- target now lists the skill, source no longer does ---
	if !pluginHasActiveSkill(t, app, "move-dst", "widget") {
		t.Errorf("target plugin missing moved skill")
	}
	if pluginHasActiveSkill(t, app, "move-src", "widget") {
		t.Errorf("source plugin still lists moved skill")
	}

	// --- a name collision in the target is a 409 ---
	createSkillForMove(t, app, owner, "move-src", "widget") // recreate in source
	rec = move("move-src", "widget", "move-dst")
	if rec.Code != http.StatusConflict {
		t.Errorf("colliding move status = %d, want 409; body=%s", rec.Code, readBody(rec))
	}
}

// pluginHasActiveSkill reports whether the named plugin currently lists an
// active skill with the given name, via handleGetPlugin.
func pluginHasActiveSkill(t *testing.T, app *App, plugin, skill string) bool {
	t.Helper()
	rec := httptest.NewRecorder()
	app.handleGetPlugin(rec, authedReq(http.MethodGet, "/api/plugins/"+plugin, "", nil, "name", plugin))
	if rec.Code != http.StatusOK {
		t.Fatalf("get plugin %s: status %d; body=%s", plugin, rec.Code, readBody(rec))
	}
	var p Plugin
	if err := json.Unmarshal(rec.Body.Bytes(), &p); err != nil {
		t.Fatalf("decode plugin: %v", err)
	}
	for _, s := range p.Skills {
		if s.Name == skill {
			return true
		}
	}
	return false
}
