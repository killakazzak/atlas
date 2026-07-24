// Package httpx provides shared HTTP helpers for Atlas HTTP handlers.
package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

// WriteJSON encodes value as JSON and writes it with the given status code.
// If encoding fails after the header has been sent, the error is logged.
func WriteJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("httpx: failed to encode JSON response: %v", err)
	}
}
