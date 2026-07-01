package telegram

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"goroutine/internal/domain"
)

type UserService interface {
	LinkTelegramByToken(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

type WebhookHandler struct {
	userService UserService
	logger      *slog.Logger
}

func NewWebhookHandler(us UserService, logger *slog.Logger) *WebhookHandler {
	return &WebhookHandler{
		userService: us,
		logger:      logger,
	}
}

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
			h.logger.WarnContext(r.Context(), "Failed to decode telegram update", slog.String("err", err.Error()))
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

	username, err := domain.NewTelegramUsername(update.Message.Chat.Username)
	if err != nil {
		h.logger.WarnContext(r.Context(), "Invalid username from telegram", slog.String("err", err.Error()))
		w.WriteHeader(http.StatusOK)
		return
	}

	err = h.userService.LinkTelegramByToken(r.Context(), linkToken, chatID, username)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to link telegram by token", slog.String("err", err.Error()))
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
}
