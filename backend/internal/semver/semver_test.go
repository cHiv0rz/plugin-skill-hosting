package semver

import "testing"

func TestParse(t *testing.T) {
	cases := []struct {
		in         string
		mj, mn, pt int
	}{
		{"1.2.3", 1, 2, 3},
		{"  0.0.0  ", 0, 0, 0},
		{"10.20.30", 10, 20, 30},
		{"1.2", 0, 0, 0},
		{"a.b.c", 0, 0, 0},
		{"", 0, 0, 0},
		{"1.2.3.4", 0, 0, 0},
	}
	for _, c := range cases {
		mj, mn, pt := Parse(c.in)
		if mj != c.mj || mn != c.mn || pt != c.pt {
			t.Errorf("Parse(%q) = %d.%d.%d, want %d.%d.%d", c.in, mj, mn, pt, c.mj, c.mn, c.pt)
		}
	}
}

func TestBumpKindForSizeChange(t *testing.T) {
	cases := []struct {
		old, new int
		want     BumpKind
	}{
		{0, 0, BumpPatch},
		{0, 1, BumpMinor},
		{100, 100, BumpPatch},
		{100, 110, BumpPatch},
		{100, 130, BumpPatch}, // exactly 30% — not >30%
		{100, 131, BumpMinor},
		{100, 50, BumpMinor},
		{1000, 999, BumpPatch},
	}
	for _, c := range cases {
		got := BumpKindForSizeChange(c.old, c.new)
		if got != c.want {
			t.Errorf("BumpKindForSizeChange(%d, %d) = %v, want %v", c.old, c.new, got, c.want)
		}
	}
}

func TestBump(t *testing.T) {
	cases := []struct {
		current string
		kind    BumpKind
		want    string
	}{
		{"1.2.3", BumpMajor, "2.0.0"},
		{"1.2.3", BumpMinor, "1.3.0"},
		{"1.2.3", BumpPatch, "1.2.4"},
		{"0.0.0", BumpPatch, "0.0.1"},
		{"garbage", BumpMinor, "0.1.0"}, // Parse returns zeros
		{"9.9.9", BumpMajor, "10.0.0"},
		{"0.99.99", BumpMinor, "0.100.0"},
	}
	for _, c := range cases {
		got := Bump(c.current, c.kind)
		if got != c.want {
			t.Errorf("Bump(%q, %v) = %q, want %q", c.current, c.kind, got, c.want)
		}
	}
}

func TestBump_UnknownKindReturnsCurrent(t *testing.T) {
	got := Bump("1.2.3", BumpKind(99))
	if got != "1.2.3" {
		t.Errorf("unknown BumpKind should return current; got %q", got)
	}
}

func TestBumpKindForSizeChange_SymmetricThreshold(t *testing.T) {
	// Shrink by exactly 30% — equal to threshold, so still patch.
	if got := BumpKindForSizeChange(100, 70); got != BumpPatch {
		t.Errorf("100→70 should be patch (= 30%% threshold), got %v", got)
	}
	// Shrink by 31% — minor.
	if got := BumpKindForSizeChange(100, 69); got != BumpMinor {
		t.Errorf("100→69 should be minor (>30%%), got %v", got)
	}
}
