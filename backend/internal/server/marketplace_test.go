package server

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarketplaceDoc_EmbedsSchemaAndRepository(t *testing.T) {
	doc := marketplaceDoc{
		Schema: MarketplaceSchemaURL,
		Name:   "m",
		Owner:  marketplaceOwner{Name: "m"},
		Plugins: []marketplacePlugin{{
			Name:        "p",
			Description: "d",
			Repository:  "https://example.com/git/p.git",
			Source:      marketplaceSource{Source: "url", URL: "https://example.com/git/p.git"},
		}},
	}
	out, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	body := string(out)
	if !strings.Contains(body, `"$schema":"https://json.schemastore.org/claude-code-marketplace.json"`) {
		t.Errorf("missing $schema in marketplace JSON: %s", body)
	}
	if !strings.Contains(body, `"repository":"https://example.com/git/p.git"`) {
		t.Errorf("missing per-plugin repository in marketplace JSON: %s", body)
	}
}

func TestEmbedTokenInBase(t *testing.T) {
	cases := []struct {
		base, token, want string
	}{
		{"https://example.com", "tok123", "https://_:tok123@example.com"},
		{"https://example.com/", "tok123", "https://_:tok123@example.com"},
		{"https://example.com/path/", "tok123", "https://_:tok123@example.com/path"},
		{"https://example.com", "", "https://example.com"},              // empty token => unchanged
		{"::not a url", "tok", "::not a url"},                           // unparseable => unchanged
		{"http://localhost:8080", "abc", "http://_:abc@localhost:8080"}, // port preserved
		{"https://example.com", "to:k", "https://_:to%3Ak@example.com"}, // colon in token escaped
		{"https://example.com/p", "tok", "https://_:tok@example.com/p"}, // path without trailing slash
	}
	for _, c := range cases {
		got := embedTokenInBase(c.base, c.token)
		if got != c.want {
			t.Errorf("embedTokenInBase(%q, %q) = %q, want %q", c.base, c.token, got, c.want)
		}
	}
}
