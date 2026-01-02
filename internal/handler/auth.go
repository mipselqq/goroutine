package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go-todo/internal/logging"
	"go-todo/internal/service"
)

type AuthService interface {
	Register(ctx context.Context, email, password string) (string, error)
}

type Auth struct {
	logger  *slog.Logger
	service AuthService
}

func NewAuth(l *slog.Logger, s AuthService) *Auth {
	return &Auth{
		service: s,
		logger:  logging.NewLoggerContext(l, "handler.auth"),
	}
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
		h.logger.Error("Failed to decode json body", slog.String("err", err.Error()))
		respondWithError(w, h.logger, http.StatusBadRequest, errors.New("invalid json body"))
		return
	}

	if body.Email == "" || body.Password == "" {
		respondWithError(w, h.logger, http.StatusBadRequest, errors.New("empty email or password"))
		return
	}

	token, err := h.service.Register(r.Context(), body.Email, body.Password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists),
			errors.Is(err, service.ErrInvalidCredentials):
			respondWithError(w, h.logger, http.StatusBadRequest, err)
		default:
			h.logger.Error("Failed to register user", slog.String("err", err.Error()))
			respondWithError(w, h.logger, http.StatusInternalServerError, service.ErrInternal)
		}
		return
	}

	h.logger.Info("Successfuly registered user", slog.String("email", body.Email))
	respondWithJSON(w, h.logger, http.StatusOK, registerResponse{Token: token})
}
