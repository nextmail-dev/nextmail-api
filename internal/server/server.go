// Package server assembles the HTTP server: it wires the middleware chain
// around the root router and manages graceful shutdown.
package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"nextmail-api/internal/app/v1/geo"
	"nextmail-api/internal/config"
	"nextmail-api/internal/middleware"
)

// Deps holds the application-wide dependencies injected into the server. Each
// field is consumed by one or more API modules.
type Deps struct {
	GeoIP geo.Lookuper
}

// Server is the HTTP API server.
type Server struct {
	cfg  config.Config
	deps Deps
}

// New creates a Server from the given configuration and dependencies.
func New(cfg config.Config, deps Deps) *Server {
	return &Server{cfg: cfg, deps: deps}
}

// Run starts the HTTP server and blocks until a shutdown signal is received or
// the server fails to start.
func (s *Server) Run() error {
	handler := middleware.Recover(middleware.Logger(newRouter(s.deps)))

	httpServer := &http.Server{
		Addr:         ":" + s.cfg.Port,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on :%s (env=%s)", s.cfg.Port, s.cfg.Env)
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-stop:
	}

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return httpServer.Shutdown(ctx)
}
