package common

import (
	"log/slog"
	"net/http"
)

type handler struct {
	openapiFilePath string
	logger          *slog.Logger
}

func RegisterHandler(
	router *http.ServeMux,
	logger *slog.Logger,
	openapiFilePath string,
) {
	h := &handler{
		logger:          logger,
		openapiFilePath: openapiFilePath,
	}

	router.Handle("GET /openapi-spec.yaml", h.serverFile())
}

func (h *handler) serverFile() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, h.openapiFilePath)
	})
}
