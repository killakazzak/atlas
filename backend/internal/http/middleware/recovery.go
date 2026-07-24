package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"

	"atlas/internal/httpx"
)

// Recovery catches panics, logs them with a stack trace, and returns HTTP 500.
func Recovery(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if val := recover(); val != nil {
					log.Error("panic recovered",
						"error", val,
						"stack", string(debug.Stack()),
						"request_id", RequestIDFromContext(r.Context()),
					)
					httpx.WriteError(w, http.StatusInternalServerError,
						"internal_error", "internal server error",
						RequestIDFromContext(r.Context()),
					)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
