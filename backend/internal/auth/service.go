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

// ErrInvalidCredentials is returned when a login/password pair is incorrect.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Service exposes auth use cases for managed users.
type Service interface {
	// CreateUser validates, hashes the password, and persists a new user.
	// The caller provides plaintext password; PasswordHash on the input is ignored.
	CreateUser(ctx context.Context, login, email, password string, role Role) (*User, error)
	GetUser(ctx context.Context, id string) (*User, error)
	ListUsers(ctx context.Context) ([]User, error)
	UpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, id string) error
	// Authenticate verifies credentials and returns the matching user.
	Authenticate(ctx context.Context, login, password string) (*User, error)
}

type service struct {
	users  UserRepository
	hasher PasswordHasher
}

// NewService constructs an auth Service backed by the given repository and hasher.
func NewService(users UserRepository, hasher PasswordHasher) Service {
	return &service{users: users, hasher: hasher}
}

func (s *service) CreateUser(ctx context.Context, login, email, password string, role Role) (*User, error) {
	user := &User{
		Username: login,
		Email:    email,
		Role:     role,
	}

	if err := validateUser(user); err != nil {
		return nil, err
	}
	if password == "" {
		return nil, fmt.Errorf("%w: password is required", ErrInvalidUser)
	}

	existing, err := s.users.GetByUsername(ctx, login)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("create user: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("create user: %w", ErrUsernameConflict)
	}

	hash, err := s.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("create user: hash password: %w", err)
	}
	user.PasswordHash = hash

	if err := s.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return user, nil
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

func (s *service) Authenticate(ctx context.Context, login, password string) (*User, error) {
	user, err := s.users.GetByUsername(ctx, login)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("authenticate: %w", err)
	}
	if err := s.hasher.Compare(user.PasswordHash, password); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
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
