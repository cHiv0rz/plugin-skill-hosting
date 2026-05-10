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
// using the supplied conflict message) or a generic 500.
func respondDBOrConflict(w http.ResponseWriter, err error, conflictMsg string) {
	if isUniqueViolation(err) {
		writeErr(w, http.StatusConflict, conflictMsg)
		return
	}
	writeErr(w, http.StatusInternalServerError, "db error")
}
