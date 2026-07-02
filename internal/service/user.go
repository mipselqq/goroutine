package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
)

type UserRepository interface {
	UpdateTelegramInfo(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error
}

type TelegramTokenRepository interface {
	InsertLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error
	ConsumeTelegramLinkToken(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error)
}

type User struct {
	userRepo            UserRepository
	tokenRepo           TelegramTokenRepository
	telegramLinkTokenFn func() domain.TelegramLinkToken
}

func NewUser(ur UserRepository, tr TelegramTokenRepository, telegramLinkTokenFn func() domain.TelegramLinkToken) *User {
	return &User{
		userRepo:            ur,
		tokenRepo:           tr,
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
			return ErrTelegramLinkTokenNotFound
		}
		return fmt.Errorf("user service: link telegram by token: %v: %w", err, ErrInternal)
	}

	// If DB fails here, the token is going to be consumed without actual linking.
	// However, the user will be able to create a new token and try again. Nothing critical.
	err = s.userRepo.UpdateTelegramInfo(ctx, userID, chatID, username)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("user service: link update telegram info: %v: %w", err, ErrInternal)
	}

	return nil
}
