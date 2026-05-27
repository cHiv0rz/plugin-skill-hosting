package server

import (
	"net/http"
	"strings"
)

// isUniqueViolation sniffs a libpq error message for a unique-constraint
// failure; the typed error class isn't exported, so a string match is the
// path of least resistance.
func isUniqueViolation(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

// respondDBOrConflict maps a write error to either 409 (unique violation,
// using the supplied conflict message) or a logged 500. The non-conflict
// branch routes through serverErr so the underlying error lands in logs
// instead of being swallowed behind "db error".
func respondDBOrConflict(w http.ResponseWriter, r *http.Request, err error, conflictMsg string) {
	if isUniqueViolation(err) {
		writeErr(w, http.StatusConflict, conflictMsg)
		return
	}
	serverErr(w, r, err, "db error")
}
