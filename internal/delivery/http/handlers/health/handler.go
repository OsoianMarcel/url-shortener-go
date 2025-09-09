package health

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/utils"
	"github.com/OsoianMarcel/url-shortener/internal/usecases/health"
)

type healthUsecase interface {
	CheckHealth(ctx context.Context) health.HealthCheckOutput
}

type handler struct {
	usecase healthUsecase
	logger  *slog.Logger
}

func RegisterHandler(
	router *http.ServeMux,
	logger *slog.Logger,
	usecase healthUsecase,
) {
	h := &handler{
		usecase: usecase,
		logger:  logger,
	}

	router.Handle("GET /health", h.health())

}

func (h *handler) health() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			utils.JsonResponse(w, h.logger, http.StatusOK, resDTO)
		} else {
			utils.JsonResponse(w, h.logger, http.StatusInternalServerError, resDTO)
		}
	})
}
