package service

import (
	"context"
	"errors"

	"go-todo/repository"

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
		return ErrInvalidCredentials
	}

	hash, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return ErrInternal
	}

	err = s.repository.Insert(ctx, email, hash)
	if errors.Is(err, repository.ErrUniqueViolation) {
		return ErrUserAlreadyExists
	}

	if err != nil {
		// TODO: log
		return ErrInternal
	}

	return nil
}
