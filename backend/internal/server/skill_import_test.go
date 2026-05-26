package server

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
)

func TestParseSkillFrontmatter_Basic(t *testing.T) {
	in := []byte("---\nname: my-skill\ndescription: does things\n---\n\nbody text\n")
	name, desc, extra, body, err := parseSkillFrontmatter(in)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if name != "my-skill" || desc != "does things" || body != "body text\n" {
		t.Errorf("got name=%q desc=%q body=%q", name, desc, body)
	}
	if extra != "" {
		t.Errorf("expected no extras, got %q", extra)
	}
}

func TestParseSkillFrontmatter_LeadingBlankLines(t *testing.T) {
	in := []byte("\n\n---\nname: x\ndescription: y\n---\nthe body\n")
	if _, _, _, body, err := parseSkillFrontmatter(in); err != nil || body != "the body\n" {
		t.Errorf("got body=%q err=%v", body, err)
	}
}

func TestParseSkillFrontmatter_QuotedValues(t *testing.T) {
	in := []byte("---\nname: \"quoted-name\"\ndescription: 'quoted desc'\n---\nbody\n")
	name, desc, _, _, err := parseSkillFrontmatter(in)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if name != "quoted-name" || desc != "quoted desc" {
		t.Errorf("got name=%q desc=%q", name, desc)
	}
}

func TestParseSkillFrontmatter_MissingDelimiter(t *testing.T) {
	in := []byte("no frontmatter at all\n")
	if _, _, _, _, err := parseSkillFrontmatter(in); err == nil {
		t.Error("expected error for missing frontmatter")
	}
}

func TestParseSkillFrontmatter_UnterminatedFrontmatter(t *testing.T) {
	in := []byte("---\nname: a\ndescription: b\n")
	if _, _, _, _, err := parseSkillFrontmatter(in); err == nil {
		t.Error("expected error for unterminated frontmatter")
	}
}

func TestParseSkillFrontmatter_MissingName(t *testing.T) {
	in := []byte("---\ndescription: only a desc\n---\nbody\n")
	if _, _, _, _, err := parseSkillFrontmatter(in); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestParseSkillFrontmatter_FoldedDescription(t *testing.T) {
	in := []byte("---\nname: plan-driven-dev\ndescription: >\n  Drive iterative, test-first implementation from a project_plan.md file.\n  Use this skill whenever the user asks.\n---\n\nbody\n")
	_, desc, _, body, err := parseSkillFrontmatter(in)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := "Drive iterative, test-first implementation from a project_plan.md file. Use this skill whenever the user asks."
	if desc != want {
		t.Errorf("desc = %q\nwant = %q", desc, want)
	}
	if body != "body\n" {
		t.Errorf("body = %q", body)
	}
}

func TestParseSkillFrontmatter_FoldedDescriptionWithParagraphBreak(t *testing.T) {
	in := []byte("---\nname: x\ndescription: >\n  First paragraph here.\n\n  Second paragraph here.\n---\nbody\n")
	_, desc, _, _, err := parseSkillFrontmatter(in)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := "First paragraph here.\nSecond paragraph here."
	if desc != want {
		t.Errorf("desc = %q\nwant = %q", desc, want)
	}
}

func TestParseSkillFrontmatter_LiteralDescription(t *testing.T) {
	in := []byte("---\nname: x\ndescription: |\n  Line one\n  Line two\n---\nbody\n")
	_, desc, _, _, err := parseSkillFrontmatter(in)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	want := "Line one\nLine two"
	if desc != want {
		t.Errorf("desc = %q\nwant = %q", desc, want)
	}
}

func TestParseSkillFrontmatter_FoldedKeyStopsAtNextKey(t *testing.T) {
	in := []byte("---\ndescription: >\n  multi-line\n  description here\nname: my-skill\n---\nbody\n")
	name, desc, _, _, err := parseSkillFrontmatter(in)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if name != "my-skill" {
		t.Errorf("name = %q", name)
	}
	if desc != "multi-line description here" {
		t.Errorf("desc = %q", desc)
	}
}

func TestParseSkillFrontmatter_MissingDescription(t *testing.T) {
	in := []byte("---\nname: only-a-name\n---\nbody\n")
	if _, _, _, _, err := parseSkillFrontmatter(in); err == nil {
		t.Error("expected error for missing description")
	}
}

func TestParseSkillFrontmatter_PreservesExtras(t *testing.T) {
	in := []byte("---\nname: ext\ndescription: d\nallowed-tools:\n  - Read\n  - Edit\nlicense: MIT\n---\nbody\n")
	name, desc, extra, body, err := parseSkillFrontmatter(in)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if name != "ext" || desc != "d" || body != "body\n" {
		t.Errorf("name/desc/body wrong: name=%q desc=%q body=%q", name, desc, body)
	}
	want := "allowed-tools:\n  - Read\n  - Edit\nlicense: MIT"
	if extra != want {
		t.Errorf("extra mismatch:\n got: %q\nwant: %q", extra, want)
	}
}

func TestParseSkillFrontmatter_ExtrasWithFoldedDescription(t *testing.T) {
	// description uses folded scalar; its indented continuations must NOT
	// leak into extras.
	in := []byte("---\nname: x\ndescription: >\n  multi line desc\n  continues here\nallowed-tools:\n  - Read\n---\nbody\n")
	_, desc, extra, _, err := parseSkillFrontmatter(in)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if desc != "multi line desc continues here" {
		t.Errorf("desc = %q", desc)
	}
	want := "allowed-tools:\n  - Read"
	if extra != want {
		t.Errorf("extra = %q, want %q", extra, want)
	}
}

func TestShouldSkipZipEntry(t *testing.T) {
	skips := []string{
		"__MACOSX/foo",
		"__MACOSX",
		"scripts/.DS_Store",
		"references/._notes.md",
	}
	for _, p := range skips {
		if !shouldSkipZipEntry(p) {
			t.Errorf("expected %q to be skipped", p)
		}
	}
	for _, p := range []string{"SKILL.md", "scripts/run.sh", "references/notes.md"} {
		if shouldSkipZipEntry(p) {
			t.Errorf("did not expect %q to be skipped", p)
		}
	}
}

// buildZip helps tests by writing a small in-memory zip from a name → content map.
func buildZip(t *testing.T, files map[string]string) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatalf("zip create %s: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("zip write %s: %v", name, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("zip close: %v", err)
	}
	return buf.Bytes()
}

