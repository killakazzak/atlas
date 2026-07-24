package postgres

import (
	"errors"
	"testing"
	"time"

	"atlas/internal/inventory"
)

func TestNewServerRepository(t *testing.T) {
	t.Parallel()

	repo := NewServerRepository(nil)
	if repo == nil {
		t.Fatal("NewServerRepository() returned nil")
	}
	if repo.pool != nil {
		t.Fatal("expected nil pool to be stored as-is")
	}
}

func TestScanServer(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, 7, 24, 12, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 7, 24, 13, 0, 0, 0, time.UTC)

	row := &fakeRow{
		values: []any{
			"abc123",
			"srv-1",
			"srv-1.local",
			"10.0.0.1",
			"linux",
			"online",
			createdAt,
			updatedAt,
		},
	}

	server, err := scanServer(row)
	if err != nil {
		t.Fatalf("scanServer() error = %v", err)
	}
	if server.ID != "abc123" ||
		server.Name != "srv-1" ||
		server.Hostname != "srv-1.local" ||
		server.IP != "10.0.0.1" ||
		server.OperatingSystem != "linux" ||
		server.Status != inventory.ServerStatusOnline ||
		!server.CreatedAt.Equal(createdAt) ||
		!server.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("scanServer() unexpected server: %+v", server)
	}
}

func TestScanServer_PropagatesError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("scan failed")
	_, err := scanServer(&fakeRow{err: wantErr})
	if !errors.Is(err, wantErr) {
		t.Fatalf("scanServer() error = %v, want %v", err, wantErr)
	}
}

func TestNewID(t *testing.T) {
	t.Parallel()

	id, err := newID()
	if err != nil {
		t.Fatalf("newID() error = %v", err)
	}
	if len(id) != 32 {
		t.Fatalf("newID() len = %d, want 32", len(id))
	}
}

type fakeRow struct {
	values []any
	err    error
}

func (r *fakeRow) Scan(dest ...any) error {
	if r.err != nil {
		return r.err
	}
	if len(dest) != len(r.values) {
		return errors.New("destination count mismatch")
	}
	for i := range dest {
		switch d := dest[i].(type) {
		case *string:
			v, ok := r.values[i].(string)
			if !ok {
				return errors.New("expected string")
			}
			*d = v
		case *time.Time:
			v, ok := r.values[i].(time.Time)
			if !ok {
				return errors.New("expected time.Time")
			}
			*d = v
		default:
			return errors.New("unsupported destination type")
		}
	}
	return nil
}
