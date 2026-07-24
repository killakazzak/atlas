// Package postgres implements inventory persistence with PostgreSQL.
package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"atlas/internal/inventory"
)

// Ensure ServerRepository satisfies inventory.ServerRepository at compile time.
var _ inventory.ServerRepository = (*ServerRepository)(nil)

// ServerRepository stores servers in PostgreSQL via pgxpool.
type ServerRepository struct {
	pool *pgxpool.Pool
}

// NewServerRepository constructs a PostgreSQL-backed ServerRepository.
func NewServerRepository(pool *pgxpool.Pool) *ServerRepository {
	return &ServerRepository{pool: pool}
}

const serverColumns = `id, name, hostname, ip, operating_system, status, created_at, updated_at`

func (r *ServerRepository) Create(ctx context.Context, server *inventory.Server) error {
	if server == nil {
		return fmt.Errorf("postgres create server: %w: server is nil", inventory.ErrInvalidServer)
	}

	if server.ID == "" {
		id, err := newID()
		if err != nil {
			return fmt.Errorf("postgres create server: generate id: %w", err)
		}
		server.ID = id
	}

	now := time.Now().UTC()
	if server.CreatedAt.IsZero() {
		server.CreatedAt = now
	}
	server.UpdatedAt = now

	const query = `
INSERT INTO servers (id, name, hostname, ip, operating_system, status, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`

	_, err := r.pool.Exec(ctx, query,
		server.ID,
		server.Name,
		server.Hostname,
		server.IP,
		server.OperatingSystem,
		string(server.Status),
		server.CreatedAt,
		server.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres create server: %w", err)
	}
	return nil
}

func (r *ServerRepository) GetByID(ctx context.Context, id string) (*inventory.Server, error) {
	const query = `
SELECT id, name, hostname, ip, operating_system, status, created_at, updated_at
FROM servers
WHERE id = $1`

	server, err := scanServer(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres get server %s: %w", id, inventory.ErrNotFound)
		}
		return nil, fmt.Errorf("postgres get server %s: %w", id, err)
	}
	return server, nil
}

func (r *ServerRepository) List(ctx context.Context) ([]inventory.Server, error) {
	const query = `
SELECT id, name, hostname, ip, operating_system, status, created_at, updated_at
FROM servers
ORDER BY created_at ASC, id ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres list servers: %w", err)
	}
	defer rows.Close()

	servers := make([]inventory.Server, 0)
	for rows.Next() {
		server, err := scanServer(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres list servers: scan: %w", err)
		}
		servers = append(servers, *server)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres list servers: iterate: %w", err)
	}
	return servers, nil
}

func (r *ServerRepository) Update(ctx context.Context, server *inventory.Server) error {
	if server == nil {
		return fmt.Errorf("postgres update server: %w: server is nil", inventory.ErrInvalidServer)
	}

	server.UpdatedAt = time.Now().UTC()

	const query = `
UPDATE servers
SET name = $2,
    hostname = $3,
    ip = $4,
    operating_system = $5,
    status = $6,
    updated_at = $7
WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query,
		server.ID,
		server.Name,
		server.Hostname,
		server.IP,
		server.OperatingSystem,
		string(server.Status),
		server.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres update server %s: %w", server.ID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("postgres update server %s: %w", server.ID, inventory.ErrNotFound)
	}
	return nil
}

func (r *ServerRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM servers WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres delete server %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("postgres delete server %s: %w", id, inventory.ErrNotFound)
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanServer(row scannable) (*inventory.Server, error) {
	var (
		server inventory.Server
		status string
	)
	if err := row.Scan(
		&server.ID,
		&server.Name,
		&server.Hostname,
		&server.IP,
		&server.OperatingSystem,
		&status,
		&server.CreatedAt,
		&server.UpdatedAt,
	); err != nil {
		return nil, err
	}
	server.Status = inventory.ServerStatus(status)
	return &server, nil
}

func newID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
