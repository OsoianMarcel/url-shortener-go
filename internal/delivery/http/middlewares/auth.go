package middlewares

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/common"
	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/utils"
)

func AuthenticationMiddleware(apiSecret string, logger *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			auth := r.Header.Get("Authorization")
			if auth == "" {
				utils.JsonResponse(w, logger, http.StatusUnauthorized, common.ErrResponseDto{Error: "The auth token is missing."})
				return
			}

			authParts := strings.Split(auth, "Bearer ")
			if len(authParts) != 2 {
				utils.JsonResponse(w, logger, http.StatusUnauthorized, common.ErrResponseDto{Error: "Invalid Authorization header."})
				return
			}

			token := authParts[1]

			if token != apiSecret {
				utils.JsonResponse(w, logger, http.StatusUnauthorized, common.ErrResponseDto{Error: "Invalid token."})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
