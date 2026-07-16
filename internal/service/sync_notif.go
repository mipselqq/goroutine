package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/driver"
	"goroutine/internal/repository"
)

type SyncNotifUserRepository interface {
	GetChatID(ctx context.Context, userID domain.UserID) (domain.TelegramChatID, error)
}

type SyncTelegramNotif struct {
	telegramClient *driver.TelegramClient
	userRepo       SyncNotifUserRepository
}

func NewSyncTelegramNotif(telegramClient *driver.TelegramClient, userRepo SyncNotifUserRepository) *SyncTelegramNotif {
	return &SyncTelegramNotif{
		telegramClient: telegramClient,
		userRepo:       userRepo,
	}
}

func (s *SyncTelegramNotif) NotifUser(ctx context.Context, userID domain.UserID, notification fmt.Stringer) error {
	chatID, err := s.userRepo.GetChatID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return nil
		}
		return fmt.Errorf("sync telegram notify: get chat id: %v: %w", err, ErrInternal)
	}

	return s.NotifChat(ctx, chatID, notification)
}

func (s *SyncTelegramNotif) NotifChat(ctx context.Context, chatID domain.TelegramChatID, notification fmt.Stringer) error {
	message, err := domain.NewTelegramMessage(notification.String())
	if err != nil {
		return fmt.Errorf("sync telegram notify: build message: %v: %w", err, ErrInternal)
	}

	err = s.telegramClient.SendMessage(ctx, chatID.Int64(), message)
	if err != nil {
		return fmt.Errorf("sync telegram notify: send message: %v: %w", err, ErrInternal)
	}

	return nil
}
