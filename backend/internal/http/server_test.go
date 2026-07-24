package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlas/internal/config"
	atlashttp "atlas/internal/http"
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

func TestOpenAPISpec_Returns200(t *testing.T) {
	h := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/openapi/openapi.yaml", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Fatalf("expected Content-Type application/yaml, got %q", ct)
	}
	body := w.Body.String()
	if !strings.Contains(body, "openapi:") {
		t.Fatal("response does not look like an OpenAPI spec")
	}
}

func TestSwaggerUI_Returns200(t *testing.T) {
	h := newTestServer(t)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/swagger", nil)
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
		t.Fatalf("expected text/html Content-Type, got %q", ct)
	}
	if !strings.Contains(w.Body.String(), "swagger-ui") {
		t.Fatal("response does not contain swagger-ui")
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
