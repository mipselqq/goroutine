package handler

import (
	"log/slog"
	"net/http"

	"goroutine/internal/http/httpschema"
	"goroutine/internal/logging"
)

type Health struct {
	logger *slog.Logger
}

func NewHealth(logger *slog.Logger) *Health {
	moduleLogger := logging.WithModule(logger, "handler.health")

	return &Health{
		logger: moduleLogger,
	}
}

// Health godoc
// @Summary Health check
// @Description Check if the server is alive
// @Tags health
// @Produce json
// @Success 200 {object} httpschema.Status
// @Router /v1/health [get]
func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	httpschema.RespondJSON(w, h.logger, http.StatusOK, httpschema.Status{Status: "ok"})
}
