package auth

import (
	"context"
	"errors"
	"testing"
)

type fakeUserRepository struct {
	users         map[string]*User
	byUsername    map[string]*User
	getByIDErr    error
	getByUsernameErr error
	createErr     error
	updateErr     error
	deleteErr     error
}

func newFakeRepo() *fakeUserRepository {
	return &fakeUserRepository{
		users:      make(map[string]*User),
		byUsername: make(map[string]*User),
	}
}

func (f *fakeUserRepository) GetByID(_ context.Context, id string) (*User, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	u, ok := f.users[id]
	if !ok {
		return nil, ErrNotFound
	}
	clone := *u
	return &clone, nil
}

func (f *fakeUserRepository) GetByUsername(_ context.Context, username string) (*User, error) {
	if f.getByUsernameErr != nil {
		return nil, f.getByUsernameErr
	}
	u, ok := f.byUsername[username]
	if !ok {
		return nil, ErrNotFound
	}
	clone := *u
	return &clone, nil
}

func (f *fakeUserRepository) List(_ context.Context) ([]User, error) {
	out := make([]User, 0, len(f.users))
	for _, u := range f.users {
		out = append(out, *u)
	}
	return out, nil
}

func (f *fakeUserRepository) Create(_ context.Context, user *User) error {
	if f.createErr != nil {
		return f.createErr
	}
	if user.ID == "" {
		user.ID = "test-id-" + user.Username
	}
	clone := *user
	f.users[user.ID] = &clone
	f.byUsername[user.Username] = &clone
	return nil
}

func (f *fakeUserRepository) Update(_ context.Context, user *User) error {
	if f.updateErr != nil {
		return f.updateErr
	}
	if _, ok := f.users[user.ID]; !ok {
		return ErrNotFound
	}
	clone := *user
	f.users[user.ID] = &clone
	f.byUsername[user.Username] = &clone
	return nil
}

func (f *fakeUserRepository) Delete(_ context.Context, id string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	u, ok := f.users[id]
	if !ok {
		return ErrNotFound
	}
	delete(f.byUsername, u.Username)
	delete(f.users, id)
	return nil
}

func validUser() *User {
	return &User{
		Username: "alice",
		Email:    "alice@example.com",
		Role:     RoleOperator,
	}
}

func TestService_CreateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		user    *User
		wantErr error
	}{
		{
			name:    "nil user",
			user:    nil,
			wantErr: ErrInvalidUser,
		},
		{
			name:    "missing username",
			user:    &User{Email: "a@b.com", Role: RoleViewer},
			wantErr: ErrInvalidUser,
		},
		{
			name:    "missing email",
			user:    &User{Username: "bob", Role: RoleViewer},
			wantErr: ErrInvalidUser,
		},
		{
			name:    "missing role",
			user:    &User{Username: "bob", Email: "bob@example.com"},
			wantErr: ErrInvalidUser,
		},
		{
			name:    "valid user",
			user:    validUser(),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := NewService(newFakeRepo())
			err := svc.CreateUser(context.Background(), tt.user)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("CreateUser() error = %v, want nil", err)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("CreateUser() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_CreateUser_UsernameConflict(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	svc := NewService(repo)

	u := validUser()
	if err := svc.CreateUser(context.Background(), u); err != nil {
		t.Fatalf("first CreateUser() error = %v", err)
	}

	duplicate := &User{Username: "alice", Email: "other@example.com", Role: RoleViewer}
	err := svc.CreateUser(context.Background(), duplicate)
	if !errors.Is(err, ErrUsernameConflict) {
		t.Fatalf("expected ErrUsernameConflict, got %v", err)
	}
}

func TestService_GetUser(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	svc := NewService(repo)

	u := validUser()
	if err := svc.CreateUser(context.Background(), u); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	got, err := svc.GetUser(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if got.Username != u.Username {
		t.Fatalf("got username %q, want %q", got.Username, u.Username)
	}
}

func TestService_GetUser_NotFound(t *testing.T) {
	t.Parallel()

	svc := NewService(newFakeRepo())
	_, err := svc.GetUser(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_ListUsers(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	svc := NewService(repo)

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected empty list, got %d", len(users))
	}

	if err := svc.CreateUser(context.Background(), validUser()); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	users, err = svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected 1 user, got %d", len(users))
	}
}

func TestService_UpdateUser(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	svc := NewService(repo)

	u := validUser()
	if err := svc.CreateUser(context.Background(), u); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	u.Role = RoleAdministrator
	if err := svc.UpdateUser(context.Background(), u); err != nil {
		t.Fatalf("UpdateUser() error = %v", err)
	}

	got, err := svc.GetUser(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if got.Role != RoleAdministrator {
		t.Fatalf("expected role %q, got %q", RoleAdministrator, got.Role)
	}
}

func TestService_DeleteUser(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	svc := NewService(repo)

	u := validUser()
	if err := svc.CreateUser(context.Background(), u); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if err := svc.DeleteUser(context.Background(), u.ID); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	_, err := svc.GetUser(context.Background(), u.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}
