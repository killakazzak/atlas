package auth

import (
	"context"
	"errors"
	"fmt"
)

// ErrNotFound is returned when a user does not exist.
var ErrNotFound = errors.New("user not found")

// ErrInvalidUser is returned when required user fields are missing or invalid.
var ErrInvalidUser = errors.New("invalid user")

// ErrUsernameConflict is returned when a username is already taken.
var ErrUsernameConflict = errors.New("username already exists")

// Service exposes auth use cases for managed users.
type Service interface {
	CreateUser(ctx context.Context, user *User) error
	GetUser(ctx context.Context, id string) (*User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
}

type service struct {
	users UserRepository
}

// NewService constructs an auth Service backed by the given repository.
func NewService(users UserRepository) Service {
	return &service{users: users}
}

func (s *service) CreateUser(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("%w: user is nil", ErrInvalidUser)
	}
	if err := validateUser(user); err != nil {
		return err
	}
	existing, err := s.users.GetByUsername(ctx, user.Username)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("create user: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("create user: %w", ErrUsernameConflict)
	}
	if err := s.users.Create(ctx, user); err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *service) GetUser(ctx context.Context, id string) (*User, error) {
	user, err := s.users.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *service) ListUsers(ctx context.Context) ([]User, error) {
	users, err := s.users.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	return users, nil
}

func (s *service) UpdateUser(ctx context.Context, user *User) error {
	if user == nil {
		return fmt.Errorf("%w: user is nil", ErrInvalidUser)
	}
	if err := validateUser(user); err != nil {
		return err
	}
	if err := s.users.Update(ctx, user); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *service) DeleteUser(ctx context.Context, id string) error {
	if err := s.users.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func validateUser(user *User) error {
	if user.Username == "" {
		return fmt.Errorf("%w: username is required", ErrInvalidUser)
	}
	if len(user.Username) > 255 {
		return fmt.Errorf("%w: username must not exceed 255 characters", ErrInvalidUser)
	}
	if user.Email == "" {
		return fmt.Errorf("%w: email is required", ErrInvalidUser)
	}
	if user.Role == "" {
		return fmt.Errorf("%w: role is required", ErrInvalidUser)
	}
	return nil
}
