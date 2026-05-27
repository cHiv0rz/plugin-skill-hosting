package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
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
	StopReason string `json:"stop_reason,omitempty"`
	Error      *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// callClaude sends a single user turn to the Claude messages API and returns
// the model's text reply. Some newer models reject assistant-prefill, so we
// rely on prompt engineering + a tolerant JSON extractor on the caller side
// instead of pinning the response with a leading `{`. Returns a clear error
// when the model hit max_tokens, since truncated JSON is the #1 cause of
// downstream "no JSON object found" parse failures.
func (a *App) callClaude(ctx context.Context, system, user string, maxTokens int) (string, error) {
	if strings.TrimSpace(a.Cfg.AnthropicAPIKey) == "" {
		return "", errors.New("Claude API not configured (set ANTHROPIC_API_KEY)")
	}
	payload, err := json.Marshal(claudeRequest{
		Model:     a.Cfg.AnthropicModel,
		MaxTokens: maxTokens,
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
	if cr.StopReason == "max_tokens" {
		return sb.String(), fmt.Errorf("response truncated at max_tokens=%d — increase the limit", maxTokens)
	}
	return sb.String(), nil
}

// logClaudeParseFailure records the raw model output (truncated) so prompt or
// max_tokens regressions can be diagnosed without re-running the call.
func logClaudeParseFailure(endpoint string, raw string, err error) {
	const cap = 600
	snippet := raw
	if len(snippet) > cap {
		snippet = snippet[:cap] + "…(truncated)"
	}
	log.Printf("claude parse failure: endpoint=%s err=%v raw=%q", endpoint, err, snippet)
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
	raw, err := a.callClaude(ctx, skillvalidation.SystemPrompt, userMsg, 2048)
	metrics.ClaudeValidationDuration.Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.ClaudeValidationTotal.WithLabelValues("error").Inc()
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}

	report, err := skillvalidation.Parse(raw)
	if err != nil {
		metrics.ClaudeValidationTotal.WithLabelValues("error").Inc()
		logClaudeParseFailure("validate", raw, err)
		writeErr(w, http.StatusBadGateway, "could not parse Claude response: "+err.Error())
		return
	}
	metrics.ClaudeValidationTotal.WithLabelValues("success").Inc()
	writeJSON(w, http.StatusOK, report)
}

type fixFindingRequest struct {
	Name             string                  `json:"name"`
	Description      string                  `json:"description"`
	Body             string                  `json:"body"`
	ExtraFrontmatter string                  `json:"extraFrontmatter,omitempty"`
	Files            []SkillFileSummary      `json:"files,omitempty"`
	Finding          skillvalidation.Finding `json:"finding"`
}

func (a *App) handleFixFinding(w http.ResponseWriter, r *http.Request) {
	var req fixFindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid json")
		return
	}
	if strings.TrimSpace(req.Finding.Title) == "" && strings.TrimSpace(req.Finding.Detail) == "" {
		writeErr(w, http.StatusBadRequest, "finding is required")
		return
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "Skill name: %s\n\n--- Description ---\n%s\n\n--- Body (Markdown after frontmatter) ---\n%s",
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.Description),
		req.Body,
	)
	if strings.TrimSpace(req.ExtraFrontmatter) != "" {
		fmt.Fprintf(&sb, "\n\n--- Extra frontmatter (YAML lines) ---\n%s", req.ExtraFrontmatter)
	}
	if len(req.Files) > 0 {
		sb.WriteString("\n\n--- Supporting files (paths only, not contents) ---\n")
		for _, f := range req.Files {
			kind := "text"
			if f.IsBinary {
				kind = "binary"
			}
			fmt.Fprintf(&sb, "- %s (%s, %d bytes)\n", f.Path, kind, f.SizeBytes)
		}
	}
	fmt.Fprintf(&sb, "\n\n--- Finding to fix ---\nSeverity: %s\nTitle: %s\nDetail: %s",
		req.Finding.Severity, req.Finding.Title, req.Finding.Detail,
	)

	ctx, cancel := context.WithTimeout(r.Context(), 90*time.Second)
	defer cancel()

	start := time.Now()
	// Fix responses can contain the FULL rewritten body, which easily exceeds
	// the validate-call budget. Be generous so we don't truncate the JSON.
	raw, err := a.callClaude(ctx, skillvalidation.FixSystemPrompt, sb.String(), 16384)
	metrics.ClaudeFindingFixDuration.Observe(time.Since(start).Seconds())
	if err != nil {
		metrics.ClaudeFindingFixTotal.WithLabelValues("error").Inc()
		writeErr(w, http.StatusBadGateway, err.Error())
		return
	}
	fix, err := skillvalidation.ParseFix(raw)
	if err != nil {
		metrics.ClaudeFindingFixTotal.WithLabelValues("error").Inc()
		logClaudeParseFailure("finding-fix", raw, err)
		writeErr(w, http.StatusBadGateway, "could not parse Claude response: "+err.Error())
		return
	}
	metrics.ClaudeFindingFixTotal.WithLabelValues("success").Inc()
	writeJSON(w, http.StatusOK, fix)
}
