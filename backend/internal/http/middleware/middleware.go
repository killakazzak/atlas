// Package middleware provides HTTP middleware for Atlas services.
package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler.
type Middleware func(http.Handler) http.Handler

// Chain applies middleware in order: first wraps outermost, last wraps innermost.
// Given Chain(A, B, C), the call order is: A → B → C → handler.
func Chain(mw ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for i := len(mw) - 1; i >= 0; i-- {
			h = mw[i](h)
		}
		return h
	}
}
