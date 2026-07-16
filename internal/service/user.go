package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/template"
)

type UserRepository interface {
	UpdateTelegramInfo(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

type TelegramTokenRepository interface {
	InsertLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error
	ConsumeTelegramLinkToken(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error)
}

type telegramLinkNotif interface {
	NotifChat(ctx context.Context, chatID domain.TelegramChatID, notification fmt.Stringer) error
}

type User struct {
	userRepo            UserRepository
	tokenRepo           TelegramTokenRepository
	notifService        telegramLinkNotif
	telegramLinkTokenFn func() domain.TelegramLinkToken
}

func NewUser(
	userRepo UserRepository,
	tokenRepo TelegramTokenRepository,
	notifService telegramLinkNotif,
	telegramLinkTokenFn func() domain.TelegramLinkToken,
) *User {
	return &User{
		userRepo:            userRepo,
		tokenRepo:           tokenRepo,
		notifService:        notifService,
		telegramLinkTokenFn: telegramLinkTokenFn,
	}
}

func (s *User) CreateTelegramLinkToken(ctx context.Context, userID domain.UserID) (domain.TelegramLinkToken, error) {
	token := s.telegramLinkTokenFn()

	err := s.tokenRepo.InsertLinkToken(ctx, token, userID)
	if err != nil {
		if errors.Is(err, repository.ErrKeyExists) {
			return domain.TelegramLinkToken{}, fmt.Errorf("user service: create telegram link token: %v: %w", err, ErrInternal)
		}
		return domain.TelegramLinkToken{}, fmt.Errorf("user service: create telegram link token: save token: %v: %w", err, ErrInternal)
	}

	return token, nil
}

func (s *User) LinkTelegramByToken(ctx context.Context, token domain.TelegramLinkToken, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
	userID, err := s.tokenRepo.ConsumeTelegramLinkToken(ctx, token)
	if err != nil {
		if errors.Is(err, repository.ErrKeyNotFound) {
			return s.notifTelegramLinkResult(ctx, chatID, template.TelegramLinkTokenExpiredNotif{}, ErrTelegramLinkTokenNotFound)
		}

		linkErr := fmt.Errorf("user service: link telegram by token: %v: %w", err, ErrInternal)
		return s.notifTelegramLinkResult(ctx, chatID, template.TelegramLinkFailedNotif{}, linkErr)
	}

	// If DB fails here, the token is going to be consumed without actual linking.
	// However, the user will be able to create a new token and try again. Nothing critical.
	err = s.userRepo.UpdateTelegramInfo(ctx, userID, chatID, username)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return s.notifTelegramLinkResult(ctx, chatID, template.TelegramUserNotFoundNotif{}, ErrUserNotFound)
		}

		linkErr := fmt.Errorf("user service: link update telegram info: %v: %w", err, ErrInternal)
		return s.notifTelegramLinkResult(ctx, chatID, template.TelegramLinkFailedNotif{}, linkErr)
	}

	return s.notifTelegramLinkResult(ctx, chatID, template.TelegramLinkedNotif{}, nil)
}

func (s *User) notifTelegramLinkResult(
	ctx context.Context,
	chatID domain.TelegramChatID,
	notification fmt.Stringer,
	linkErr error,
) error {
	err := s.notifService.NotifChat(ctx, chatID, notification)
	if err != nil {
		return fmt.Errorf("user service: notify telegram link result: %v: %w", err, ErrInternal)
	}

	return linkErr
}
