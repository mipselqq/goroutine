package service

import (
	"context"
	"errors"
	"fmt"

	"go-todo/internal/domain"
	"go-todo/internal/repository"

	"github.com/alexedwards/argon2id"
)

type UserRepository interface {
	Insert(ctx context.Context, email domain.Email, hash string) error
}

type Auth struct {
	repository UserRepository
}

func NewAuth(r UserRepository) *Auth {
	return &Auth{repository: r}
}

func (s *Auth) Register(ctx context.Context, email domain.Email, password domain.Password) error {
	hash, err := argon2id.CreateHash(password.String(), argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("auth service: register: hash password: %v: %w", err, ErrInternal)
	}

	err = s.repository.Insert(ctx, email, hash)
	if errors.Is(err, repository.ErrUniqueViolation) {
		return fmt.Errorf("auth service: register: user insert: %w", ErrUserAlreadyExists)
	}

	if err != nil {
		return fmt.Errorf("auth service: register: user insert: %v: %w", err, ErrInternal)
	}

	return nil
}