func TestExtractSkillZip_RootSkill(t *testing.T) {
	data := buildZip(t, map[string]string{
		"SKILL.md":           "---\nname: hello\ndescription: greet\n---\n\nbody here\n",
		"scripts/run.sh":     "#!/bin/sh\necho hi\n",
		"references/note.md": "see also\n",
	})
	parsed, err := extractSkillZip(data)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if parsed.Name != "hello" || parsed.Description != "greet" || !strings.HasPrefix(parsed.Body, "body here") {
		t.Errorf("parsed = %+v", parsed)
	}
	if len(parsed.Files) != 2 {
		t.Fatalf("got %d files, want 2", len(parsed.Files))
	}
	paths := map[string]bool{}
	for _, f := range parsed.Files {
		paths[f.Path] = true
	}
	if !paths["scripts/run.sh"] || !paths["references/note.md"] {
		t.Errorf("missing expected files: %v", paths)
	}
}

func TestExtractSkillZip_StripsSingleTopLevelDir(t *testing.T) {
	data := buildZip(t, map[string]string{
		"my-skill/SKILL.md":        "---\nname: nested\ndescription: d\n---\nbody\n",
		"my-skill/scripts/run.sh":  "echo\n",
		"my-skill/assets/logo.txt": "art\n",
	})
	parsed, err := extractSkillZip(data)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if parsed.Name != "nested" {
		t.Errorf("name = %q", parsed.Name)
	}
	paths := map[string]bool{}
	for _, f := range parsed.Files {
		paths[f.Path] = true
	}
	if !paths["scripts/run.sh"] || !paths["assets/logo.txt"] {
		t.Errorf("top-level prefix not stripped: %v", paths)
	}
}

func TestExtractSkillZip_RejectsMissingSkillMd(t *testing.T) {
	data := buildZip(t, map[string]string{
		"scripts/run.sh": "echo\n",
	})
	if _, err := extractSkillZip(data); err == nil {
		t.Error("expected error when SKILL.md is missing")
	}
}

