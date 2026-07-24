package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// noRequestID is a no-op getter used in unit tests that run outside the full
// middleware chain and therefore have no request ID in context.
func noRequestID(context.Context) string { return "" }

func testTokenSvc() *JWTService {
	return NewJWTService(JWTConfig{
		Secret: []byte("test-secret"),
		Issuer: "test",
		TTL:    time.Hour,
	})
}

func testTokenFor(t *testing.T, svc *JWTService, userID, login string, role Role) string {
	t.Helper()
	tok, err := svc.Generate(&User{ID: userID, Username: login, Role: role})
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return tok
}

func applyMiddleware(tokenSvc TokenService, next http.Handler) http.Handler {
	return RequireAuth(tokenSvc, noRequestID)(next)
}

func TestRequireAuth_MissingHeader(t *testing.T) {
	t.Parallel()

	svc := testTokenSvc()
	h := applyMiddleware(svc, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_WrongScheme(t *testing.T) {
	t.Parallel()

	svc := testTokenSvc()
	h := applyMiddleware(svc, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_InvalidToken(t *testing.T) {
	t.Parallel()

	svc := testTokenSvc()
	h := applyMiddleware(svc, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer not.a.valid.jwt")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestRequireAuth_ValidToken_PassesClaimsToContext(t *testing.T) {
	t.Parallel()

	svc := testTokenSvc()
	tok := testTokenFor(t, svc, "u-1", "alice", RoleOperator)

	var got *Claims
	h := applyMiddleware(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if got == nil {
		t.Fatal("claims not stored in context")
	}
	if got.UserID != "u-1" {
		t.Errorf("UserID = %q, want %q", got.UserID, "u-1")
	}
	if got.Login != "alice" {
		t.Errorf("Login = %q, want %q", got.Login, "alice")
	}
	if got.Role != RoleOperator {
		t.Errorf("Role = %q, want %q", got.Role, RoleOperator)
	}
}

func TestRequireAuth_ExpiredToken(t *testing.T) {
	t.Parallel()

	expired := NewJWTService(JWTConfig{
		Secret: []byte("test-secret"),
		Issuer: "test",
		TTL:    -time.Second,
	})
	tok := testTokenFor(t, expired, "u-1", "alice", RoleViewer)

	// Parse with a normal (non-expired) service — token is still expired.
	svc := testTokenSvc()
	h := applyMiddleware(svc, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", w.Code)
	}
}
