// Command server is the entrypoint for the Atlas HTTP API.
//
// Project layout (under backend/):
//
//	cmd/server         — process entrypoint; loads config and starts the server
//	internal/config    — environment-based configuration
//	internal/http      — HTTP routing and server lifecycle
//	internal/logger    — structured logging setup
//	internal/version   — service name and version metadata
//	pkg/               — reusable libraries safe for external consumers
//	api/               — API contracts (OpenAPI specs, shared shapes)
//	configs/           — configuration files and examples
//
// This MVP uses only the Go standard library: no database, messaging, auth, or Docker.
package main

import (
	"os"

	"atlas/internal/config"
	atlashttp "atlas/internal/http"
	"atlas/internal/logger"
)

func main() {
	log := logger.New()
	cfg := config.Load()

	srv := atlashttp.New(cfg, log)
	if err := srv.Run(); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
