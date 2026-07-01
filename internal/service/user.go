package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
)

type TelegramTokenRepository interface {
	InsertLinkToken(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error
}

type User struct {
	tokenRepository     TelegramTokenRepository
	telegramLinkTokenFn func() domain.TelegramLinkToken
}

func NewUser(tr TelegramTokenRepository, telegramLinkTokenFn func() domain.TelegramLinkToken) *User {
	return &User{
		tokenRepository:     tr,
		telegramLinkTokenFn: telegramLinkTokenFn,
	}
}

func (s *User) CreateTelegramLinkToken(ctx context.Context, userID domain.UserID) (domain.TelegramLinkToken, error) {
	token := s.telegramLinkTokenFn()

	err := s.tokenRepository.InsertLinkToken(ctx, token, userID)
	if err != nil {
		if errors.Is(err, repository.ErrKeyExists) {
			return domain.TelegramLinkToken{}, fmt.Errorf("user service: create telegram link token: %v: %w", err, ErrInternal)
		}
		return domain.TelegramLinkToken{}, fmt.Errorf("user service: create telegram link token: save token: %v: %w", err, ErrInternal)
	}

	return token, nil
}
