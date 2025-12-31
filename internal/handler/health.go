package handler

import "net/http"

type Health struct{}

func NewHealth() *Health {
	return &Health{}
}

func (h *Health) Health(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
