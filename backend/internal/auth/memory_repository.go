package auth

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Ensure MemoryUserRepository satisfies UserRepository at compile time.
var _ UserRepository = (*MemoryUserRepository)(nil)

// MemoryUserRepository is a thread-safe in-memory UserRepository for testing and local dev.
type MemoryUserRepository struct {
	mu         sync.RWMutex
	users      map[string]*User
	byUsername map[string]*User
}

// NewMemoryUserRepository constructs an empty MemoryUserRepository.
func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{
		users:      make(map[string]*User),
		byUsername: make(map[string]*User),
	}
}

// Create inserts a new user.
func (r *MemoryUserRepository) Create(_ context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == "" {
		user.ID = uuid.NewString()
	}

	now := time.Now().UTC()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	clone := *user
	r.users[user.ID] = &clone
	r.byUsername[user.Username] = &clone
	return nil
}

// GetByID returns the user with the given id.
func (r *MemoryUserRepository) GetByID(_ context.Context, id string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.users[id]
	if !ok {
		return nil, fmt.Errorf("memory get user %s: %w", id, ErrNotFound)
	}
	clone := *u
	return &clone, nil
}

// GetByUsername returns the user with the given username.
func (r *MemoryUserRepository) GetByUsername(_ context.Context, username string) (*User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.byUsername[username]
	if !ok {
		return nil, fmt.Errorf("memory get user by username %s: %w", username, ErrNotFound)
	}
	clone := *u
	return &clone, nil
}

// List returns all users.
func (r *MemoryUserRepository) List(_ context.Context) ([]User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]User, 0, len(r.users))
	for _, u := range r.users {
		out = append(out, *u)
	}
	return out, nil
}

// Update saves changes to an existing user.
func (r *MemoryUserRepository) Update(_ context.Context, user *User) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.users[user.ID]
	if !ok {
		return fmt.Errorf("memory update user %s: %w", user.ID, ErrNotFound)
	}

	delete(r.byUsername, existing.Username)

	user.UpdatedAt = time.Now().UTC()
	clone := *user
	r.users[user.ID] = &clone
	r.byUsername[user.Username] = &clone
	return nil
}

// Delete removes the user with the given id.
func (r *MemoryUserRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	u, ok := r.users[id]
	if !ok {
		return fmt.Errorf("memory delete user %s: %w", id, ErrNotFound)
	}

	delete(r.byUsername, u.Username)
	delete(r.users, id)
	return nil
}
