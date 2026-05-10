package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"marketplace/internal/metrics"
	"marketplace/internal/skillvalidation"
)

const (
	claudeAPIURL     = "https://api.anthropic.com/v1/messages"
	claudeAPIVersion = "2023-06-01"
)

type claudeMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type claudeRequest struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []claudeMessage `json:"messages"`
}

type claudeResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// callClaude sends a single user turn to the Claude messages API and returns
// the model's text reply. Some newer models reject assistant-prefill, so we
// rely on prompt engineering + a tolerant JSON extractor on the caller side
// instead of pinning the response with a leading `{`.
func (a *App) callClaude(ctx context.Context, system, user string) (string, error) {
	if strings.TrimSpace(a.Cfg.AnthropicAPIKey) == "" {
		return "", errors.New("Claude API not configured (set ANTHROPIC_API_KEY)")
	}
	payload, err := json.Marshal(claudeRequest{
		Model:     a.Cfg.AnthropicModel,
		MaxTokens: 2048,
		System:    system,
		Messages:  []claudeMessage{{Role: "user", Content: user}},
	})
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", claudeAPIURL, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("anthropic-version", claudeAPIVersion)
	req.Header.Set("x-api-key", a.Cfg.AnthropicAPIKey)

	client := &http.Client{Timeout: 90 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var cr claudeResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return "", fmt.Errorf("decode claude response: %w", err)
	}
	if cr.Error != nil {
		return "", fmt.Errorf("claude api: %s", cr.Error.Message)
	}
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("claude api status %d", resp.StatusCode)
	}
	var sb strings.Builder
	for _, c := range cr.Content {
		if c.Type == "text" {
			sb.WriteString(c.Text)
		}
	}
	return sb.String(), nil
}

type validateSkillRequest struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Body        string             `json:"body"`
	Files       []SkillFileSummary `json:"files,omitempty"`
}

func (a *App) handleValidateSkill(w http.ResponseWriter, r *http.Request) {
	var req validateSkillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(req.Description) == "" && strings.TrimSpace(req.Body) == "" {
		writeErr(w, http.StatusBadRequest, "description or body is required")
		return
	}

	userMsg := fmt.Sprintf(
		"Skill name: %s\n\n--- Description ---\n%s\n\n--- Body (Markdown after frontmatter) ---\n%s",
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.Description),
		req.Body,
	)
	if len(req.Files) > 0 {
		var sb strings.Builder
		sb.WriteString("\n\n--- Supporting files (paths only, not contents) ---\n")
		for _, f := range req.Files {
			kind := "text"
			if f.IsBinary {
				kind = "binary"
			}
			fmt.Fprintf(&sb, "- %s (%s, %d bytes)\n", f.Path, kind, f.SizeBytes)
		}
		userMsg += sb.String()
	}

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	start := time.Now()
	raw, err := a.callClaude(ctx, skillvalidation.SystemPrompt, userMsg)
	metrics.ClaudeValidationDuration.Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.ClaudeValidationTotal.WithLabelValues("error").Inc()
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	report, err := skillvalidation.Parse(raw)
	if err != nil {
		metrics.ClaudeValidationTotal.WithLabelValues("error").Inc()
		writeErr(w, http.StatusBadGateway, "could not parse Claude response: "+err.Error())
		return
	}
	metrics.ClaudeValidationTotal.WithLabelValues("success").Inc()
	writeJSON(w, http.StatusOK, report)
}
