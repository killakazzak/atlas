package auth

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// fakeHasher is a PasswordHasher stub that prepends "hashed:" to make
// hashing deterministic and cheap in tests.
type fakeHasher struct{}

func (fakeHasher) Hash(password string) (string, error) { return "hashed:" + password, nil }
func (fakeHasher) Compare(hash, password string) error {
	if hash != "hashed:"+password {
		return ErrInvalidCredentials
	}
	return nil
}

type fakeUserRepository struct {
	users            map[string]*User
	byUsername       map[string]*User
	getByIDErr       error
	getByUsernameErr error
	createErr        error
	updateErr        error
	deleteErr        error
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

func newSvc() (Service, *fakeUserRepository) {
	repo := newFakeRepo()
	return NewService(repo, fakeHasher{}), repo
}

func TestService_CreateUser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		login    string
		email    string
		password string
		role     Role
		wantErr  error
	}{
		{name: "missing login", login: "", email: "a@b.com", password: "pw", role: RoleViewer, wantErr: ErrInvalidUser},
		{name: "missing email", login: "bob", email: "", password: "pw", role: RoleViewer, wantErr: ErrInvalidUser},
		{name: "missing role", login: "bob", email: "b@b.com", password: "pw", role: "", wantErr: ErrInvalidUser},
		{name: "missing password", login: "bob", email: "b@b.com", password: "", role: RoleViewer, wantErr: ErrInvalidUser},
		{name: "valid", login: "alice", email: "alice@example.com", password: "s3cr3t", role: RoleOperator, wantErr: nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := newSvc()
			_, err := svc.CreateUser(context.Background(), tt.login, tt.email, tt.password, tt.role)
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

func TestService_CreateUser_HashesPassword(t *testing.T) {
	t.Parallel()

	svc, repo := newSvc()
	u, err := svc.CreateUser(context.Background(), "alice", "alice@example.com", "s3cr3t", RoleOperator)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	stored := repo.users[u.ID]
	if stored.PasswordHash == "s3cr3t" {
		t.Fatal("PasswordHash must not equal plaintext password")
	}
	if !strings.HasPrefix(stored.PasswordHash, "hashed:") {
		t.Fatalf("unexpected PasswordHash %q", stored.PasswordHash)
	}
}

func TestService_CreateUser_UsernameConflict(t *testing.T) {
	t.Parallel()

	svc, _ := newSvc()
	if _, err := svc.CreateUser(context.Background(), "alice", "alice@example.com", "pw", RoleViewer); err != nil {
		t.Fatalf("first CreateUser() error = %v", err)
	}

	_, err := svc.CreateUser(context.Background(), "alice", "other@example.com", "pw", RoleViewer)
	if !errors.Is(err, ErrUsernameConflict) {
		t.Fatalf("expected ErrUsernameConflict, got %v", err)
	}
}

func TestService_GetUser(t *testing.T) {
	t.Parallel()

	svc, _ := newSvc()
	u, err := svc.CreateUser(context.Background(), "alice", "alice@example.com", "pw", RoleOperator)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	got, err := svc.GetUser(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("GetUser() error = %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("got username %q, want %q", got.Username, "alice")
	}
}

func TestService_GetUser_NotFound(t *testing.T) {
	t.Parallel()

	svc, _ := newSvc()
	_, err := svc.GetUser(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestService_ListUsers(t *testing.T) {
	t.Parallel()

	svc, _ := newSvc()

	users, err := svc.ListUsers(context.Background())
	if err != nil {
		t.Fatalf("ListUsers() error = %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected empty list, got %d", len(users))
	}

	if _, err := svc.CreateUser(context.Background(), "alice", "alice@example.com", "pw", RoleOperator); err != nil {
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

	svc, _ := newSvc()
	u, err := svc.CreateUser(context.Background(), "alice", "alice@example.com", "pw", RoleOperator)
	if err != nil {
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

	svc, _ := newSvc()
	u, err := svc.CreateUser(context.Background(), "alice", "alice@example.com", "pw", RoleOperator)
	if err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	if err := svc.DeleteUser(context.Background(), u.ID); err != nil {
		t.Fatalf("DeleteUser() error = %v", err)
	}

	_, err = svc.GetUser(context.Background(), u.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestService_Authenticate_Success(t *testing.T) {
	t.Parallel()

	svc, _ := newSvc()
	if _, err := svc.CreateUser(context.Background(), "alice", "alice@example.com", "s3cr3t", RoleOperator); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	got, err := svc.Authenticate(context.Background(), "alice", "s3cr3t")
	if err != nil {
		t.Fatalf("Authenticate() error = %v", err)
	}
	if got.Username != "alice" {
		t.Fatalf("got username %q, want %q", got.Username, "alice")
	}
}

func TestService_Authenticate_WrongPassword(t *testing.T) {
	t.Parallel()

	svc, _ := newSvc()
	if _, err := svc.CreateUser(context.Background(), "alice", "alice@example.com", "s3cr3t", RoleOperator); err != nil {
		t.Fatalf("CreateUser() error = %v", err)
	}

	_, err := svc.Authenticate(context.Background(), "alice", "wrong")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestService_Authenticate_UnknownLogin(t *testing.T) {
	t.Parallel()

	svc, _ := newSvc()
	_, err := svc.Authenticate(context.Background(), "ghost", "pw")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}
