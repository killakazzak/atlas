package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	atlashttp "atlas/internal/http"
	"atlas/internal/config"
	"atlas/internal/inventory"
	"atlas/internal/logger"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()
	repo := inventory.NewMemoryServerRepository()
	svc := inventory.NewService(repo)
	srv := atlashttp.New(config.Config{Port: "0"}, logger.New(), svc)
	return srv.Handler()
}

func TestHealth_RequestID(t *testing.T) {
	h := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header on /health")
	}
}

func TestServers_RequestID(t *testing.T) {
	h := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header on /api/v1/servers")
	}
}

func TestHealth_ResponseBody(t *testing.T) {
	h := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	h.ServeHTTP(w, r)

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("unexpected body: %v", body)
	}
}

func TestServers_EmptyList(t *testing.T) {
	h := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil)
	h.ServeHTTP(w, r)

	var body []any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if len(body) != 0 {
		t.Fatalf("expected empty list, got %v", body)
	}
}
