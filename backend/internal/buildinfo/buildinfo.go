// Package buildinfo holds build-time identification populated via -ldflags.
package buildinfo

// Populated at build time via:
//
//	-ldflags "-X marketplace/internal/buildinfo.Version=...
//	          -X marketplace/internal/buildinfo.Commit=...
//	          -X marketplace/internal/buildinfo.Time=..."
var (
	Version = "dev"
	Commit  = "unknown"
	Time    = "unknown"
)
