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
	logger         *slog.Logger
	userService    UserService
	errorResponder *httpschema.ErrorResponder
}

func NewUser(l *slog.Logger, userService UserService, errorResponder *httpschema.ErrorResponder) *User {
	return &User{
		logger:         logging.WithModule(l, "handler.user"),
		userService:    userService,
		errorResponder: errorResponder,
	}
}

func (u *User) CreateTelegramLinkToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := extractUserIDOrHandleMissing(w, r, u.logger, u.errorResponder)
	if !ok {
		return
	}

	token, err := u.userService.CreateTelegramLinkToken(r.Context(), userID)
	if err != nil {
		u.errorResponder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, u.logger, http.StatusOK, map[string]string{"token": token.RevealSecret()})
}
