package short

import (
	"log/slog"
	"net/http"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/httputil"
	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/middleware"
	"github.com/OsoianMarcel/url-shortener/internal/domain"
)

type handler struct {
	logger                  *slog.Logger
	usecase                 domain.ShortLinkUsecase
	linkNotFoundRedirectURL string
}

func RegisterHandler(
	router *http.ServeMux,
	logger *slog.Logger,
	usecase domain.ShortLinkUsecase,
	apiSecret string,
	linkNotFoundRedirectURL string,
) {
	h := &handler{
		logger:                  logger,
		usecase:                 usecase,
		linkNotFoundRedirectURL: linkNotFoundRedirectURL,
	}

	router.Handle("POST /api/shortener", middleware.Chain(
		h.shorten(),
		middleware.AuthenticationMiddleware(apiSecret, logger),
	))
	router.Handle("DELETE /api/shortener/{linkKey}", middleware.Chain(
		h.delete(),
		middleware.AuthenticationMiddleware(apiSecret, logger),
	))
	router.Handle("GET /api/shortener/{linkKey}/redirect", h.redirect())
	router.Handle("GET /api/shortener/{linkKey}/expand", h.expand())
	router.Handle("GET /api/shortener/{linkKey}/stats", middleware.Chain(
		h.stats(),
		middleware.AuthenticationMiddleware(apiSecret, logger),
	))
}

func (h *handler) shorten() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responder := httputil.NewJsonResponder(w, h.logger)

		requestDTO, err := httputil.JsonBodyDecode[shortenRequestDTO](r)
		if err != nil {
			responder.InvalidJsonError()
			return
		}

		out, err := h.usecase.Create(r.Context(), domain.CreateAction{
			OriginalURL: requestDTO.URL,
		})
		if err != nil {
			switch err {
			case domain.ErrInvalidURL:
				responder.BadRequest("Invalid URL.")
			default:
				h.logger.Error(
					"Handler.shorten",
					slog.Any("error", err),
				)
				responder.ServerError()
			}
			return
		}

		resDTO := shortenResponseDTO{
			ShortURL: out.ShortURL,
			Key:      out.Key,
		}
		responder.Created(resDTO)
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
			responder := httputil.NewJsonResponder(w, h.logger)
			responder.ServerError()
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})
}

func (h *handler) redirect() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		originalURL, err := h.usecase.OriginalURL(r.Context(), r.PathValue("linkKey"))
		if err != nil {
			if err == domain.ErrShortLinkNotFound {
				http.Redirect(w, r, h.linkNotFoundRedirectURL, http.StatusFound)
				return
			}

			h.logger.Error(
				"Handler.redirect",
				slog.Any("error", err),
			)
			responder := httputil.NewJsonResponder(w, h.logger)
			responder.ServerError()
			return
		}

		http.Redirect(w, r, originalURL, http.StatusFound)
	})
}

func (h *handler) expand() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responder := httputil.NewJsonResponder(w, h.logger)
		shortUrlEntity, err := h.usecase.Expand(r.Context(), r.PathValue("linkKey"))
		if err != nil {
			if err == domain.ErrShortLinkNotFound {
				responder.NotFound("Link not found.")
				return
			}

			h.logger.Error(
				"Handler.expand",
				slog.Any("error", err),
			)
			responder.ServerError()
			return
		}

		resDTO := expandResponseDTO{
			URL: shortUrlEntity.OriginalURL,
		}

		responder.OK(resDTO)
	})
}

func (h *handler) stats() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		responder := httputil.NewJsonResponder(w, h.logger)
		stats, err := h.usecase.Stats(r.Context(), r.PathValue("linkKey"))
		if err != nil {
			if err == domain.ErrShortLinkNotFound {
				responder.NotFound("Link not found.")
				return
			}

			h.logger.Error(
				"Handler.stats",
				slog.Any("error", err),
			)
			responder.ServerError()
			return
		}

		resDTO := statsResponseDTO{
			Hits:      stats.Hits,
			CreatedAt: stats.CreatedAt,
		}

		responder.OK(resDTO)
	})
}
