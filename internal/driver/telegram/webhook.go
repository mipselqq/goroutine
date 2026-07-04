package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"goroutine/internal/domain"
	"goroutine/internal/service"
)

type UserService interface {
	LinkTelegramByToken(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

type Notifier interface {
	Notify(ctx context.Context, chatID domain.TelegramChatID, text domain.TelegramMessage) error
}

type WebhookHandler struct {
	userService UserService
	notifier    Notifier
	logger      *slog.Logger
}

func NewWebhookHandler(us UserService, notifier Notifier, logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		userService: us,
		notifier:    notifier,
		logger:      logger,
	}
}

// ServeHTTP godoc
// @Summary Receive Telegram webhook updates
// @Description Receives update objects from Telegram Bot API. Processes /start command with a link token to link a Telegram account to the user.
// @Tags webhook
// @Accept json
// @Produce json
// @Success 200 "Always returns 200 OK per Telegram webhook protocol"
// @Router /webhook/telegram [post]
func (h *WebhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	const maxBodySize = 10 * 1024 // 10KB is more than enough for a Telegram update

	var update struct {
		Message struct {
			Text string `json:"text"`
			Chat struct {
				ID       int64  `json:"id"`
				Username string `json:"username"`
			} `json:"chat"`
		} `json:"message"`
	}

	err := json.NewDecoder(http.MaxBytesReader(w, r.Body, maxBodySize)).Decode(&update)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			h.logger.WarnContext(r.Context(), "Telegram update body too large")
		} else {
			h.logger.ErrorContext(r.Context(), "Failed to decode telegram update", slog.String("err", err.Error()))
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	tokenStr, ok := strings.CutPrefix(update.Message.Text, "/start ")
	if !ok {
		h.logger.DebugContext(r.Context(), "Ignoring non-start message")
		w.WriteHeader(http.StatusOK)
		return
	}

	linkToken, err := domain.NewTelegramLinkToken(tokenStr)
	if err != nil {
		h.logger.DebugContext(r.Context(), "Ignoring invalid link token in /start")
		w.WriteHeader(http.StatusOK)
		return
	}

	chatID, err := domain.NewTelegramChatID(update.Message.Chat.ID)
	if err != nil {
		h.logger.WarnContext(r.Context(), "Invalid chat id from telegram", slog.String("err", err.Error()))
		w.WriteHeader(http.StatusOK)
		return
	}

	username, err := domain.NewTelegramUsername("@" + update.Message.Chat.Username)
	if err != nil {
		h.logger.WarnContext(r.Context(), "Invalid username from telegram", slog.String("err", err.Error()))
		w.WriteHeader(http.StatusOK)
		return
	}

	msg := domain.MustTelegramMessage("Something went wrong. Please try again later.")
	err = h.userService.LinkTelegramByToken(r.Context(), linkToken, chatID, username)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrTelegramLinkTokenNotFound):
			msg = domain.MustTelegramMessage("This link has expired or is invalid. Please generate a new link in the app.")
		case errors.Is(err, service.ErrUserNotFound):
			msg = domain.MustTelegramMessage("User account not found.")
		}

		h.logger.ErrorContext(r.Context(), "Failed to link telegram by token", slog.String("err", err.Error()))
	} else {
		msg = domain.MustTelegramMessage("Successfully linked your account <3")
	}

	err = h.notifier.Notify(r.Context(), chatID, msg)
	if err != nil {
		h.logger.WarnContext(r.Context(), "telegram link notify failed", slog.String("err", err.Error()))
	}

	w.WriteHeader(http.StatusOK)
}
