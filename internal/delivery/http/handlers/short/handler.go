package short

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/common"
	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/middlewares"
	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/utils"
	"github.com/OsoianMarcel/url-shortener/internal/entities"
	shortUsecase "github.com/OsoianMarcel/url-shortener/internal/usecases/short"
)

type shortLinkUsecase interface {
	Create(ctx context.Context, createInput shortUsecase.CreateInput) (shortUsecase.CreateOutput, error)
	OriginalURL(ctx context.Context, key string) (string, error)
	Expand(ctx context.Context, key string) (entities.ShortLink, error)
	Delete(ctx context.Context, key string) error
	Stats(ctx context.Context, key string) (shortUsecase.StatsOutput, error)
}

type handler struct {
	logger                  *slog.Logger
	usecase                 shortLinkUsecase
	linkNotFoundRedirectURL string
}

func RegisterHandler(
	router *http.ServeMux,
	logger *slog.Logger,
	usecase shortLinkUsecase,
	apiSecret string,
	linkNotFoundRedirectURL string,
) {
	h := &handler{
		logger:                  logger,
		usecase:                 usecase,
		linkNotFoundRedirectURL: linkNotFoundRedirectURL,
	}

	router.Handle("POST /api/shortener", middlewares.Chain(
		h.shorten(),
		middlewares.AuthenticationMiddleware(apiSecret, logger),
	))
	router.Handle("DELETE /api/shortener/{linkKey}", middlewares.Chain(
		h.delete(),
		middlewares.AuthenticationMiddleware(apiSecret, logger),
	))
	router.Handle("GET /api/shortener/{linkKey}/redirect", h.redirect())
	router.Handle("GET /api/shortener/{linkKey}/expand", h.expand())
	router.Handle("GET /api/shortener/{linkKey}/stats", middlewares.Chain(
		h.stats(),
		middlewares.AuthenticationMiddleware(apiSecret, logger),
	))
}

func (h *handler) shorten() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestDTO, err := utils.JsonBodyDecode[shortenRequestDTO](r)
		if err != nil {
			utils.JsonResponse(w, h.logger, http.StatusBadRequest, common.ErrResponseDto{Error: "Invalid JSON body."})
			return
		}

		out, err := h.usecase.Create(r.Context(), shortUsecase.CreateInput{
			OriginalURL: requestDTO.URL,
		})
		if err != nil {
			switch err {
			case shortUsecase.ErrInvalidURL:
				utils.JsonResponse(w, h.logger, http.StatusBadRequest, common.ErrResponseDto{Error: "Invalid URL."})
			default:
				h.logger.Error(
					"Handler.shorten",
					slog.Any("error", err),
				)
				utils.JsonResponse(w, h.logger, http.StatusInternalServerError, common.ErrResponseDto{
					Error: "Failed to create short URL, try again later.",
				})
			}
			return
		}

		resDTO := shortenResponseDTO{
			ShortURL: out.ShortURL,
			Key:      out.Key,
		}

		utils.JsonResponse(w, h.logger, http.StatusCreated, resDTO)
	})
}

func (h *handler) delete() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := h.usecase.Delete(r.Context(), r.PathValue("linkKey"))
		if err != nil {
			h.logger.Error(
				"Handler.delete",
				slog.Any("error", err),
			)
			utils.JsonResponse(w, h.logger, http.StatusInternalServerError, common.ErrResponseDto{
				Error: "Failed to delete the short URL, try again later.",
			})
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func (h *handler) redirect() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originalURL, err := h.usecase.OriginalURL(r.Context(), r.PathValue("linkKey"))
		if err != nil {
			if err == shortUsecase.ErrShortLinkNotFound {
				http.Redirect(w, r, h.linkNotFoundRedirectURL, http.StatusFound)
				return
			}

			h.logger.Error(
				"Handler.redirect",
				slog.Any("error", err),
			)
			utils.JsonResponse(w, h.logger, http.StatusInternalServerError, common.ErrResponseDto{
				Error: "Server error, please refresh the page.",
			})
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	})
}

func (h *handler) expand() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		shortUrlEntity, err := h.usecase.Expand(r.Context(), r.PathValue("linkKey"))
		if err != nil {
			if err == shortUsecase.ErrShortLinkNotFound {
				utils.JsonResponse(w, h.logger, http.StatusNotFound, common.ErrResponseDto{
					Error: "Link not found.",
				})
				return
			}

			h.logger.Error(
				"Handler.expand",
				slog.Any("error", err),
			)
			utils.JsonResponse(w, h.logger, http.StatusInternalServerError, common.ErrResponseDto{
				Error: "Failed to expand the short URL, try again later.",
			})
			return
		}

		resDTO := expandResponseDTO{
			URL: shortUrlEntity.OriginalURL,
		}

		utils.JsonResponse(w, h.logger, http.StatusOK, resDTO)
	})
}

func (h *handler) stats() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats, err := h.usecase.Stats(r.Context(), r.PathValue("linkKey"))
		if err != nil {
			if err == shortUsecase.ErrShortLinkNotFound {
				utils.JsonResponse(w, h.logger, http.StatusNotFound, common.ErrResponseDto{Error: "Not found."})
				return
			}

			h.logger.Error(
				"Handler.stats",
				slog.Any("error", err),
			)
			utils.JsonResponse(w, h.logger, http.StatusInternalServerError, common.ErrResponseDto{
				Error: "Failed to fetch stats, try again later.",
			})
			return
		}

		resDTO := statsResponseDTO{
			Hits:      stats.Hits,
			CreatedAt: stats.CreatedAt,
		}

		utils.JsonResponse(w, h.logger, http.StatusOK, resDTO)
	})
}
