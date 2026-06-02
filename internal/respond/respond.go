// Package respond writes common HTTP responses, shared by the domains. They're
// flat functions: what you see is what happens, with no state or reflection.
package respond

import (
	"encoding/json"
	"log"
	"net/http"
)

// JSON writes v as JSON with the given status.
func JSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Error writes a {"error": msg} body with the given status.
func Error(w http.ResponseWriter, status int, msg string) {
	JSON(w, status, map[string]string{"error": msg})
}

// ServerError logs the REAL error (with method and path, for debugging) and
// responds with a generic 500 to the client, without leaking internal details.
// Use it in JSON handlers when an operation fails due to the server.
func ServerError(w http.ResponseWriter, r *http.Request, err error) {
	log.Printf("error en %s %s: %v", r.Method, r.URL.Path, err)
	Error(w, http.StatusInternalServerError, "error interno")
}
