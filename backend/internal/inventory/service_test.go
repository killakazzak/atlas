package inventory

import (
	"context"
	"errors"
	"testing"
)

type fakeServerRepository struct {
	createCalls int
	lastCreated *Server
	createErr   error
}

func (f *fakeServerRepository) GetByID(context.Context, string) (*Server, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeServerRepository) List(context.Context) ([]Server, error) {
	return nil, errors.New("not implemented")
}

func (f *fakeServerRepository) Create(_ context.Context, server *Server) error {
	f.createCalls++
	if server != nil {
		clone := *server
		f.lastCreated = &clone
	}
	return f.createErr
}

func (f *fakeServerRepository) Update(context.Context, *Server) error {
	return errors.New("not implemented")
}

func (f *fakeServerRepository) Delete(context.Context, string) error {
	return errors.New("not implemented")
}

func TestService_RegisterServer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		server      *Server
		wantErr     error
		wantCreates int
	}{
		{
			name:        "empty name",
			server:      &Server{Hostname: "host-1.local"},
			wantErr:     ErrInvalidServer,
			wantCreates: 0,
		},
		{
			name:        "empty hostname",
			server:      &Server{Name: "srv-1"},
			wantErr:     ErrInvalidServer,
			wantCreates: 0,
		},
		{
			name: "valid server",
			server: &Server{
				Name:     "srv-1",
				Hostname: "host-1.local",
			},
			wantErr:     nil,
			wantCreates: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeServerRepository{}
			svc := NewService(repo)

			err := svc.RegisterServer(context.Background(), tt.server)
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("RegisterServer() error = %v, want nil", err)
				}
			} else if !errors.Is(err, tt.wantErr) {
				t.Fatalf("RegisterServer() error = %v, want %v", err, tt.wantErr)
			}

			if repo.createCalls != tt.wantCreates {
				t.Fatalf("Create calls = %d, want %d", repo.createCalls, tt.wantCreates)
			}

			if tt.wantCreates == 1 {
				if repo.lastCreated == nil {
					t.Fatal("Create was not called with a server")
				}
				if repo.lastCreated.Name != tt.server.Name || repo.lastCreated.Hostname != tt.server.Hostname {
					t.Fatalf("Create got %+v, want %+v", repo.lastCreated, tt.server)
				}
			}
		})
	}
}

func TestService_RegisterServer_callsCreateOnce(t *testing.T) {
	t.Parallel()

	repo := &fakeServerRepository{}
	svc := NewService(repo)

	server := &Server{
		Name:     "srv-1",
		Hostname: "host-1.local",
	}
	if err := svc.RegisterServer(context.Background(), server); err != nil {
		t.Fatalf("RegisterServer() error = %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatalf("Create calls = %d, want 1", repo.createCalls)
	}
}
