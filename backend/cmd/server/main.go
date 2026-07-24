// Command server is the entrypoint for the Atlas HTTP API.
//
// Project layout (under backend/):
//
//	cmd/server         — process entrypoint; loads config and starts the app
//	internal/app       — composition root; wires dependencies and lifecycle
//	internal/config    — environment-based configuration
//	internal/http      — HTTP routing and server lifecycle
//	internal/inventory — inventory domain, service, and temporary memory store
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

	"atlas/internal/app"
	"atlas/internal/config"
	"atlas/internal/logger"
)

func main() {
	log := logger.New()
	cfg := config.Load()

	application, err := app.New(cfg)
	if err != nil {
		log.Error("failed to create app", "error", err)
		os.Exit(1)
	}

	if err := application.Run(); err != nil {
		log.Error("server stopped", "error", err)
		os.Exit(1)
	}
}
