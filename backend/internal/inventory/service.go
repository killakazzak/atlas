package inventory

import (
	"context"
	"errors"
	"fmt"
)

// ErrInvalidServer is returned when required server fields are missing.
var ErrInvalidServer = errors.New("invalid server")

// Service exposes inventory use cases for managed servers.
type Service interface {
	RegisterServer(ctx context.Context, server *Server) error
	GetServer(ctx context.Context, id string) (*Server, error)
	ListServers(ctx context.Context) ([]Server, error)
	UpdateServer(ctx context.Context, server *Server) error
	DeleteServer(ctx context.Context, id string) error
}

type service struct {
	servers ServerRepository
}

// NewService constructs an inventory Service backed by the given repository.
func NewService(serverRepository ServerRepository) Service {
	return &service{servers: serverRepository}
}

func (s *service) RegisterServer(ctx context.Context, server *Server) error {
	if server == nil {
		return fmt.Errorf("%w: server is nil", ErrInvalidServer)
	}
	if err := validateServer(server); err != nil {
		return err
	}
	if err := s.servers.Create(ctx, server); err != nil {
		return fmt.Errorf("register server: %w", err)
	}
	return nil
}

func (s *service) GetServer(ctx context.Context, id string) (*Server, error) {
	server, err := s.servers.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get server: %w", err)
	}
	return server, nil
}

func (s *service) ListServers(ctx context.Context) ([]Server, error) {
	servers, err := s.servers.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list servers: %w", err)
	}
	return servers, nil
}

func (s *service) UpdateServer(ctx context.Context, server *Server) error {
	if err := validateServer(server); err != nil {
		return err
	}
	if err := s.servers.Update(ctx, server); err != nil {
		return fmt.Errorf("update server: %w", err)
	}
	return nil
}

func validateServer(server *Server) error {
	if server.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidServer)
	}
	if len(server.Name) > 255 {
		return fmt.Errorf("%w: name must not exceed 255 characters", ErrInvalidServer)
	}
	if server.Hostname == "" {
		return fmt.Errorf("%w: hostname is required", ErrInvalidServer)
	}
	if len(server.Hostname) > 255 {
		return fmt.Errorf("%w: hostname must not exceed 255 characters", ErrInvalidServer)
	}
	return nil
}

func (s *service) DeleteServer(ctx context.Context, id string) error {
	if err := s.servers.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete server: %w", err)
	}
	return nil
}
