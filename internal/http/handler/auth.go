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
	logger    *slog.Logger
	service   AuthService
	responder *httpschema.ErrorResponder
}

func NewAuth(l *slog.Logger, s AuthService, responder *httpschema.ErrorResponder) *Auth {
	return &Auth{
		service:   s,
		logger:    logging.NewLoggerContext(l, "handler.auth"),
		responder: responder,
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
// @Success 200 {object} httpschema.Status
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR or INVALID_CREDENTIALS"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/register [post]
func (h *Auth) Register(w http.ResponseWriter, r *http.Request) {
	var body registerBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.responder.BadRequest(
			w, "VALIDATION_ERROR",
			[]httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}},
		)
		return
	}

	details := []httpschema.Detail{}
	email := validateCredential("email", body.Email, domain.NewEmail, &details)
	password := validateCredential("password", body.Password, domain.NewPassword, &details)
	if len(details) > 0 {
		h.responder.BadRequest(w, "VALIDATION_ERROR", details)
		return
	}

	err = h.service.Register(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrUserAlreadyExists),
			errors.Is(err, service.ErrInvalidCredentials):
			h.responder.BadRequest(
				w, "INVALID_CREDENTIALS",
				[]httpschema.Detail{{Field: "email or password", Issues: []string{"Invalid credentials"}}},
			)
		default:
			h.logger.Error("Failed to register user", slog.String("err", err.Error()))
			h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
		}
		return
	}

	h.logger.Info("Successfuly registered user", slog.String("email", body.Email))

	httpschema.RespondJSON(w, h.logger, http.StatusOK, httpschema.Status{Status: "ok"})
}

type loginBody struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"secret-password"`
}

type loginResponse struct {
	Token string `json:"token" example:"jwt-token"`
}

// TODO: add extractor with truncation in debug
func SlogRequestIDFromRequest(r *http.Request) any {
	return slog.Any("request_id", r.Context().Value(httpschema.ContextKeyRequestID))
}

// Login godoc
// @Summary Login a user
// @Description Login with email and password to get a JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body loginBody true "Login credentials"
// @Success 200 {object} loginResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "INVALID_CREDENTIALS or USER_NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/login [post]
func (h *Auth) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.responder.BadRequest(
			w, "VALIDATION_ERROR",
			[]httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}},
		)
		return
	}

	details := []httpschema.Detail{}
	email := validateCredential("email", body.Email, domain.NewEmail, &details)
	password := validateCredential("password", body.Password, domain.NewPassword, &details)
	if len(details) > 0 {
		h.responder.BadRequest(w, "VALIDATION_ERROR", details)
		return
	}

	token, err := h.service.Login(r.Context(), email, password)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			h.responder.Unauthorized(w, "INVALID_CREDENTIALS", []httpschema.Detail{{Field: "email or password", Issues: []string{"Invalid"}}})
		case errors.Is(err, service.ErrUserNotFound):
			h.responder.Unauthorized(w, "USER_NOT_FOUND", []httpschema.Detail{})
		default:
			h.logger.Error("Failed to login user", slog.String("err", err.Error()), SlogRequestIDFromRequest(r))
			h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
		}
		return
	}

	h.logger.Info("Successfuly logged in user", slog.String("email", body.Email))
	httpschema.RespondJSON(w, h.logger, http.StatusOK, loginResponse{Token: token})
}

func validateCredential[T any](field, val string, constructor func(string) (T, []string), details *[]httpschema.Detail) T {
	res, errs := constructor(val)
	if len(errs) > 0 {
		*details = append(*details, httpschema.Detail{Field: field, Issues: errs})
	}
	return res
}

type whoAmIResponse struct {
	UID string `json:"uid" example:"018e1000-0000-7000-8000-000000000000"`
}

// WhoAmI godoc
// @Summary Get current user info
// @Description Get current user ID from token. This is a protected endpoint — in addition to the responses listed below, the auth middleware may return 401 with codes INVALID_AUTH_HEADER or INVALID_TOKEN.
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} whoAmIResponse
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN (handler) or INVALID_AUTH_HEADER / INVALID_TOKEN (auth middleware)"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/whoami [get]
func (h *Auth) WhoAmI(w http.ResponseWriter, r *http.Request) {
	// TODO: cover error path with tests
	uid, ok := r.Context().Value(httpschema.ContextKeyUserID).(domain.UserID)
	if !ok {
		h.logger.Error("BUG: failed to get user id from context", SlogRequestIDFromRequest(r))
		h.responder.Error(
			w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR",
		)
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, whoAmIResponse{UID: uid.String()})
}
