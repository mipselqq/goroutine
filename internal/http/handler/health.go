package handler

import (
	"log/slog"
	"net/http"

	"goroutine/internal/http/httpschema"
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
// @Success 200 {object} httpschema.StatusResponse
// @Router /health [get]
func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	httpschema.RespondWithJSON(w, h.logger, http.StatusOK, httpschema.StatusResponse{Status: "ok"})
}
