package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go-todo/internal/service"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (string, error)
}

type Auth struct {
	logger  *slog.Logger
	service AuthService
}

func NewAuth(s AuthService) *Auth {
	return &Auth{service: s}
}

type registerBody struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	Token string `json:"token"`
}

func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var body registerBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, errors.New("invalid json in body"))
		return
	}

	if body.Email == "" || body.Password == "" {
		respondWithError(w, http.StatusBadRequest, errors.New("empty email or password"))
		return
	}

	token, err := h.service.Register(r.Context(), body.Email, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists),
			errors.Is(err, service.ErrInvalidCredentials):
			respondWithError(w, http.StatusBadRequest, err)
		default:
			respondWithError(w, http.StatusInternalServerError, service.ErrInternal)
		}
		return
	}

	respondWithJSON(w, http.StatusOK, registerResponse{Token: token})
}
