package httputil

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/OsoianMarcel/url-shortener/internal/delivery/http/dto"
)

func JsonResponse[T any](w http.ResponseWriter, logger *slog.Logger, status int, v T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Warn(
			"failed to send JSON response",
			slog.String("func", "JsonResponse"),
			slog.Any("error", err),
		)
	}
}

func JsonBodyDecode[T any](r *http.Request) (T, error) {
	var v T

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	return v, nil
}

type jsonResponder struct {
	responseWriter http.ResponseWriter
	logger         *slog.Logger
}

func NewJsonResponder(w http.ResponseWriter, logger *slog.Logger) *jsonResponder {
	return &jsonResponder{
		responseWriter: w,
		logger:         logger,
	}
}

func (j *jsonResponder) Respond(httpStatus int, v any) {
	JsonResponse(j.responseWriter, j.logger, httpStatus, v)
}

func (j *jsonResponder) OK(v any) {
	j.Respond(http.StatusOK, v)
}

func (j *jsonResponder) Created(v any) {
	j.Respond(http.StatusCreated, v)
}

func (j *jsonResponder) MessageOK(message string) {
	j.OK(dto.MessageResponse{Message: message})
}

func (j *jsonResponder) Error(httpStatus int, message string) {
	j.Respond(httpStatus, dto.ErrResponse{
		Error: message,
	})
}

func (j *jsonResponder) NotFound(message string) {
	j.Error(http.StatusNotFound, message)
}

func (j *jsonResponder) Unauthorized(message string) {
	j.Error(http.StatusUnauthorized, message)
}

func (j *jsonResponder) Forbidden(message string) {
	j.Error(http.StatusForbidden, message)
}

func (j *jsonResponder) Conflict(message string) {
	j.Error(http.StatusConflict, message)
}

func (j *jsonResponder) BadRequest(message string) {
	j.Error(http.StatusBadRequest, message)
}

func (j *jsonResponder) NoContent() {
	j.responseWriter.WriteHeader(http.StatusNoContent)
}

func (j *jsonResponder) ServerError() {
	j.Respond(http.StatusInternalServerError, dto.ErrResponse{Error: "Internal server error."})
}

func (j *jsonResponder) InvalidJsonError() {
	j.BadRequest("Invalid JSON.")
}
