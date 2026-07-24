package httpx

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// DecodeJSON decodes a JSON request body into dst.
// It rejects empty bodies, malformed JSON, unknown fields, and multiple objects.
func DecodeJSON(r *http.Request, dst any) error {
	if r.Body == nil || r.ContentLength == 0 {
		return errors.New("request body is empty")
	}

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	if err := dec.Decode(dst); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalErr *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf("malformed JSON at position %d", syntaxErr.Offset)
		case errors.As(err, &unmarshalErr):
			return fmt.Errorf("invalid value for field %q", unmarshalErr.Field)
		case errors.Is(err, io.EOF):
			return errors.New("request body is empty")
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			field := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown field %s", field)
		default:
			return err
		}
	}

	// Reject trailing data (multiple JSON objects).
	if dec.More() {
		return errors.New("request body must contain a single JSON object")
	}

	return nil
}
