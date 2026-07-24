package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"atlas/internal/httpx"
)

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	httpx.WriteError(w, http.StatusBadRequest, "validation_error", "hostname is required", "req-123")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var got struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if got.Error.Code != "validation_error" {
		t.Errorf("unexpected code: %q", got.Error.Code)
	}
	if got.Error.Message != "hostname is required" {
		t.Errorf("unexpected message: %q", got.Error.Message)
	}
	if got.RequestID != "req-123" {
		t.Errorf("unexpected requestId: %q", got.RequestID)
	}
}

func TestWriteError_NoRequestID(t *testing.T) {
	w := httptest.NewRecorder()
	httpx.WriteError(w, http.StatusInternalServerError, "internal_error", "something went wrong", "")

	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("response is not valid JSON: %v", err)
	}
	if _, ok := got["requestId"]; ok {
		t.Error("requestId should be omitted when empty")
	}
}
