package inventory

import (
	"context"
	"errors"
	"testing"
)

func TestMemoryServerRepository_CRUD(t *testing.T) {
	ctx := context.Background()
	repo := NewMemoryServerRepository()

	server := &Server{
		Name:     "srv-1",
		Hostname: "srv-1.local",
		Status:   ServerStatusOnline,
	}
	if err := repo.Create(ctx, server); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if server.ID == "" {
		t.Fatal("Create: expected generated ID")
	}

	got, err := repo.GetByID(ctx, server.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != server.Name || got.Hostname != server.Hostname {
		t.Fatalf("GetByID: unexpected server %+v", got)
	}

	got.Name = "srv-1-renamed"
	if err := repo.Update(ctx, got); err != nil {
		t.Fatalf("Update: %v", err)
	}

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].Name != "srv-1-renamed" {
		t.Fatalf("List: unexpected result %+v", list)
	}

	if err := repo.Delete(ctx, server.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := repo.GetByID(ctx, server.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID after Delete: want ErrNotFound, got %v", err)
	}
}
