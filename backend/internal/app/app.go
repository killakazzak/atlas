// Package app wires Atlas process dependencies and owns application lifecycle.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"atlas/internal/auth"
	authpg "atlas/internal/auth/postgres"
	"atlas/internal/config"
	"atlas/internal/database"
	atlashttp "atlas/internal/http"
	"atlas/internal/inventory"
	inventorypg "atlas/internal/inventory/postgres"
	"atlas/internal/logger"
)

// poolConnectTimeout bounds how long startup waits for the database.
const poolConnectTimeout = 10 * time.Second

// App owns runtime dependencies and the HTTP server.
type App struct {
	cfg       config.Config
	logger    *slog.Logger
	inventory inventory.Service
	auth      auth.Service
	server    *atlashttp.Server
	pool      *pgxpool.Pool
}

// New constructs an App with explicitly wired dependencies.
//
// Storage is selected from configuration: when DATABASE_URL is set, the
// inventory service is backed by PostgreSQL; otherwise it uses a temporary
// in-memory store. The returned App owns any database pool it opens.
func New(cfg config.Config) (*App, error) {
	log := logger.New()

	serverRepo, pool, err := newServerRepository(cfg, log)
	if err != nil {
		return nil, err
	}

	userRepo := newUserRepository(cfg, pool, log)

	inventoryService := inventory.NewService(serverRepo)
	authService := auth.NewService(userRepo, auth.BcryptHasher{})
	httpServer := atlashttp.New(cfg, log, inventoryService)

	return &App{
		cfg:       cfg,
		logger:    log,
		inventory: inventoryService,
		auth:      authService,
		server:    httpServer,
		pool:      pool,
	}, nil
}

// newUserRepository selects the auth storage backend.
// It reuses an existing pool when available; otherwise it falls back to in-memory.
func newUserRepository(cfg config.Config, pool *pgxpool.Pool, log *slog.Logger) auth.UserRepository {
	if cfg.DatabaseURL != "" && pool != nil {
		log.Info("using PostgreSQL auth store")
		return authpg.NewUserRepository(pool)
	}
	log.Info("using in-memory auth store (DATABASE_URL not set)")
	return auth.NewMemoryUserRepository()
}

// newServerRepository selects the inventory storage backend from configuration.
// It returns the repository and, when PostgreSQL is used, the owning pool.
func newServerRepository(cfg config.Config, log *slog.Logger) (inventory.ServerRepository, *pgxpool.Pool, error) {
	if cfg.DatabaseURL == "" {
		log.Info("using in-memory inventory store (DATABASE_URL not set)")
		return inventory.NewMemoryServerRepository(), nil, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), poolConnectTimeout)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("connect database: %w", err)
	}

	log.Info("using PostgreSQL inventory store")
	return inventorypg.NewServerRepository(pool), pool, nil
}

// Run starts the HTTP server and blocks until it stops.
func (a *App) Run() error {
	if err := a.server.Run(); err != nil {
		return fmt.Errorf("run http server: %w", err)
	}
	return nil
}

// Shutdown gracefully stops the HTTP server and releases the database pool.
func (a *App) Shutdown(ctx context.Context) error {
	err := a.server.Shutdown(ctx)
	if a.pool != nil {
		a.pool.Close()
	}
	if err != nil {
		return fmt.Errorf("shutdown http server: %w", err)
	}
	return nil
}
