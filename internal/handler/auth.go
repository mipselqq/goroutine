package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"go-todo/internal/domain"
	"go-todo/internal/logging"
	"go-todo/internal/service"
)

type AuthService interface {
	Register(ctx context.Context, email domain.Email, password domain.Password) error
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
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret-password"`
}

type statusResponse struct {
	Status string `json:"status" example:"ok"`
}

type errorResponse struct {
	Error string `json:"error" example:"invalid email format"`
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param body body registerBody true "Registration details"
// @Success 200 {object} statusResponse
// @Failure 400 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Router /register [post]
func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var body registerBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.logger.Error("Failed to decode json body", slog.String("err", err.Error()))
		respondWithError(w, h.logger, http.StatusBadRequest, errors.New("invalid json body"))
		return
	}

	email, err := domain.NewEmail(body.Email)
	if err != nil {
		respondWithError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	password, err := domain.NewPassword(body.Password)
	if err != nil {
		respondWithError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	err = h.service.Register(r.Context(), email, password)
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
	respondWithJSON(w, h.logger, http.StatusOK, statusResponse{Status: "ok"})
}
