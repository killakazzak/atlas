package middleware_test

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"io"
	"testing"

	"atlas/internal/http/middleware"
)

func nopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestRequestID_HeaderPresent(t *testing.T) {
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header to be set")
	}
}

func TestRequestID_InContext(t *testing.T) {
	var gotID string
	handler := middleware.RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotID = middleware.RequestIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	if gotID == "" {
		t.Fatal("expected request ID in context")
	}
	if gotID != w.Header().Get("X-Request-ID") {
		t.Fatalf("context ID %q != header ID %q", gotID, w.Header().Get("X-Request-ID"))
	}
}

func TestLogging_PreservesStatus(t *testing.T) {
	handler := middleware.Logging(nopLogger())(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/test", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestRecovery_PanicReturns500(t *testing.T) {
	chain := middleware.Chain(
		middleware.Recovery(nopLogger()),
		middleware.RequestID,
	)
	handler := chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went wrong")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestRecovery_PanicErrorFormat(t *testing.T) {
	chain := middleware.Chain(
		middleware.Recovery(nopLogger()),
		middleware.RequestID,
	)
	handler := chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	handler.ServeHTTP(w, r)

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if body.Error.Code != "internal_error" {
		t.Errorf("unexpected error code: %q", body.Error.Code)
	}
	if body.Error.Message == "" {
		t.Error("expected non-empty error message")
	}
}
