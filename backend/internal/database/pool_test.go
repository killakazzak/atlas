package database

import (
	"context"
	"strings"
	"testing"
)

func TestNewPool_EmptyURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		url  string
	}{
		{name: "empty", url: ""},
		{name: "whitespace", url: "   "},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool, err := NewPool(context.Background(), tt.url)
			if err == nil {
				t.Fatal("NewPool() error = nil, want error")
			}
			if pool != nil {
				t.Fatal("NewPool() pool != nil, want nil")
			}
			if !strings.Contains(err.Error(), "database URL is required") {
				t.Fatalf("NewPool() error = %v, want mention of required URL", err)
			}
		})
	}
}
