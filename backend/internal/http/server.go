// Package http wires HTTP routes and runs the Atlas API server.
// The package is named http to match the directory; the standard library
// is imported as stdhttp to avoid a name collision.
package http

import (
	"encoding/json"
	"log/slog"
	stdhttp "net/http"

	"atlas/internal/config"
	"atlas/internal/version"
)

// Server owns the HTTP mux and listen configuration.
type Server struct {
	cfg    config.Config
	logger *slog.Logger
	mux    *stdhttp.ServeMux
}

// New constructs a Server with registered routes.
func New(cfg config.Config, logger *slog.Logger) *Server {
	s := &Server{
		cfg:    cfg,
		logger: logger,
		mux:    stdhttp.NewServeMux(),
	}
	s.routes()
	return s
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /version", s.handleVersion)
}

// Run starts the HTTP listener and blocks until it fails.
func (s *Server) Run() error {
	addr := s.cfg.Addr()
	s.logger.Info("starting HTTP server", "addr", addr)
	return stdhttp.ListenAndServe(addr, s.mux)
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
