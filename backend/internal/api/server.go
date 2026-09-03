// Package api wires the Chi router, middleware, and HTTP handlers.
package api

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Server bundles the HTTP server, router, and its dependencies.
type Server struct {
	router    http.Handler
	pool      *pgxpool.Pool
	products  ProductStore
	inventory InventoryStore
	logger    *slog.Logger
	version   string
}

// New assembles the API server with middleware and routes.
func New(pool *pgxpool.Pool, products ProductStore, inventory InventoryStore, logger *slog.Logger, version string) *Server {
	r := chi.NewRouter()
	s := &Server{
		router:    r,
		pool:      pool,
		products:  products,
		inventory: inventory,
		logger:    logger,
		version:   version,
	}
	r.Use(middleware.RequestID)
	r.Use(requestLogger(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(corsMiddleware())
	r.Get("/health/live", s.healthLive)
	r.Get("/health/ready", s.healthReady)
	if s.products != nil && s.inventory != nil {
		s.registerCatalogRoutes(r)
	}
	return s
}

// ServeHTTP delegates to the configured router.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

// Handler returns the configured http.Handler for use in tests.
func (s *Server) Handler() http.Handler { return s.router }

func (s *Server) healthLive(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Version: s.version})
}

func (s *Server) healthReady(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()
	if err := s.pool.Ping(ctx); err != nil {
		s.logger.ErrorContext(ctx, "readiness probe failed", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "unavailable"})
		return
	}
	writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Version: s.version})
}

type healthResponse struct {
	Status  string `json:"status"`
	Version string `json:"version,omitempty"`
}
