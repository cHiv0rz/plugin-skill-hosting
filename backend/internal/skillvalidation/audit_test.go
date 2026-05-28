package skillvalidation

import "testing"

func TestParseAudit_Clean(t *testing.T) {
	in := `{"riskScore":0,"riskLevel":"low","categories":[],"summary":"benign","findings":[]}`
	rep, err := ParseAudit(in)
	if err != nil {
		t.Fatalf("ParseAudit: %v", err)
	}
	if rep.RiskScore != 0 || rep.RiskLevel != "low" || len(rep.Findings) != 0 {
		t.Errorf("unexpected report: %+v", rep)
	}
}

func TestParseAudit_RecomputesLevelFromScore(t *testing.T) {
	// Model claims "low" but score says critical — server must trust the score.
	in := `{"riskScore":92,"riskLevel":"low","categories":["malicious-code"],"summary":"bad","findings":[]}`
	rep, err := ParseAudit(in)
	if err != nil {
		t.Fatalf("ParseAudit: %v", err)
	}
	if rep.RiskLevel != "critical" {
		t.Errorf("riskLevel = %q, want critical", rep.RiskLevel)
	}
}

func TestParseAudit_ClampsScore(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want int
	}{
		{`{"riskScore":150,"findings":[]}`, 100},
		{`{"riskScore":-7,"findings":[]}`, 0},
	} {
		rep, err := ParseAudit(tc.in)
		if err != nil {
			t.Fatalf("ParseAudit(%s): %v", tc.in, err)
		}
		if rep.RiskScore != tc.want {
			t.Errorf("score = %d, want %d", rep.RiskScore, tc.want)
		}
	}
}

func TestParseAudit_NormalisesFindingSeverity(t *testing.T) {
	in := `{"riskScore":60,"findings":[
		{"category":"destructive","severity":"HIGH","detail":"rm -rf"},
		{"category":"x","severity":"bogus","detail":"y"}
	]}`
	rep, err := ParseAudit(in)
	if err != nil {
		t.Fatalf("ParseAudit: %v", err)
	}
	if rep.Findings[0].Severity != "high" {
		t.Errorf("findings[0].Severity = %q, want high", rep.Findings[0].Severity)
	}
	if rep.Findings[1].Severity != "medium" {
		t.Errorf("findings[1].Severity = %q, want medium (default)", rep.Findings[1].Severity)
	}
}

func TestParseAudit_StripsCodeFence(t *testing.T) {
	in := "```json\n{\"riskScore\":10,\"findings\":[]}\n```"
	rep, err := ParseAudit(in)
	if err != nil {
		t.Fatalf("ParseAudit: %v", err)
	}
	if rep.RiskScore != 10 || rep.RiskLevel != "low" {
		t.Errorf("unexpected: %+v", rep)
	}
}

func TestRiskLevelForScore(t *testing.T) {
	cases := map[int]string{0: "low", 24: "low", 25: "medium", 49: "medium", 50: "high", 79: "high", 80: "critical", 100: "critical"}
	for score, want := range cases {
		if got := RiskLevelForScore(score); got != want {
			t.Errorf("RiskLevelForScore(%d) = %q, want %q", score, got, want)
		}
	}
}
