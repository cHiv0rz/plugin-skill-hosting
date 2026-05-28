// Package config loads and validates process configuration from environment
// variables. It is the only place getenv-style defaults live.
package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL   string
	JWTSecret     string
	ListenAddr    string
	DataDir       string
	PublicBaseURL string

	MarketplaceName string
	DefaultLicense  string

	AuthMode string // "password" (default) or "oidc"

	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string // defaults to PublicBaseURL + "/auth/callback/oidc"
	OIDCScopes       string // space-separated; defaults to "openid email profile"

	// AllowedGoogleWorkspaceDomains, when non-empty, restricts Google sign-in
	// to ID tokens whose `hd` claim is in this list. Only applied when the
	// issuer is Google — generic OIDC providers (e.g. dev/test IdPs) are not
	// affected, so an empty list also means "no restriction".
	AllowedGoogleWorkspaceDomains []string

	AnthropicAPIKey string
	AnthropicModel  string

	// MetricsToken, when non-empty, gates /metrics with Bearer auth. Default
	// is open — relies on the public ingress not routing /metrics.
	MetricsToken string

	// RematerializeOnStartup, when true, re-builds all git repos from the
	// database in a background goroutine after the server starts. Use this
	// when the data dir is ephemeral (emptyDir / no PVC).
	RematerializeOnStartup bool

	// ExternalGitRemoteURL, when non-empty, enables one-way sync of the whole
	// marketplace to a single external git repository (GitHub, GitLab, etc.).
	// On every plugin materialize or delete the backend rewrites the
	// plugins/<name>/ subtree in a checked-out clone of this remote, commits,
	// and pushes. Internal per-plugin repos under /data/repos/ are unaffected.
	// Push failures always log a WARN and let the internal write succeed —
	// the database remains the source of truth.
	ExternalGitRemoteURL string
	ExternalGitBranch    string
	ExternalGitUsername  string
	ExternalGitToken     string

	// MCPOAuthClientID / MCPOAuthClientSecret enable OAuth 2.1 Authorization
	// Code + PKCE on the /mcp endpoint. Both must be set or both empty.
	// MCPOAuthRedirectURIs is the allowlist of callback URIs the OAuth client
	// may request; defaults to the two well-known Claude callback URLs.
	MCPOAuthClientID     string
	MCPOAuthClientSecret string
	MCPOAuthRedirectURIs []string

	// Skill security audit. When AuditEnabled is true and an Anthropic API key
	// is configured, a background job re-evaluates every skill on AuditInterval
	// for malicious/harmful behavior and stores the verdict. Results whose risk
	// score reaches AuditThreshold (0-100) trigger an alert email to
	// AuditAlertEmails (requires SMTP to be configured).
	AuditEnabled     bool
	AuditInterval    time.Duration
	AuditThreshold   int
	AuditAlertEmails []string

	// SMTP settings for outbound notification email. Empty SMTPHost disables
	// all email; the audit job then logs alerts instead of sending them.
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	SMTPFrom     string
	SMTPUseTLS   bool
}

// RequiresUserApproval reports whether new users must be approved by an
// existing approved user before they can access the system. The flow is
// engaged only for OIDC mode without a Google Workspace domain allowlist —
// password and "corporate" (domain-restricted) OIDC deployments still admit
// users immediately.
func (c Config) RequiresUserApproval() bool {
	return c.AuthMode == "oidc" && len(c.AllowedGoogleWorkspaceDomains) == 0
}

