package health

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/httputil"
	"github.com/OsoianMarcel/url-shortener/internal/domain"
)

type handler struct {
	usecase domain.HealthUsecase
	logger  *slog.Logger
}

func RegisterHandler(
	router *http.ServeMux,
	logger *slog.Logger,
	usecase domain.HealthUsecase,
) {
	h := &handler{
		usecase: usecase,
		logger:  logger,
	}

	router.Handle("GET /health", h.health())

}

func (h *handler) health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responder := httputil.NewJsonResponder(w, h.logger)

		healthCheck := h.usecase.CheckHealth(r.Context())

		serviceHealthDTOs := make([]serviceHealthDTO, 0, len(healthCheck.Services))
		for _, serviceCheck := range healthCheck.Services {
			serviceHealthDTOs = append(serviceHealthDTOs, serviceHealthDTO{
				Name:          serviceCheck.Name,
				Healthy:       serviceCheck.Healthy,
				Error:         serviceCheck.Error,
				CheckDuration: serviceCheck.CheckDuration.String(),
			})
		}

		resDTO := healthResponseDTO{
			AllHealthy: healthCheck.AllHealthy,
			Services:   serviceHealthDTOs,
			ServerTime: time.Now(),
		}

		// return 200 if all services are healthy, otherwise return 500 error
		if healthCheck.AllHealthy {
			responder.OK(resDTO)
		} else {
			responder.Respond(http.StatusInternalServerError, resDTO)
		}
	})
}
