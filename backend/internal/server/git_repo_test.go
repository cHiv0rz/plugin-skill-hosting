package server

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"marketplace/internal/config"
)

func TestBuildSkillMarkdown_WithBody(t *testing.T) {
	s := Skill{Name: "tester", Description: "does things", Body: "## Tester\n\nbody here"}
	out := buildSkillMarkdown(s)
	if !strings.HasPrefix(out, "---\n") {
		t.Error("missing frontmatter opener")
	}
	if !strings.Contains(out, "name: tester\n") {
		t.Error("missing name line")
	}
	if !strings.Contains(out, "description: does things\n") {
		t.Error("missing description line")
	}
	if !strings.HasSuffix(out, "body here\n") {
		t.Error("expected body to end with newline")
	}
}

func TestBuildSkillMarkdown_DescriptionNewlinesFlattened(t *testing.T) {
	s := Skill{Name: "x", Description: "line1\nline2", Body: "b"}
	out := buildSkillMarkdown(s)
	if !strings.Contains(out, "description: line1 line2\n") {
		t.Errorf("description newlines not flattened to spaces; got: %q", out)
	}
}

func TestBuildSkillMarkdown_EmptyBodyDefaults(t *testing.T) {
	s := Skill{Name: "x", Description: "d", Body: ""}
	out := buildSkillMarkdown(s)
	if !strings.Contains(out, "## x\n\nd\n") {
		t.Errorf("empty body should fall back to default heading; got: %q", out)
	}
}

func TestBuildSkillMarkdown_BodyWithoutTrailingNewline(t *testing.T) {
	s := Skill{Name: "n", Description: "d", Body: "no newline"}
	out := buildSkillMarkdown(s)
	if !strings.HasSuffix(out, "no newline\n") {
		t.Errorf("expected trailing newline appended; got: %q", out)
	}
}

func TestBuildSkillMarkdown_EmitsExtraFrontmatter(t *testing.T) {
	s := Skill{
		Name:             "ext",
		Description:      "d",
		ExtraFrontmatter: "allowed-tools:\n  - Read\n  - Edit\nlicense: MIT",
		Body:             "body",
	}
	out := buildSkillMarkdown(s)
	want := "---\nname: ext\ndescription: d\nallowed-tools:\n  - Read\n  - Edit\nlicense: MIT\n---\n\nbody\n"
	if out != want {
		t.Errorf("buildSkillMarkdown =\n%q\nwant=\n%q", out, want)
	}
}

func TestBuildSkillMarkdown_RoundTripWithExtras(t *testing.T) {
	original := Skill{
		Name:             "rt",
		Description:      "round trip",
		ExtraFrontmatter: "allowed-tools:\n  - Read",
		Body:             "the body\n",
	}
	out := buildSkillMarkdown(original)
	name, desc, extra, body, err := parseSkillFrontmatter([]byte(out))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if name != original.Name || desc != original.Description || extra != original.ExtraFrontmatter || body != original.Body {
		t.Errorf("round-trip mismatch: name=%q desc=%q extra=%q body=%q", name, desc, extra, body)
	}
}

func TestRepoAndWorkPaths(t *testing.T) {
	a := &App{Cfg: config.Config{DataDir: "/var/data"}}
	if got := a.repoPath("foo"); got != "/var/data/repos/foo.git" {
		t.Errorf("repoPath = %q", got)
	}
	if got := a.workPath("foo"); got != "/var/data/work/foo" {
		t.Errorf("workPath = %q", got)
	}
}

func TestPluginRepoURL(t *testing.T) {
	cases := []struct {
		base, name, want string
	}{
		{"https://example.com", "foo", "https://example.com/git/foo.git"},
		{"https://example.com/", "foo", "https://example.com/git/foo.git"}, // trailing slash trimmed
		{"https://example.com/p/", "bar", "https://example.com/p/git/bar.git"},
		{"", "foo", ""}, // no base => empty (avoids emitting just "/git/foo.git")
	}
	for _, c := range cases {
		a := &App{Cfg: config.Config{PublicBaseURL: c.base}}
		if got := a.pluginRepoURL(c.name); got != c.want {
			t.Errorf("pluginRepoURL(base=%q, name=%q) = %q, want %q", c.base, c.name, got, c.want)
		}
	}
}

