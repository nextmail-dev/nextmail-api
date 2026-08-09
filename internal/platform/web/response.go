// Package web provides shared HTTP helpers for writing consistent JSON
// responses across all API handlers.
package web

import (
	"encoding/json"
	"net/http"
)

// WriteJSON encodes body as JSON and writes it with the given status code.
func WriteJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

// Error writes a standard error response with shape {"error": msg}.
func Error(w http.ResponseWriter, status int, msg string) {
	WriteJSON(w, status, map[string]string{"error": msg})
}
