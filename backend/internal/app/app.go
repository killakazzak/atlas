// Package app wires Atlas process dependencies and owns application lifecycle.
package app

import (
	"context"
	"fmt"
	"log/slog"

	"atlas/internal/config"
	atlashttp "atlas/internal/http"
	"atlas/internal/inventory"
	"atlas/internal/logger"
)

// App owns runtime dependencies and the HTTP server.
type App struct {
	cfg       config.Config
	logger    *slog.Logger
	inventory inventory.Service
	server    *atlashttp.Server
}

// New constructs an App with explicitly wired dependencies.
func New(cfg config.Config) (*App, error) {
	log := logger.New()

	serverRepo := inventory.NewMemoryServerRepository()
	inventoryService := inventory.NewService(serverRepo)
	httpServer := atlashttp.New(cfg, log, inventoryService)

	return &App{
		cfg:       cfg,
		logger:    log,
		inventory: inventoryService,
		server:    httpServer,
	}, nil
}

// Run starts the HTTP server and blocks until it stops.
func (a *App) Run() error {
	if err := a.server.Run(); err != nil {
		return fmt.Errorf("run http server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server.
func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return nil
}
