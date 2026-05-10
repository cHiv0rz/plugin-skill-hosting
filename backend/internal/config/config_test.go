package config

import "testing"

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
