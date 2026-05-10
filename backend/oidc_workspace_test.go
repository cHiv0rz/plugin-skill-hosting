package main

import "testing"

func TestIsGoogleIssuer(t *testing.T) {
	cases := []struct {
		in   string
		want bool
	}{
		{"https://accounts.google.com", true},
		{"accounts.google.com", true},
		{"  HTTPS://Accounts.Google.com  ", true},
		{"https://login.microsoftonline.com/common/v2.0", false},
		{"https://accounts.google.com/", false}, // trailing slash is not what Google issues
		{"", false},
	}
	for _, c := range cases {
		if got := isGoogleIssuer(c.in); got != c.want {
			t.Errorf("isGoogleIssuer(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestValidateGoogleWorkspaceHD(t *testing.T) {
	allowed := []string{"yourcompany.com", "Other.Co"}

	cases := []struct {
		name    string
		issuer  string
		hd      string
		allowed []string
		wantErr bool
	}{
		{"allowed domain", "https://accounts.google.com", "yourcompany.com", allowed, false},
		{"allowed second domain", "https://accounts.google.com", "other.co", allowed, false},
		{"case-insensitive hd", "https://accounts.google.com", "YourCompany.com", allowed, false},
		{"disallowed domain", "https://accounts.google.com", "evil.com", allowed, true},
		{"missing hd", "https://accounts.google.com", "", allowed, true},
		{"whitespace-only hd", "https://accounts.google.com", "   ", allowed, true},
		{"empty allowlist skips check (Google)", "https://accounts.google.com", "", nil, false},
		{"empty allowlist skips check (personal Gmail)", "https://accounts.google.com", "", []string{}, false},
		{"non-Google issuer is always allowed", "https://login.microsoftonline.com/common/v2.0", "", allowed, false},
		{"non-Google issuer with arbitrary hd", "https://idp.example.com", "evil.com", allowed, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := validateGoogleWorkspaceHD(c.issuer, c.hd, c.allowed)
			if c.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseDomainList(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{"", nil},
		{"   ", nil},
		{"yourcompany.com", []string{"yourcompany.com"}},
		{"a.com,b.com", []string{"a.com", "b.com"}},
		{" A.com , B.COM ,, ", []string{"a.com", "b.com"}},
	}
	for _, c := range cases {
		got := parseDomainList(c.in)
		if len(got) != len(c.want) {
			t.Errorf("parseDomainList(%q) len = %d, want %d (%v vs %v)", c.in, len(got), len(c.want), got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("parseDomainList(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}
