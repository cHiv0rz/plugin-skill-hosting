package main

import "net/http"

// Populated at build time via -ldflags "-X main.BuildVersion=... -X main.BuildCommit=... -X main.BuildTime=...".
var (
	BuildVersion = "dev"
	BuildCommit  = "unknown"
	BuildTime    = "unknown"
)

type buildInfoResp struct {
	Name      string `json:"name"`
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
}

func (a *App) handleVersion(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, buildInfoResp{
		Name:      "plugin-skill-hosting-backend",
		Version:   BuildVersion,
		GitCommit: BuildCommit,
		BuildTime: BuildTime,
	})
}
