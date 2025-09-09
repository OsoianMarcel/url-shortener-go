package middlewares

import "net/http"

// Middleware type is just a function that wraps a handler
type Middleware func(http.Handler) http.Handler
