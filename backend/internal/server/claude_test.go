package server

import (
	"context"
	"strings"
	"testing"

	"marketplace/internal/config"
)

func TestCallClaude_NoAPIKeyConfigured(t *testing.T) {
	a := &App{Cfg: config.Config{AnthropicAPIKey: ""}}
	if _, err := a.callClaude(context.Background(), "sys", "user", 1024); err == nil {
		t.Error("expected error when API key is unset")
	} else if !strings.Contains(err.Error(), "ANTHROPIC_API_KEY") {
		t.Errorf("error should mention env var; got %v", err)
	}
}

func TestCallClaude_WhitespaceKeyTreatedAsUnset(t *testing.T) {
	a := &App{Cfg: config.Config{AnthropicAPIKey: "   "}}
	if _, err := a.callClaude(context.Background(), "sys", "user", 1024); err == nil {
		t.Error("expected error when API key is only whitespace")
	}
}
