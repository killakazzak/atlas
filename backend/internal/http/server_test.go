package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"atlas/internal/auth"
	atlashttp "atlas/internal/http"
	"atlas/internal/config"
	"atlas/internal/inventory"
	"atlas/internal/logger"
)

func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	repo := inventory.NewMemoryServerRepository()
	svc := inventory.NewService(repo)

	userRepo := auth.NewMemoryUserRepository()
	authSvc := auth.NewService(userRepo, auth.BcryptHasher{})
	tokenSvc := auth.NewJWTService(auth.JWTConfig{
		Secret: []byte("test-secret"),
		Issuer: "test",
		TTL:    time.Hour,
	})

	cfg := config.Config{Port: "0", JWTAccessTokenTTL: time.Hour}
	srv := atlashttp.New(cfg, logger.New(), svc, authSvc, tokenSvc)
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

// --- Auth endpoint tests ---

func postJSON(h http.Handler, path string, body any) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	r := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	return w
}

func createUserAndLogin(t *testing.T, svc auth.Service, h http.Handler) string {
	t.Helper()
	ctx := t.Context()
	_, err := svc.CreateUser(ctx, "alice", "alice@example.com", "s3cr3t", auth.RoleOperator)
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	w := postJSON(h, "/api/v1/auth/login", map[string]string{"login": "alice", "password": "s3cr3t"})
	if w.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", w.Code, w.Body)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("login: invalid JSON: %v", err)
	}
	token, ok := resp["accessToken"].(string)
	if !ok || token == "" {
		t.Fatal("login: missing accessToken in response")
	}
	return token
}

func newTestServerWithAuth(t *testing.T) (http.Handler, auth.Service) {
	t.Helper()
	userRepo := auth.NewMemoryUserRepository()
	authSvc := auth.NewService(userRepo, auth.BcryptHasher{})
	tokenSvc := auth.NewJWTService(auth.JWTConfig{
		Secret: []byte("test-secret"),
		Issuer: "test",
		TTL:    time.Hour,
	})

	repo := inventory.NewMemoryServerRepository()
	invSvc := inventory.NewService(repo)
	cfg := config.Config{Port: "0", JWTAccessTokenTTL: time.Hour}
	srv := atlashttp.New(cfg, logger.New(), invSvc, authSvc, tokenSvc)
	return srv.Handler(), authSvc
}

func TestAuth_Login_Success(t *testing.T) {
	h, svc := newTestServerWithAuth(t)
	createUserAndLogin(t, svc, h) // asserts internally
}

func TestAuth_Login_WrongPassword(t *testing.T) {
	h, svc := newTestServerWithAuth(t)
	ctx := t.Context()
	if _, err := svc.CreateUser(ctx, "bob", "bob@example.com", "correct", auth.RoleViewer); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	w := postJSON(h, "/api/v1/auth/login", map[string]string{"login": "bob", "password": "wrong"})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuth_Login_MissingFields(t *testing.T) {
	h, _ := newTestServerWithAuth(t)
	w := postJSON(h, "/api/v1/auth/login", map[string]string{"login": ""})
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAuth_Logout_Returns204(t *testing.T) {
	h, _ := newTestServerWithAuth(t)
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestAuth_Me_Success(t *testing.T) {
	h, svc := newTestServerWithAuth(t)
	token := createUserAndLogin(t, svc, h)

	r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if body["login"] != "alice" {
		t.Fatalf("expected login=alice, got %v", body["login"])
	}
}

func TestAuth_Me_NoToken(t *testing.T) {
	h, _ := newTestServerWithAuth(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestAuth_Me_InvalidToken(t *testing.T) {
	h, _ := newTestServerWithAuth(t)
	r := httptest.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	r.Header.Set("Authorization", "Bearer garbage.token.here")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}
