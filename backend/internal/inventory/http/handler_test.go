package http_test

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlas/internal/inventory"
	inventoryhttp "atlas/internal/inventory/http"
)

func nopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func newTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	repo := inventory.NewMemoryServerRepository()
	svc := inventory.NewService(repo)
	mux := http.NewServeMux()
	inventoryhttp.NewHandler(svc, nopLogger()).Register(mux)
	return mux
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
	mux := newTestMux(t)

	// Empty list
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/servers", nil))
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
	mux.ServeHTTP(rec, httptest.NewRequest(
		http.MethodPost,
		"/api/v1/servers",
		bytes.NewBufferString(`{"name":""}`),
	))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create invalid status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if code := errorCode(t, rec.Body.Bytes()); code != "validation_error" {
		t.Fatalf("expected validation_error, got %q", code)
	}

	// Create
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(
		http.MethodPost,
		"/api/v1/servers",
		bytes.NewBufferString(`{"name":"srv-1","hostname":"srv-1.local","ip":"10.0.0.1"}`),
	))
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
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/servers/"+id, nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("get status = %d, want %d", rec.Code, http.StatusOK)
	}

	// Get missing → 404 not_found
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/servers/missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if code := errorCode(t, rec.Body.Bytes()); code != "not_found" {
		t.Fatalf("expected not_found, got %q", code)
	}

	// Update
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(
		http.MethodPut,
		"/api/v1/servers/"+id,
		bytes.NewBufferString(`{"name":"srv-1-renamed","hostname":"srv-1.local","ip":"10.0.0.2","status":"online"}`),
	))
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
	mux.ServeHTTP(rec, httptest.NewRequest(
		http.MethodPut,
		"/api/v1/servers/missing",
		bytes.NewBufferString(`{"name":"x","hostname":"x.local"}`),
	))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("update missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if code := errorCode(t, rec.Body.Bytes()); code != "not_found" {
		t.Fatalf("expected not_found, got %q", code)
	}

	// Delete
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/servers/"+id, nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	// Delete missing → 404 not_found
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/servers/"+id, nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if code := errorCode(t, rec.Body.Bytes()); code != "not_found" {
		t.Fatalf("expected not_found, got %q", code)
	}
}