func Load() Config {
	c := Config{
		DatabaseURL:   getenv("DATABASE_URL", "postgres://marketplace:marketplace@localhost:5432/marketplace?sslmode=disable"),
		JWTSecret:     getenv("JWT_SECRET", "dev-secret-change-me-please-32-chars-min"),
		ListenAddr:    getenv("LISTEN_ADDR", ":8080"),
		DataDir:       getenv("DATA_DIR", "./data"),
		PublicBaseURL: getenv("PUBLIC_BASE_URL", "http://localhost:8080"),

		MarketplaceName: getenv("MARKETPLACE_NAME", "oglimmer-marketplace"),
		DefaultLicense:  getenv("DEFAULT_LICENSE", "MIT"),

		AuthMode: strings.ToLower(getenv("AUTH_MODE", "password")),

		OIDCIssuerURL:    strings.TrimRight(getenv("OIDC_ISSUER_URL", ""), "/"),
		OIDCClientID:     getenv("OIDC_CLIENT_ID", ""),
		OIDCClientSecret: getenv("OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:  getenv("OIDC_REDIRECT_URL", ""),
		OIDCScopes:       getenv("OIDC_SCOPES", "openid email profile"),

		AllowedGoogleWorkspaceDomains: parseDomainList(getenv("OIDC_GOOGLE_WORKSPACE_DOMAINS", "")),

		AnthropicAPIKey: getenv("ANTHROPIC_API_KEY", ""),
		AnthropicModel:  getenv("ANTHROPIC_MODEL", "claude-sonnet-4-6"),

		MetricsToken: getenv("METRICS_TOKEN", ""),

		RematerializeOnStartup: os.Getenv("REMATERIALIZE_ON_STARTUP") == "true",

		ExternalGitRemoteURL: strings.TrimSpace(getenv("EXTERNAL_GIT_REMOTE_URL", "")),
		ExternalGitBranch:    strings.TrimSpace(getenv("EXTERNAL_GIT_BRANCH", "main")),
		ExternalGitUsername:  getenv("EXTERNAL_GIT_USERNAME", "x-access-token"),
		ExternalGitToken:     getenv("EXTERNAL_GIT_TOKEN", ""),

		MCPOAuthClientID:     getenv("MCP_OAUTH_CLIENT_ID", ""),
		MCPOAuthClientSecret: getenv("MCP_OAUTH_CLIENT_SECRET", ""),
		MCPOAuthRedirectURIs: parseURIList(getenv("MCP_OAUTH_REDIRECT_URIS",
			"https://claude.ai/api/mcp/auth_callback,https://claude.ai/api/auth/callback")),

		AuditEnabled:     getenv("AUDIT_ENABLED", "true") != "false",
		AuditInterval:    parseDuration(getenv("AUDIT_INTERVAL", "24h"), 24*time.Hour),
		AuditThreshold:   parseInt(getenv("AUDIT_ALERT_THRESHOLD", "70"), 70),
		AuditAlertEmails: parseDomainList(getenv("AUDIT_ALERT_EMAILS", "")),

		SMTPHost:     strings.TrimSpace(getenv("SMTP_HOST", "")),
		SMTPPort:     parseInt(getenv("SMTP_PORT", "587"), 587),
		SMTPUsername: getenv("SMTP_USERNAME", ""),
		SMTPPassword: getenv("SMTP_PASSWORD", ""),
		SMTPFrom:     strings.TrimSpace(getenv("SMTP_FROM", "")),
		SMTPUseTLS:   getenv("SMTP_USE_TLS", "true") == "true",
	}
	if c.AuditThreshold < 0 {
		c.AuditThreshold = 0
	}
	if c.AuditThreshold > 100 {
		c.AuditThreshold = 100
	}
	if c.ExternalGitBranch == "" {
		c.ExternalGitBranch = "main"
	}
	if (c.MCPOAuthClientID == "") != (c.MCPOAuthClientSecret == "") {
		log.Fatalf("MCP_OAUTH_CLIENT_ID and MCP_OAUTH_CLIENT_SECRET must both be set or both be empty")
	}
	if c.AuthMode != "password" && c.AuthMode != "oidc" {
		log.Fatalf("AUTH_MODE must be 'password' or 'oidc', got %q", c.AuthMode)
	}
	// DataDir is used as a git remote URL for the per-plugin work tree's
	// `origin`; a relative path would be resolved against the work tree's
	// cwd at push time and fail. Absolute paths from Docker/Helm pass
	// through unchanged.
	if abs, err := filepath.Abs(c.DataDir); err == nil {
		c.DataDir = abs
	}
	if c.OIDCRedirectURL == "" {
		c.OIDCRedirectURL = strings.TrimRight(c.PublicBaseURL, "/") + "/api/auth/oidc/callback"
	}
	if c.AuthMode == "oidc" && len(c.AllowedGoogleWorkspaceDomains) == 0 {
		log.Printf("WARN: AUTH_MODE=oidc but OIDC_GOOGLE_WORKSPACE_DOMAINS is empty — Google Workspace domain restriction is disabled")
	}
	return c
}

// parseDuration parses a Go duration string (e.g. "24h", "168h", "30m"),
// falling back to def on empty or invalid input.
func parseDuration(s string, def time.Duration) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		log.Printf("WARN: invalid duration %q, using %s", s, def)
		return def
	}
	return d
}

// parseInt parses a base-10 integer, falling back to def on empty or invalid input.
func parseInt(s string, def int) int {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		log.Printf("WARN: invalid integer %q, using %d", s, def)
		return def
	}
	return n
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parseURIList splits a comma-separated list of URIs, trimming whitespace.
// Unlike parseDomainList it does NOT lowercase, because URI paths are case-sensitive.
func parseURIList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseDomainList(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
