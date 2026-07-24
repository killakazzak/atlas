package httpx

import "net/http"

type errorBody struct {
	Error     errorDetail `json:"error"`
	RequestID string      `json:"requestId,omitempty"`
}

type errorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError writes a structured JSON error response.
func WriteError(w http.ResponseWriter, status int, code, message, requestID string) {
	WriteJSON(w, status, errorBody{
		Error:     errorDetail{Code: code, Message: message},
		RequestID: requestID,
	})
}
