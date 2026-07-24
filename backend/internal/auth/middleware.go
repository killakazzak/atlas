package auth

import (
	"context"
	"net/http"
	"strings"

	"atlas/internal/httpx"
)

// RequireRole returns a middleware that allows only requests whose JWT claims
// carry exactly the given role.  It must be chained after RequireAuth.
func RequireRole(role Role, getRequestID func(context.Context) string) func(http.Handler) http.Handler {
	return RequireRoles(getRequestID, role)
}

// RequireRoles returns a middleware that allows requests carrying any of the
// listed roles.  It must be chained after RequireAuth.
// Requests with a valid token but an insufficient role receive 403 Forbidden.
func RequireRoles(getRequestID func(context.Context) string, roles ...Role) func(http.Handler) http.Handler {
	allowed := make(map[Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ClaimsFromContext(r.Context())
			if claims == nil {
				// RequireAuth was not applied before this middleware.
				reqID := getRequestID(r.Context())
				httpx.WriteError(w, http.StatusUnauthorized, "unauthorized", "missing token claims", reqID)
				return
			}

			if _, ok := allowed[claims.Role]; !ok {
				reqID := getRequestID(r.Context())
				httpx.WriteError(w, http.StatusForbidden, "forbidden", "insufficient role", reqID)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

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
