package server

import (
	"os"
	"path/filepath"
	"testing"
)

// writeTree materialises a map of slash-relative path -> contents under root,
// creating parent dirs as needed.
func writeTree(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for rel, content := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatalf("mkdir for %s: %v", rel, err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
}

func TestDirsEqual(t *testing.T) {
	base := map[string]string{
		".claude-plugin/plugin.json": `{"name":"x"}`,
		"skills/foo/SKILL.md":        "# foo\n",
		"README.md":                  "readme\n",
	}

	cases := []struct {
		name   string
		mutate func(b map[string]string)
		want   bool
	}{
		{"identical", func(map[string]string) {}, true},
		{"changed content", func(b map[string]string) { b["README.md"] = "different\n" }, false},
		{"extra file", func(b map[string]string) { b["skills/foo/extra.txt"] = "x" }, false},
		{"missing file", func(b map[string]string) { delete(b, "README.md") }, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			a := t.TempDir()
			b := t.TempDir()
			writeTree(t, a, base)

			bf := make(map[string]string, len(base))
			for k, v := range base {
				bf[k] = v
			}
			tc.mutate(bf)
			writeTree(t, b, bf)

			got, err := dirsEqual(a, b)
			if err != nil {
				t.Fatalf("dirsEqual: %v", err)
			}
			if got != tc.want {
				t.Errorf("dirsEqual = %v, want %v", got, tc.want)
			}
		})
	}
}

// TestHashTree_MissingRootIsEmpty confirms a non-existent directory hashes to an
// empty map (not an error), so a plugin absent from the mirror compares as
// different from any non-empty render rather than blowing up.
func TestHashTree_MissingRootIsEmpty(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	m, err := hashTree(missing)
	if err != nil {
		t.Fatalf("hashTree(missing): %v", err)
	}
	if len(m) != 0 {
		t.Errorf("hashTree(missing) = %v, want empty", m)
	}
}
