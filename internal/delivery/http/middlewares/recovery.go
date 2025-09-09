package middlewares

import (
	"log/slog"
	"net/http"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/common"
	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/utils"
)

func RecoverMiddleware(logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// recover from panics and handle errors
			defer func() {
				if r := recover(); r != nil {
					logger.Error("Middlewares.RecoverMiddleware.", slog.Any("error", r))
					utils.JsonResponse(w, logger, http.StatusInternalServerError, common.ErrResponseDto{Error: "Internal server error."})
				}
			}()

			// call the next handler
			next.ServeHTTP(w, r)
		})
	}
}
