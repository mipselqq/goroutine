package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

func respondWithJSON(w http.ResponseWriter, logger *slog.Logger, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		logger.Error("Failed to send response:", slog.String("err", err.Error()))
	}
}

func respondWithError(w http.ResponseWriter, logger *slog.Logger, code int, message error) {
	respondWithJSON(w, logger, code, errorResponse{Error: message.Error()})
}

type statusResponse struct {
	Status string `json:"status" example:"ok"`
}

type errorResponse struct {
	Error string `json:"error" example:"invalid email format"`
}
