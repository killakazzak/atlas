package inventory

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ErrNotFound is returned when an entity does not exist in a repository.
var ErrNotFound = errors.New("not found")

// MemoryServerRepository is a temporary in-memory ServerRepository for local wiring.
type MemoryServerRepository struct {
	mu      sync.RWMutex
	servers map[string]*Server
}

// NewMemoryServerRepository constructs an empty in-memory server store.
func NewMemoryServerRepository() *MemoryServerRepository {
	return &MemoryServerRepository{
		servers: make(map[string]*Server),
	}
}

func (r *MemoryServerRepository) GetByID(_ context.Context, id string) (*Server, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	server, ok := r.servers[id]
	if !ok {
		return nil, fmt.Errorf("server %s: %w", id, ErrNotFound)
	}
	copy := *server
	return &copy, nil
}

func (r *MemoryServerRepository) List(_ context.Context) ([]Server, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]Server, 0, len(r.servers))
	for _, server := range r.servers {
		out = append(out, *server)
	}
	return out, nil
}

func (r *MemoryServerRepository) Create(_ context.Context, server *Server) error {
	if server == nil {
		return fmt.Errorf("%w: server is nil", ErrInvalidServer)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if server.ID == "" {
		server.ID = newID()
	}
	if _, exists := r.servers[server.ID]; exists {
		return fmt.Errorf("server %s: already exists", server.ID)
	}

	now := time.Now().UTC()
	if server.CreatedAt.IsZero() {
		server.CreatedAt = now
	}
	server.UpdatedAt = now

	stored := *server
	r.servers[server.ID] = &stored
	return nil
}

func (r *MemoryServerRepository) Update(_ context.Context, server *Server) error {
	if server == nil {
		return fmt.Errorf("%w: server is nil", ErrInvalidServer)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.servers[server.ID]; !ok {
		return fmt.Errorf("server %s: %w", server.ID, ErrNotFound)
	}

	server.UpdatedAt = time.Now().UTC()
	stored := *server
	r.servers[server.ID] = &stored
	return nil
}

func (r *MemoryServerRepository) Delete(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.servers[id]; !ok {
		return fmt.Errorf("server %s: %w", id, ErrNotFound)
	}
	delete(r.servers, id)
	return nil
}

func newID() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		panic(fmt.Sprintf("inventory: generate id: %v", err))
	}
	return hex.EncodeToString(buf)
}
