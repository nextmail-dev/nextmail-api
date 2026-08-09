// Package v1 assembles all API version 1 modules and registers them on a
// single mux. Paths registered here are relative to the /api/v1 prefix.
//
// To add a new v1 endpoint:
//  1. Create a package under internal/app/v1/<feature>/ exposing a Register
//     function that mounts its routes on a *http.ServeMux.
//  2. Add one line in Register below wiring it up.
package v1

import (
	"net/http"

	"nextmail-api/internal/app/v1/geo"
	"nextmail-api/internal/app/v1/health"
)

// Deps holds the dependencies required by v1 API modules. Add fields as new
// modules require new infrastructure (DB, cache, etc.).
type Deps struct {
	GeoIP geo.Lookuper
}

// Register mounts all v1 API modules on the given mux.
func Register(mux *http.ServeMux, deps Deps) {
	geo.Register(mux, deps.GeoIP)
	health.Register(mux)
}
