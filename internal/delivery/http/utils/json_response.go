package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func JsonResponse[T any](w http.ResponseWriter, logger *slog.Logger, status int, v T) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		logger.Warn("failed to send JSON response", slog.Any("error", err))
	}
}

func JsonBodyDecode[T any](r *http.Request) (T, error) {
	var v T

	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	return v, nil
}
