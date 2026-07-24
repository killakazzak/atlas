// Package http wires HTTP routes and runs the Atlas API server.
// The package is named http to match the directory; the standard library
// is imported as stdhttp to avoid a name collision.
package http

import (
	"context"
	"errors"
	"log/slog"
	stdhttp "net/http"

	"atlas/internal/config"
	"atlas/internal/http/middleware"
	"atlas/internal/httpx"
	"atlas/internal/inventory"
	inventoryhttp "atlas/internal/inventory/http"
	"atlas/internal/version"
	atlasdocs "atlas/openapi"
)

// Server owns the HTTP mux and listen configuration.
type Server struct {
	cfg    config.Config
	logger *slog.Logger
	mux    *stdhttp.ServeMux
	http   *stdhttp.Server
}

// New constructs a Server with registered routes.
func New(cfg config.Config, logger *slog.Logger, inventoryService inventory.Service) *Server {
	mux := stdhttp.NewServeMux()
	s := &Server{
		cfg:    cfg,
		logger: logger,
		mux:    mux,
		http: &stdhttp.Server{
			Addr:    cfg.Addr(),
			Handler: mux,
		},
	}
	s.routes(inventoryService)

	chain := middleware.Chain(
		middleware.Recovery(logger),
		middleware.RequestID,
		middleware.Logging(logger),
	)
	s.http.Handler = chain(mux)

	return s
}

func (s *Server) routes(inventoryService inventory.Service) {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /version", s.handleVersion)
	inventoryhttp.NewHandler(inventoryService, s.logger).Register(s.mux)
	atlasdocs.Register(s.mux)
}

// Handler returns the server's root http.Handler. Useful for testing.
func (s *Server) Handler() stdhttp.Handler {
	return s.http.Handler
}

// Run starts the HTTP listener and blocks until it fails or Shutdown is called.
func (s *Server) Run() error {
	s.logger.Info("starting HTTP server", "addr", s.http.Addr)
	err := s.http.ListenAndServe()
	if errors.Is(err, stdhttp.ErrServerClosed) {
		return nil
	}
	return err
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

func (s *Server) handleHealth(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	httpx.WriteJSON(w, stdhttp.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleVersion(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	httpx.WriteJSON(w, stdhttp.StatusOK, version.Current())
}
