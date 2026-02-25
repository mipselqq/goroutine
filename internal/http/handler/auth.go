package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/logging"
	"goroutine/internal/service"
)

type AuthService interface {
	Register(ctx context.Context, email domain.Email, password domain.Password) error
	Login(ctx context.Context, email domain.Email, password domain.Password) (string, error)
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

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param body body registerBody true "Registration details"
// @Success 200 {object} httpschema.StatusResponse
// @Failure 400 {object} httpschema.ErrorResponse "Invalid input (email format or password)"
// @Failure 500 {object} httpschema.ErrorResponse "Internal Server Error"
// @Router /register [post]
func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var body registerBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.logger.Error("Failed to decode json body", slog.String("err", err.Error()))
		httpschema.RespondWithError(w, h.logger, http.StatusBadRequest, errors.New("invalid json body"))
		return
	}

	email, err := domain.NewEmail(body.Email)
	if err != nil {
		httpschema.RespondWithError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	password, err := domain.NewPassword(body.Password)
	if err != nil {
		httpschema.RespondWithError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	err = h.service.Register(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists),
			errors.Is(err, service.ErrInvalidCredentials):
			httpschema.RespondWithError(w, h.logger, http.StatusBadRequest, err)
		default:
			h.logger.Error("Failed to register user", slog.String("err", err.Error()))
			httpschema.RespondWithError(w, h.logger, http.StatusInternalServerError, service.ErrInternal)
		}
		return
	}

	h.logger.Info("Successfuly registered user", slog.String("email", body.Email))
	httpschema.RespondWithJSON(w, h.logger, http.StatusOK, httpschema.StatusResponse{Status: "ok"})
}

type loginBody struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret-password"`
}

type loginResponse struct {
	Token string `json:"token" example:"jwt-token"`
}

// Login godoc
// @Summary Login a user
// @Description Login with email and password to get a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body loginBody true "Login credentials"
// @Success 200 {object} loginResponse
// @Failure 401 {object} httpschema.ErrorResponse "Invalid credentials"
// @Failure 500 {object} httpschema.ErrorResponse "Internal Server Error"
// @Router /login [post]
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.logger.Error("Failed to decode json body", slog.String("err", err.Error()))
		httpschema.RespondWithError(w, h.logger, http.StatusBadRequest, errors.New("invalid json body"))
		return
	}

	email, err := domain.NewEmail(body.Email)
	if err != nil {
		httpschema.RespondWithError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	password, err := domain.NewPassword(body.Password)
	if err != nil {
		httpschema.RespondWithError(w, h.logger, http.StatusBadRequest, err)
		return
	}

	token, err := h.service.Login(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			httpschema.RespondWithError(w, h.logger, http.StatusUnauthorized, err)
		case errors.Is(err, service.ErrUserNotFound):
			httpschema.RespondWithError(w, h.logger, http.StatusUnauthorized, err)
		default:
			h.logger.Error("Failed to login user", slog.String("err", err.Error()))
			httpschema.RespondWithError(w, h.logger, http.StatusInternalServerError, service.ErrInternal)
		}
		return
	}

	h.logger.Info("Successfuly logged in user", slog.String("email", body.Email))
	httpschema.RespondWithJSON(w, h.logger, http.StatusOK, loginResponse{Token: token})
}

type whoAmIResponse struct {
	UID int64 `json:"uid" example:"1"`
}

// WhoAmI godoc
// @Summary Get current user info
// @Description Get current user ID from token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} whoAmIResponse
// @Failure 401 {object} httpschema.ErrorResponse "Unauthorized"
// @Router /whoami [get]
func (h *Auth) WhoAmI(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value(httpschema.ContextKeyUserID).(int64)
	if !ok {
		h.logger.Error("Failed to get user id from context")
		httpschema.RespondWithError(w, h.logger, http.StatusUnauthorized, errors.New("unauthorized"))
		return
	}

	httpschema.RespondWithJSON(w, h.logger, http.StatusOK, whoAmIResponse{UID: uid})
}
