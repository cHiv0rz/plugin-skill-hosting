package main

import (
	"errors"
	"strings"
)

// googleIssuers are the issuer values Google Sign-In emits in ID tokens.
// Source: https://accounts.google.com/.well-known/openid-configuration
var googleIssuers = map[string]struct{}{
	"https://accounts.google.com": {},
	"accounts.google.com":         {},
}

// errWorkspaceDomainRejected is returned when the `hd` claim is missing or not
// in the allowlist. Callers map this to HTTP 401; the message is intentionally
// generic so we don't leak the configured domains.
var errWorkspaceDomainRejected = errors.New("workspace domain not allowed")

// isGoogleIssuer reports whether the given issuer string is Google's. Only
// Google embeds the `hd` (hosted domain) claim, so on non-Google issuers we
// don't apply the workspace check — that lets generic test IdPs (Keycloak,
// Auth0 dev tenants, etc.) work without any allowlist setup.
func isGoogleIssuer(issuer string) bool {
	_, ok := googleIssuers[strings.ToLower(strings.TrimSpace(issuer))]
	return ok
}

// validateGoogleWorkspaceHD enforces the configured Google Workspace domain
// allowlist against the `hd` claim of an ID token.
//
// Behaviour:
//   - Issuer is not Google → always allowed (generic OIDC provider).
//   - Allowlist is empty   → always allowed (opt-in restriction).
//   - Otherwise the lower-cased `hd` value must appear in the allowlist.
//
// A missing `hd` on a Google token means a personal Gmail / consumer account
// and is rejected when the allowlist is non-empty.
func validateGoogleWorkspaceHD(issuer, hd string, allowed []string) error {
	if !isGoogleIssuer(issuer) {
		return nil
	}
	if len(allowed) == 0 {
		return nil
	}
	hd = strings.ToLower(strings.TrimSpace(hd))
	if hd == "" {
		return errWorkspaceDomainRejected
	}
	for _, d := range allowed {
		if strings.ToLower(strings.TrimSpace(d)) == hd {
			return nil
		}
	}
	return errWorkspaceDomainRejected
}
