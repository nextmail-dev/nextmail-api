// Package health implements a simple service health endpoint.
package health

import (
	"net/http"

	"nextmail-api/internal/platform/web"
)

// Handle responds with the service health.
func Handle(w http.ResponseWriter, r *http.Request) {
	web.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Register mounts the health endpoint on the given mux at GET /health.
func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /health", Handle)
}