func TestPluginManifest_EmbedsSchemaAndRepository(t *testing.T) {
	m := pluginManifest{
		Schema:     PluginManifestSchemaURL,
		Name:       "foo",
		Repository: "https://example.com/git/foo.git",
	}
	out, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(out)
	if !strings.Contains(body, `"$schema":"https://json.schemastore.org/claude-code-plugin-manifest.json"`) {
		t.Errorf("missing $schema in plugin manifest: %s", body)
	}
	if !strings.Contains(body, `"repository":"https://example.com/git/foo.git"`) {
		t.Errorf("missing repository in plugin manifest: %s", body)
	}
}

func TestWipeWorkTree_PreservesGitDir(t *testing.T) {
	dir := t.TempDir()
	mustMkdir(t, filepath.Join(dir, ".git"))
	mustWrite(t, filepath.Join(dir, ".git", "HEAD"), "ref: refs/heads/main\n")
	mustWrite(t, filepath.Join(dir, "README.md"), "hi")
	mustMkdir(t, filepath.Join(dir, "skills"))
	mustWrite(t, filepath.Join(dir, "skills", "a.txt"), "x")

	if err := wipeWorkTree(dir); err != nil {
		t.Fatalf("wipeWorkTree: %v", err)
	}

	if _, err := os.Stat(filepath.Join(dir, ".git", "HEAD")); err != nil {
		t.Errorf(".git contents removed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "README.md")); !os.IsNotExist(err) {
		t.Errorf("README.md should be gone, err = %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "skills")); !os.IsNotExist(err) {
		t.Errorf("skills/ should be gone, err = %v", err)
	}
}

func TestWipeWorkTree_MissingDir(t *testing.T) {
	if err := wipeWorkTree(filepath.Join(t.TempDir(), "does-not-exist")); err == nil {
		t.Error("expected error for missing directory")
	}
}

func TestRemoveInternalRepo_RemovesBareAndWorkTree(t *testing.T) {
	dir := t.TempDir()
	a := &App{Cfg: config.Config{DataDir: dir}}
	mustMkdir(t, a.repoPath("gone"))
	mustWrite(t, filepath.Join(a.repoPath("gone"), "HEAD"), "ref: refs/heads/main\n")
	mustMkdir(t, filepath.Join(a.workPath("gone"), ".git"))
	mustWrite(t, filepath.Join(a.workPath("gone"), "README.md"), "stale")

	if err := a.removeInternalRepo("gone"); err != nil {
		t.Fatalf("removeInternalRepo: %v", err)
	}
	if _, err := os.Stat(a.repoPath("gone")); !os.IsNotExist(err) {
		t.Errorf("bare repo should be removed, err=%v", err)
	}
	if _, err := os.Stat(a.workPath("gone")); !os.IsNotExist(err) {
		t.Errorf("worktree should be removed, err=%v", err)
	}
}

func TestWriteSkillFileToWorkTree_Text(t *testing.T) {
	dir := t.TempDir()
	f := SkillFile{Path: "scripts/run.sh", Content: "#!/bin/sh\necho hi\n"}
	if err := writeSkillFileToWorkTree(dir, f); err != nil {
		t.Fatalf("writeSkillFileToWorkTree: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "scripts", "run.sh"))
	if err != nil {
		t.Fatalf("read written file: %v", err)
	}
	if string(got) != f.Content {
		t.Errorf("content = %q, want %q", string(got), f.Content)
	}
}

func TestWriteSkillFileToWorkTree_Binary(t *testing.T) {
	dir := t.TempDir()
	raw := []byte{0x00, 0x01, 0xff, 0xfe}
	f := SkillFile{
		Path:     "assets/tiny.bin",
		IsBinary: true,
		Content:  base64.StdEncoding.EncodeToString(raw),
	}
	if err := writeSkillFileToWorkTree(dir, f); err != nil {
		t.Fatalf("writeSkillFileToWorkTree: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "assets", "tiny.bin"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(raw) {
		t.Errorf("binary mismatch")
	}
}

func TestWriteSkillFileToWorkTree_BadBase64(t *testing.T) {
	dir := t.TempDir()
	f := SkillFile{Path: "assets/x.bin", IsBinary: true, Content: "not_base64!!!"}
	if err := writeSkillFileToWorkTree(dir, f); err == nil {
		t.Error("expected error for invalid base64 content")
	}
}

func mustMkdir(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
