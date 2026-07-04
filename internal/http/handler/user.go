package handler

import (
	"context"
	"log/slog"
	"net/http"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/logging"
)

type UserService interface {
	CreateTelegramLinkToken(ctx context.Context, userID domain.UserID) (domain.TelegramLinkToken, error)
}

type User struct {
	logger      *slog.Logger
	userService UserService
	responder   *httpschema.ErrorResponder
}

func NewUser(l *slog.Logger, userService UserService, responder *httpschema.ErrorResponder) *User {
	return &User{
		logger:      logging.WithModule(l, "handler.user"),
		userService: userService,
		responder:   responder,
	}
}

type telegramLinkTokenResponse struct {
	Token string `json:"token" example:"018e1000-0000-7000-8000-000000000000"`
}

// CreateTelegramLinkToken godoc
// @Summary Generate a Telegram link token
// @Description Creates a one-time token that a user can send to the Telegram bot via /start to link their Telegram account to the app.
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} telegramLinkTokenResponse
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/users/me/telegram/link [post]
func (u *User) CreateTelegramLinkToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := extractUserIDOrHandleMissing(w, r, u.logger, u.responder)
	if !ok {
		return
	}

	token, err := u.userService.CreateTelegramLinkToken(r.Context(), userID)
	if err != nil {
		u.responder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, u.logger, http.StatusOK, telegramLinkTokenResponse{Token: token.RevealSecret()})
}
