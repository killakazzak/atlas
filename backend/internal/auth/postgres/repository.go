// Package postgres implements auth persistence with PostgreSQL.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"atlas/internal/auth"
)

// Ensure UserRepository satisfies auth.UserRepository at compile time.
var _ auth.UserRepository = (*UserRepository)(nil)

// UserRepository stores users in PostgreSQL via pgxpool.
type UserRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository constructs a PostgreSQL-backed UserRepository.
func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

// Create inserts a new user into the database.
func (r *UserRepository) Create(ctx context.Context, user *auth.User) error {
	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	const query = `
INSERT INTO users (id, login, email, password_hash, role, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		string(user.Role),
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres create user: %w", err)
	}
	return nil
}

// GetByID returns the user with the given id, or auth.ErrNotFound.
func (r *UserRepository) GetByID(ctx context.Context, id string) (*auth.User, error) {
	const query = `
SELECT id, login, email, password_hash, role, created_at, updated_at
FROM users
WHERE id = $1`

	user, err := scanUser(r.pool.QueryRow(ctx, query, id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres get user %s: %w", id, auth.ErrNotFound)
		}
		return nil, fmt.Errorf("postgres get user %s: %w", id, err)
	}
	return user, nil
}

// GetByUsername returns the user with the given username (login column), or auth.ErrNotFound.
func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	const query = `
SELECT id, login, email, password_hash, role, created_at, updated_at
FROM users
WHERE login = $1`

	user, err := scanUser(r.pool.QueryRow(ctx, query, username))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("postgres get user by username %s: %w", username, auth.ErrNotFound)
		}
		return nil, fmt.Errorf("postgres get user by username %s: %w", username, err)
	}
	return user, nil
}

// List returns all users ordered by creation time.
func (r *UserRepository) List(ctx context.Context) ([]auth.User, error) {
	const query = `
SELECT id, login, email, password_hash, role, created_at, updated_at
FROM users
ORDER BY created_at ASC, id ASC`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("postgres list users: %w", err)
	}
	defer rows.Close()

	users := make([]auth.User, 0)
	for rows.Next() {
		user, err := scanUser(rows)
		if err != nil {
			return nil, fmt.Errorf("postgres list users: scan: %w", err)
		}
		users = append(users, *user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres list users: iterate: %w", err)
	}
	return users, nil
}

// Update saves changes to an existing user.
func (r *UserRepository) Update(ctx context.Context, user *auth.User) error {
	user.UpdatedAt = time.Now().UTC()

	const query = `
UPDATE users
SET login         = $2,
    email         = $3,
    password_hash = $4,
    role          = $5,
    updated_at    = $6
WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query,
		user.ID,
		user.Username,
		user.Email,
		user.PasswordHash,
		string(user.Role),
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("postgres update user %s: %w", user.ID, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("postgres update user %s: %w", user.ID, auth.ErrNotFound)
	}
	return nil
}

// Delete removes the user with the given id.
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	const query = `DELETE FROM users WHERE id = $1`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("postgres delete user %s: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("postgres delete user %s: %w", id, auth.ErrNotFound)
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanUser(row scannable) (*auth.User, error) {
	var (
		user auth.User
		role string
	)
	if err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&role,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return nil, err
	}
	user.Role = auth.Role(role)
	return &user, nil
}
