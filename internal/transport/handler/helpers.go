package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/matveevaolga/feature-vote/internal/transport/handler/dto"
)

func RespondWithError(w http.ResponseWriter, status int, message string, err error) {
	slog.Error("HTTP error response",
		"status", status,
		"message", message,
		"error", err,
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(dto.ErrorResponse{Error: message})
}

func RespondWithJSON(w http.ResponseWriter, status int, data interface{}) {
	slog.Debug("HTTP response",
		"status", status,
		"response_type", "json",
	)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			slog.Error("Failed to encode JSON response", "error", err)
		}
	}
}

func ExtractIDFromPath(path, prefix string) string {
	return strings.TrimPrefix(path, prefix)
}
