// Package semver implements the project's version-bump rules: a small,
// deliberately-loose semver helper used to advance plugin versions on edits.
package semver

import (
	"fmt"
	"strconv"
	"strings"
)

type BumpKind int

const (
	BumpMajor BumpKind = iota
	BumpMinor
	BumpPatch
)

func Parse(v string) (int, int, int) {
	parts := strings.SplitN(strings.TrimSpace(v), ".", 3)
	if len(parts) != 3 {
		return 0, 0, 0
	}
	mj, e1 := strconv.Atoi(parts[0])
	mn, e2 := strconv.Atoi(parts[1])
	pt, e3 := strconv.Atoi(parts[2])
	if e1 != nil || e2 != nil || e3 != nil {
		return 0, 0, 0
	}
	return mj, mn, pt
}

// BumpKindForSizeChange picks a bump for a content edit: minor when the size
// changed by more than 30% relative to the old size, otherwise patch. An empty
// old body is treated as a >30% change for any non-empty new body.
func BumpKindForSizeChange(oldSize, newSize int) BumpKind {
	if oldSize == 0 {
		if newSize == 0 {
			return BumpPatch
		}
		return BumpMinor
	}
	delta := newSize - oldSize
	if delta < 0 {
		delta = -delta
	}
	if delta*10 > oldSize*3 {
		return BumpMinor
	}
	return BumpPatch
}

func Bump(current string, kind BumpKind) string {
	mj, mn, pt := Parse(current)
	switch kind {
	case BumpMajor:
		return fmt.Sprintf("%d.0.0", mj+1)
	case BumpMinor:
		return fmt.Sprintf("%d.%d.0", mj, mn+1)
	case BumpPatch:
		return fmt.Sprintf("%d.%d.%d", mj, mn, pt+1)
	}
	return current
}
