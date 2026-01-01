package service

import (
	"context"
	"errors"
	"fmt"

	"go-todo/internal/repository"

	"github.com/alexedwards/argon2id"
)

type UserRepository interface {
	Insert(ctx context.Context, email, hash string) error
}

type Auth struct {
	repository UserRepository
}

func NewAuth(r UserRepository) *Auth {
	return &Auth{repository: r}
}

func (s *Auth) Register(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		return fmt.Errorf("auth service: register: creds check: %w", ErrInvalidCredentials)
	}

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("auth service: register: hash password: %w", ErrInternal)
	}

	err = s.repository.Insert(ctx, email, hash)
	if errors.Is(err, repository.ErrUniqueViolation) {
		return fmt.Errorf("auth service: register: user insert: %w", ErrUserAlreadyExists)
	}

	if err != nil {
		// TODO: log
		return fmt.Errorf("auth service: register: user insert: %w", ErrInternal)
	}

	return nil
}
