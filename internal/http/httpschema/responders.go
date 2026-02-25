package httpschema

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func RespondWithJSON(w http.ResponseWriter, logger *slog.Logger, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		logger.Warn("Failed to send response:", slog.String("err", err.Error()))
	}
}

func RespondWithError(w http.ResponseWriter, logger *slog.Logger, code int, message error) {
	RespondWithJSON(w, logger, code, ErrorResponse{Error: message.Error()})
}

type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"invalid email format"`
}
