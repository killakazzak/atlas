package auth

import "context"

type contextKey struct{}

// WithClaims stores claims in ctx.
func WithClaims(ctx context.Context, claims *Claims) context.Context {
	return context.WithValue(ctx, contextKey{}, claims)
}

// ClaimsFromContext retrieves claims stored by WithClaims.
// Returns nil when the context carries no claims.
func ClaimsFromContext(ctx context.Context) *Claims {
	c, _ := ctx.Value(contextKey{}).(*Claims)
	return c
}
