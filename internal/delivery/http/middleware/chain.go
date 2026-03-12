package middleware

import "net/http"

// Middleware type is just a function that wraps a handler
type Middleware func(http.Handler) http.Handler

// Chain applies a list of middlewares around a final handler.
// It applies them in the given order (first in the list runs first).
func Chain(h http.Handler, m ...Middleware) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}

	return h
}
