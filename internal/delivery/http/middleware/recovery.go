package middleware

import (
	"log/slog"
	"net/http"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/httputil"
)

func RecoverMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// recover from panics and handle errors
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Middlewares.RecoverMiddleware.", slog.Any("error", r))
					responder := httputil.NewJsonResponder(w, logger)
					responder.ServerError()
				}
			}()

			// call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
