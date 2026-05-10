package workspaceauth

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
		if got := IsGoogleIssuer(c.in); got != c.want {
			t.Errorf("IsGoogleIssuer(%q) = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestValidateGoogleHD(t *testing.T) {
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
			err := ValidateGoogleHD(c.issuer, c.hd, c.allowed)
			if c.wantErr && err == nil {
				t.Fatal("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
