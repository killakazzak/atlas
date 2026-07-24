// Command server is the entrypoint for the Atlas HTTP API.
//
// Project layout (under backend/):
//
//	cmd/server  — process entrypoint (this package); wires HTTP and starts the process
//	internal/   — private application code (not importable by other modules)
//	pkg/        — reusable libraries safe for external consumers
//	api/        — API contracts (OpenAPI specs, shared request/response shapes)
//	configs/    — configuration files and examples for local/runtime settings
//
// This MVP uses only the Go standard library: no database, messaging, auth, or Docker.
package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", healthHandler)

	addr := ":8080"
	logger.Info("starting HTTP server", "addr", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

// healthHandler responds with a simple readiness payload for load balancers and probes.
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
