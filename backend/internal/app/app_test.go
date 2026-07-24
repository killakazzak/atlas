package app_test

import (
	"context"
	"testing"
	"time"

	"atlas/internal/app"
	"atlas/internal/config"
)

func TestNewWiresApp(t *testing.T) {
	application, err := app.New(config.Config{Port: "0"})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// Shutdown before Run is a no-op-friendly call on an unstarted server.
	if err := application.Shutdown(ctx); err != nil {
		t.Fatalf("Shutdown: %v", err)
	}
}
