package handler

import (
	"encoding/json"
	"net/http"
)

type Health struct{}

func NewHealth() *Health {
	return &Health{}
}

// Health godoc
// @Summary Health check
// @Description Check if the server is alive
// @Tags health
// @Produce json
// @Success 200 {object} statusResponse
// @Router /health [get]
func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(statusResponse{Status: "ok"})
}
