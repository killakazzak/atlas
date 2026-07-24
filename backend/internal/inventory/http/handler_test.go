package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"atlas/internal/auth"
	"atlas/internal/inventory"
	inventoryhttp "atlas/internal/inventory/http"
)

func nopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// testTokenSvc returns a JWT service and a pre-signed token for the given role.
func testTokenSvc(t *testing.T, role auth.Role) (auth.TokenService, string) {
	t.Helper()
	svc := auth.NewJWTService(auth.JWTConfig{
		Secret: []byte("test-secret"),
		Issuer: "test",
		TTL:    time.Hour,
	})
	tok, err := svc.Generate(&auth.User{ID: "u-1", Username: "alice", Role: role})
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return svc, tok
}

func newTestMux(t *testing.T, role auth.Role) (*http.ServeMux, string) {
	t.Helper()
	tokenSvc, tok := testTokenSvc(t, role)
	repo := inventory.NewMemoryServerRepository()
	svc := inventory.NewService(repo)
	mux := http.NewServeMux()
	inventoryhttp.NewHandler(svc, tokenSvc, nopLogger()).Register(mux)
	return mux, tok
}

func bearer(r *http.Request, tok string) *http.Request {
	r.Header.Set("Authorization", "Bearer "+tok)
	return r
}

func errorCode(t *testing.T, body []byte) string {
	t.Helper()
	var resp struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	return resp.Error.Code
}

func TestHandler_ServersAPI(t *testing.T) {
	t.Parallel()
	mux, adminTok := newTestMux(t, auth.RoleAdministrator)

	// Empty list
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil), adminTok))
	if rec.Code != http.StatusOK {
		t.Fatalf("list status = %d, want %d", rec.Code, http.StatusOK)
	}
	var list []map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("list decode: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("list len = %d, want 0", len(list))
	}

	// Invalid create — missing name
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(
		http.MethodPost, "/api/v1/servers",
		bytes.NewBufferString(`{"name":""}`),
	), adminTok))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create invalid status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if code := errorCode(t, rec.Body.Bytes()); code != "validation_error" {
		t.Fatalf("expected validation_error, got %q", code)
	}

	// Create
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(
		http.MethodPost, "/api/v1/servers",
		bytes.NewBufferString(`{"name":"srv-1","hostname":"srv-1.local","ip":"10.0.0.1"}`),
	), adminTok))
	if rec.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	var created map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("create decode: %v", err)
	}
	id, _ := created["id"].(string)
	if id == "" {
		t.Fatal("create: missing id")
	}

	// Get
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+id, nil), adminTok))
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Get missing → 404 not_found
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(http.MethodGet, "/api/v1/servers/missing", nil), adminTok))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if code := errorCode(t, rec.Body.Bytes()); code != "not_found" {
		t.Fatalf("expected not_found, got %q", code)
	}

	// Update
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(
		http.MethodPut, "/api/v1/servers/"+id,
		bytes.NewBufferString(`{"name":"srv-1-renamed","hostname":"srv-1.local","ip":"10.0.0.2","status":"online"}`),
	), adminTok))
	if rec.Code != http.StatusOK {
		t.Fatalf("update status = %d, want %d body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	var updated map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &updated); err != nil {
		t.Fatalf("update decode: %v", err)
	}
	if updated["name"] != "srv-1-renamed" || updated["ip"] != "10.0.0.2" || updated["status"] != "online" {
		t.Fatalf("update fields not applied: %v", updated)
	}
	if updated["id"] != id {
		t.Fatalf("update changed id: got %v, want %s", updated["id"], id)
	}

	// Update missing → 404 not_found
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(
		http.MethodPut, "/api/v1/servers/missing",
		bytes.NewBufferString(`{"name":"x","hostname":"x.local"}`),
	), adminTok))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("update missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if code := errorCode(t, rec.Body.Bytes()); code != "not_found" {
		t.Fatalf("expected not_found, got %q", code)
	}

	// Delete
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(http.MethodDelete, "/api/v1/servers/"+id, nil), adminTok))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	// Delete missing → 404 not_found
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(http.MethodDelete, "/api/v1/servers/"+id, nil), adminTok))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if code := errorCode(t, rec.Body.Bytes()); code != "not_found" {
		t.Fatalf("expected not_found, got %q", code)
	}
}

func TestHandler_RoleEnforcement(t *testing.T) {
	t.Parallel()

	// Viewer can read but not write or delete.
	mux, viewerTok := newTestMux(t, auth.RoleViewer)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil), viewerTok))
	if rec.Code != http.StatusOK {
		t.Fatalf("viewer list: expected 200, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(
		http.MethodPost, "/api/v1/servers",
		bytes.NewBufferString(`{"name":"x","hostname":"x.local"}`),
	), viewerTok))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("viewer create: expected 403, got %d", rec.Code)
	}

	// Operator can write but not delete.
	mux, operatorTok := newTestMux(t, auth.RoleOperator)

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(
		http.MethodPost, "/api/v1/servers",
		bytes.NewBufferString(`{"name":"srv","hostname":"srv.local"}`),
	), operatorTok))
	if rec.Code != http.StatusCreated {
		t.Fatalf("operator create: expected 201, got %d: %s", rec.Code, rec.Body)
	}
	var created map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &created)
	id := created["id"].(string)

	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, bearer(httptest.NewRequest(http.MethodDelete, "/api/v1/servers/"+id, nil), operatorTok))
	if rec.Code != http.StatusForbidden {
		t.Fatalf("operator delete: expected 403, got %d", rec.Code)
	}

	// No token → 401.
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("no token: expected 401, got %d", rec.Code)
	}
}
