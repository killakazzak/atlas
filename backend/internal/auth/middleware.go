package auth

import (
	"context"
	"net/http"
	"strings"

	"atlas/internal/httpx"
)

// RequireAuth returns an HTTP middleware that enforces a valid Bearer token.
//
// On success the parsed *Claims are stored in the request context and
// can be retrieved with ClaimsFromContext.  The getRequestID parameter is
// called to populate the requestId field of error responses; pass
// middleware.RequestIDFromContext (or a no-op func) depending on what the
// outer middleware chain provides.
func RequireAuth(tokens TokenService, getRequestID func(context.Context) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := getRequestID(r.Context())

			raw := r.Header.Get("Authorization")
			if raw == "" {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing Authorization header", reqID)
				return
			}

			const prefix = "Bearer "
			if !strings.HasPrefix(raw, prefix) {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "Authorization header must use Bearer scheme", reqID)
				return
			}

			claims, err := tokens.Parse(strings.TrimPrefix(raw, prefix))
			if err != nil {
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token", reqID)
				return
			}

			next.ServeHTTP(w, r.WithContext(WithClaims(r.Context(), claims)))
		})
	}
}
