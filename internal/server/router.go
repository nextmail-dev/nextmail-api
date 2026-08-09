package server

import (
	"net/http"

	"nextmail-api/internal/app/v1"
	"nextmail-api/internal/app/v1/health"
)

// newRouter builds the root HTTP router. It mounts non-versioned routes (such
// as health probes) directly, and each API version as an isolated sub-router
// under its /api/vN prefix.
func newRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()

	// Non-versioned routes: liveness/readiness probes should not be coupled
	// to an API version.
	mux.HandleFunc("GET /health", health.Handle)

	// API v1: each version owns its routes relative to the version prefix.
	v1Mux := http.NewServeMux()
	v1.Register(v1Mux, v1.Deps{GeoIP: deps.GeoIP})
	mux.Handle("/api/v1/", http.StripPrefix("/api/v1", v1Mux))

	// Future versions mount the same way, e.g.:
	// v2Mux := http.NewServeMux()
	// v2.Register(v2Mux, v2.Deps{...})
	// mux.Handle("/api/v2/", http.StripPrefix("/api/v2", v2Mux))

	return mux
}
