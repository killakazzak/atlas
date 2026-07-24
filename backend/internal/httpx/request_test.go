package httpx_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"atlas/internal/httpx"
)

type testPayload struct {
	Name string `json:"name"`
}

func makeRequest(body string) *http.Request {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	return r
}

func TestDecodeJSON_Success(t *testing.T) {
	r := makeRequest(`{"name":"atlas"}`)
	var dst testPayload
	if err := httpx.DecodeJSON(r, &dst); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dst.Name != "atlas" {
		t.Fatalf("expected name=atlas, got %q", dst.Name)
	}
}

func TestDecodeJSON_EmptyBody(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", nil)
	var dst testPayload
	if err := httpx.DecodeJSON(r, &dst); err == nil {
		t.Fatal("expected error for empty body")
	}
}

func TestDecodeJSON_MalformedJSON(t *testing.T) {
	r := makeRequest(`{not valid json}`)
	var dst testPayload
	if err := httpx.DecodeJSON(r, &dst); err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestDecodeJSON_UnknownField(t *testing.T) {
	r := makeRequest(`{"name":"atlas","unknown":"field"}`)
	var dst testPayload
	if err := httpx.DecodeJSON(r, &dst); err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestDecodeJSON_MultipleObjects(t *testing.T) {
	r := makeRequest(`{"name":"atlas"}{"name":"second"}`)
	var dst testPayload
	if err := httpx.DecodeJSON(r, &dst); err == nil {
		t.Fatal("expected error for multiple JSON objects")
	}
}
