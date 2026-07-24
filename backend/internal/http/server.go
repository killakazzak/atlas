// Package http wires HTTP routes and runs the Atlas API server.
// The package is named http to match the directory; the standard library
// is imported as stdhttp to avoid a name collision.
package http

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	stdhttp "net/http"

	"atlas/internal/config"
	atlashttp "atlas/internal/http/middleware"
	"atlas/internal/inventory"
	inventoryhttp "atlas/internal/inventory/http"
	"atlas/internal/version"
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

	chain := atlashttp.Chain(
		atlashttp.Recovery(logger),
		atlashttp.RequestID,
		atlashttp.Logging(logger),
	)
	s.http.Handler = chain(mux)

	return s
}

func (s *Server) routes(inventoryService inventory.Service) {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /version", s.handleVersion)
	inventoryhttp.NewHandler(inventoryService).Register(s.mux)
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
	writeJSON(w, stdhttp.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleVersion(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, version.Current())
}

func writeJSON(w stdhttp.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
