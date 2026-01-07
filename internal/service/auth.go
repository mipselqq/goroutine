package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go-todo/internal/domain"
	"go-todo/internal/repository"
	"go-todo/internal/secrecy"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

type UserRepository interface {
	Insert(ctx context.Context, email domain.Email, hash string) error
	GetPasswordHashByEmail(ctx context.Context, email domain.Email) (string, error)
}

type JWTOptions struct {
	JWTSecret secrecy.SecretString
	Exp       time.Duration
}

type Auth struct {
	repository UserRepository
	jwtOptions JWTOptions
}

func NewAuth(r UserRepository, opts JWTOptions) *Auth {
	return &Auth{repository: r, jwtOptions: opts}
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

	token, err := CreateToken(email, s.jwtOptions.JWTSecret.RevealSecret(), s.jwtOptions.Exp)
	if err != nil {
		return "", fmt.Errorf("auth service: login: create token: %v: %w", err, ErrInternal)
	}
	return token, nil
}

func CreateToken(email domain.Email, secret string, exp time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"sub": email.String(),
		"exp": time.Now().Add(exp).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(secret))
}

func (s *Auth) VerifyToken(ctx context.Context, tokenString string) (domain.Email, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtOptions.JWTSecret.RevealSecret()), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.Email{}, fmt.Errorf("auth service: verify token: %w", ErrTokenExpired)
		}
		return domain.Email{}, fmt.Errorf("auth service: verify token: %v: %w", err, ErrInvalidToken)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		emailStr, ok := claims["sub"].(string)
		if !ok {
			return domain.Email{}, fmt.Errorf("auth service: verify token: sub claim not found: %w", ErrInvalidToken)
		}
		email, err := domain.NewEmail(emailStr)
		if err != nil {
			return domain.Email{}, fmt.Errorf("auth service: verify token: invalid email in sub: %v: %w", err, ErrInvalidToken)
		}
		return email, nil
	}

	return domain.Email{}, fmt.Errorf("auth service: verify token: %w", ErrInvalidToken)
}
