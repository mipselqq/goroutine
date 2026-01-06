package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-todo/internal/domain"
	"go-todo/internal/repository"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

type UserRepository interface {
	Insert(ctx context.Context, email domain.Email, hash string) error
	GetPasswordHashByEmail(ctx context.Context, email domain.Email) (string, error)
}

type Auth struct {
	repository UserRepository
	JWTSecret  string
}

func NewAuth(r UserRepository, s string) *Auth {
	return &Auth{repository: r, JWTSecret: s}
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

func (s *Auth) Login(ctx context.Context, email domain.Email, password domain.Password) (string, error) {
	hash, err := s.repository.GetPasswordHashByEmail(ctx, email)
	if errors.Is(err, repository.ErrRowNotFound) {
		return "", fmt.Errorf("auth service: login: hash by email: %w", ErrUserNotFound)
	}
	if err != nil {
		return "", fmt.Errorf("auth service: login: hash by email: %v: %w", err, ErrInternal)
	}

	isMatch, err := argon2id.ComparePasswordAndHash(password.String(), hash)
	if err != nil {
		return "", fmt.Errorf("auth service: login: compare: %v: %w", err, ErrInternal)
	}
	if !isMatch {
		return "", ErrInvalidCredentials
	}

	token, err := CreateToken(email, s.JWTSecret)
	if err != nil {
		return "", fmt.Errorf("auth service: login: create token: %v: %w", err, ErrInternal)
	}
	return token, nil
}

func CreateToken(email domain.Email, secret string) (string, error) {
	claims := jwt.MapClaims{
		"sub": email,
		"exp": time.Now().Add(time.Hour * 24).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}
