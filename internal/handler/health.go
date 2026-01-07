package handler

import (
	"log/slog"
	"net/http"
)

type Health struct {
	logger *slog.Logger
}

func NewHealth(logger *slog.Logger) *Health {
	return &Health{
		logger: logger,
	}
}

// Health godoc
// @Summary Health check
// @Description Check if the server is alive
// @Tags health
// @Produce json
// @Success 200 {object} statusResponse
// @Router /health [get]
func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	respondWithJSON(w, h.logger, http.StatusOK, statusResponse{Status: "ok"})
}
