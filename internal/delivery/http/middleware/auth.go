package middleware

import (
	"crypto/subtle"
	"log/slog"
	"net/http"
	"strings"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/httputil"
)

func AuthenticationMiddleware(apiSecret string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responder := httputil.NewJsonResponder(w, logger)

			auth := strings.TrimSpace(r.Header.Get("Authorization"))
			if auth == "" {
				responder.Unauthorized("The auth token is missing.")
				return
			}

			authParts := strings.Fields(auth)
			if len(authParts) != 2 || !strings.EqualFold(authParts[0], "Bearer") {
				responder.Unauthorized("Invalid Authorization header.")
				return
			}

			token := authParts[1]
			if subtle.ConstantTimeCompare([]byte(token), []byte(apiSecret)) != 1 {
				responder.Unauthorized("Invalid token.")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
