package handler

import "net/http"

type Health struct{}

func NewHealth() *Health {
	return &Health{}
}

// Health godoc
// @Summary Health check
// @Description Check if the server is alive
// @Tags health
// @Produce json
// @Success 200
// @Router /health [get]
func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
