package middleware

import (
	"errors"
	"net/http"
	"strings"

	"atlas/internal/auth"
	"atlas/internal/httpx"
)

// JWTAuth returns a middleware that enforces a valid Bearer token.
// Requests without a token or with an invalid token receive 401.
// On success the parsed Claims are stored in the request context.
func JWTAuth(tokens auth.TokenService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := RequestIDFromContext(r.Context())

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
				if errors.Is(err, auth.ErrInvalidCredentials) {
					httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token", reqID)
					return
				}
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid or expired token", reqID)
				return
			}

			next.ServeHTTP(w, r.WithContext(auth.WithClaims(r.Context(), claims)))
		})
	}
}
