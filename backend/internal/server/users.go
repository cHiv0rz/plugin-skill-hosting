package server

import (
	"net/http"
	"time"
)

// UserSummary is the public-safe projection of a user row used by the
// directory endpoint. It deliberately omits password_hash, api_token, and
// oidc_* so the listing can never leak secrets.
type UserSummary struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"createdAt"`
}

func (a *App) handleListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := a.DB.QueryContext(r.Context(),
		`SELECT id, username, email, created_at FROM users ORDER BY created_at ASC`)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	defer rows.Close()

	users := []UserSummary{}
	for rows.Next() {
		var u UserSummary
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.CreatedAt); err != nil {
			writeErr(w, http.StatusInternalServerError, "db error")
			return
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		writeErr(w, http.StatusInternalServerError, "db error")
		return
	}
	writeJSON(w, http.StatusOK, users)
}