func TestExtractSkillZip_AcceptsArbitraryFolders(t *testing.T) {
	// Folder names outside the conventional scripts/references/assets trio are
	// accepted — the marketplace stores whatever folder the user chooses via
	// the API or import. evals/ is still dropped explicitly (see
	// unsupportedSkillRoots) but other names flow through.
	data := buildZip(t, map[string]string{
		"SKILL.md":      "---\nname: x\ndescription: y\n---\nbody\n",
		"random/x.txt":  "ok\n",
		"docs/notes.md": "ok\n",
	})
	parsed, err := extractSkillZip(data)
	if err != nil {
		t.Fatalf("extractSkillZip: %v", err)
	}
	paths := map[string]bool{}
	for _, f := range parsed.Files {
		paths[f.Path] = true
	}
	if !paths["random/x.txt"] || !paths["docs/notes.md"] {
		t.Errorf("expected arbitrary folders to be kept, got %+v", parsed.Files)
	}
}

func TestExtractSkillZip_RejectsBadChars(t *testing.T) {
	// Folder names still have to pass the segment regex: no spaces or exotic
	// characters that would round-trip badly through git or filesystem layers.
	data := buildZip(t, map[string]string{
		"SKILL.md":         "---\nname: x\ndescription: y\n---\nbody\n",
		"bad folder/x.txt": "nope\n",
	})
	if _, err := extractSkillZip(data); err == nil {
		t.Error("expected error when folder name contains invalid chars")
	}
}

func TestExtractSkillZip_RejectsTraversal(t *testing.T) {
	data := buildZip(t, map[string]string{
		"SKILL.md":      "---\nname: x\ndescription: y\n---\nbody\n",
		"../etc/passwd": "nope\n",
	})
	if _, err := extractSkillZip(data); err == nil {
		t.Error("expected error for path traversal")
	}
}

func TestExtractSkillZip_SkipsMacOSXJunk(t *testing.T) {
	data := buildZip(t, map[string]string{
		"SKILL.md":          "---\nname: a\ndescription: b\n---\nbody\n",
		"__MACOSX/SKILL.md": "junk",
		"scripts/.DS_Store": "junk",
		"scripts/run.sh":    "echo\n",
	})
	parsed, err := extractSkillZip(data)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(parsed.Files) != 1 || parsed.Files[0].Path != "scripts/run.sh" {
		t.Errorf("expected only scripts/run.sh, got %+v", parsed.Files)
	}
}

func TestExtractSkillZip_SilentlySkipsEvals(t *testing.T) {
	data := buildZip(t, map[string]string{
		"SKILL.md":       "---\nname: e\ndescription: d\n---\nbody\n",
		"scripts/run.sh": "echo\n",
		"evals/case.md":  "test case\n",
		"evals/sub/x.md": "nested\n",
	})
	parsed, err := extractSkillZip(data)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(parsed.Files) != 1 || parsed.Files[0].Path != "scripts/run.sh" {
		t.Errorf("evals/ should be silently dropped, got %+v", parsed.Files)
	}
}

func TestExtractSkillZip_SilentlySkipsEvalsUnderTopLevelDir(t *testing.T) {
	data := buildZip(t, map[string]string{
		"my-skill/SKILL.md":      "---\nname: n\ndescription: d\n---\nbody\n",
		"my-skill/scripts/r.sh":  "echo\n",
		"my-skill/evals/case.md": "test\n",
	})
	parsed, err := extractSkillZip(data)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(parsed.Files) != 1 || parsed.Files[0].Path != "scripts/r.sh" {
		t.Errorf("nested evals/ should be silently dropped, got %+v", parsed.Files)
	}
}

func TestExtractSkillZip_DetectsBinary(t *testing.T) {
	data := buildZip(t, map[string]string{
		"SKILL.md":        "---\nname: x\ndescription: y\n---\nbody\n",
		"assets/blob.bin": "\x00\x01\x02\xff\xfe",
	})
	parsed, err := extractSkillZip(data)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	if len(parsed.Files) != 1 || !parsed.Files[0].IsBinary {
		t.Errorf("expected blob marked binary, got %+v", parsed.Files)
	}
}

func TestExtractSkillZip_InvalidZip(t *testing.T) {
	if _, err := extractSkillZip([]byte("not a zip")); err == nil {
		t.Error("expected error for malformed zip")
	}
}
