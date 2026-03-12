package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/httputil"
)

func AuthenticationMiddleware(apiSecret string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responder := httputil.NewJsonResponder(w, logger)

			auth := r.Header.Get("Authorization")
			if auth == "" {
				responder.Unauthorized("The auth token is missing.")
				return
			}

			authParts := strings.Split(auth, "Bearer ")
			if len(authParts) != 2 {
				responder.Unauthorized("Invalid Authorization header.")
				return
			}

			token := authParts[1]

			if token != apiSecret {
				responder.Unauthorized("Invalid token.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
