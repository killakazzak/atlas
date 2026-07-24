package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlas/internal/inventory"
	inventoryhttp "atlas/internal/inventory/http"
)

func newTestMux(t *testing.T) *http.ServeMux {
	t.Helper()
	repo := inventory.NewMemoryServerRepository()
	svc := inventory.NewService(repo)
	mux := http.NewServeMux()
	inventoryhttp.NewHandler(svc).Register(mux)
	return mux
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

	// Invalid create
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(
		http.MethodPost,
		"/api/v1/servers",
		bytes.NewBufferString(`{"name":""}`),
	))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("create invalid status = %d, want %d", rec.Code, http.StatusBadRequest)
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

	// Get missing
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/servers/missing", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("get missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}

	// Delete
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/servers/"+id, nil))
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	// Delete missing
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodDelete, "/api/v1/servers/"+id, nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("delete missing status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}
