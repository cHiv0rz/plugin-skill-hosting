package server

import (
	"net/http"

	"marketplace/internal/buildinfo"
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
		Version:   buildinfo.Version,
		GitCommit: buildinfo.Commit,
		BuildTime: buildinfo.Time,
	})
}
