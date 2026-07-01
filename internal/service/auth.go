// Package service provides business logic for the application, doing service-level error handling,
// managing authentication, and repository invocations.
package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/secrecy"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

type UserRepository interface {
	InsertUser(ctx context.Context, email domain.Email, hash string) error
	GetUserByEmail(ctx context.Context, email domain.Email) (id domain.UserID, hash string, err error)
}

type JWTOptions struct {
	JWTSecret     secrecy.SecretString
	Exp           time.Duration
	SigningMethod jwt.SigningMethod
}

type Auth struct {
	repository UserRepository
	jwtOptions JWTOptions
}

func NewAuth(ur UserRepository, opts JWTOptions) *Auth {
	return &Auth{
		repository: ur,
		jwtOptions: opts,
	}
}

func (s *Auth) Register(ctx context.Context, email domain.Email, password domain.UserPassword) error {
	hash, err := argon2id.CreateHash(password.RevealSecret(), argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("auth service: register: hash password: %v: %w", err, ErrInternal)
	}

	err = s.repository.InsertUser(ctx, email, hash)
	if errors.Is(err, repository.ErrUniqueViolation) {
		return fmt.Errorf("auth service: register: user insert: %w", ErrUserAlreadyExists)
	}
	if err != nil {
		return fmt.Errorf("auth service: register: user insert: %v: %w", err, ErrInternal)
	}

	return nil
}

func (s *Auth) Login(ctx context.Context, email domain.Email, password domain.UserPassword) (domain.AuthToken, error) {
	id, hash, err := s.repository.GetUserByEmail(ctx, email)
	if errors.Is(err, repository.ErrRowNotFound) {
		return domain.AuthToken{}, fmt.Errorf("auth service: login: hash by email: %w", ErrUserNotFound)
	}
	if err != nil {
		return domain.AuthToken{}, fmt.Errorf("auth service: login: hash by email: %v: %w", err, ErrInternal)
	}

	isMatch, err := argon2id.ComparePasswordAndHash(password.RevealSecret(), hash)
	if err != nil {
		return domain.AuthToken{}, fmt.Errorf("auth service: login: compare password and hash: %v: %w", err, ErrInternal)
	}
	if !isMatch {
		return domain.AuthToken{}, ErrInvalidCredentials
	}

	token, err := s.CreateToken(id, s.jwtOptions.Exp)
	if err != nil {
		return domain.AuthToken{}, fmt.Errorf("auth service: login: create token: %v: %w", err, ErrInternal)
	}
	return token, nil
}

func (s *Auth) CreateToken(userID domain.UserID, exp time.Duration) (domain.AuthToken, error) {
	claims := jwt.MapClaims{
		"sub": userID.String(),
		"exp": time.Now().Add(exp).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(s.jwtOptions.SigningMethod, claims)

	tokenString, err := token.SignedString([]byte(s.jwtOptions.JWTSecret.RevealSecret()))
	if err != nil {
		return domain.AuthToken{}, err
	}

	jwtToken, err := domain.NewJWTString(tokenString)
	if err != nil {
		return domain.AuthToken{}, err
	}

	return jwtToken, nil
}

func (s *Auth) VerifyToken(ctx context.Context, token domain.AuthToken) (domain.UserID, error) {
	tokenString := token.RevealSecret()
	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if token.Method.Alg() != s.jwtOptions.SigningMethod.Alg() {
			return nil, ErrInvalidSigningMethod
		}
		return []byte(s.jwtOptions.JWTSecret.RevealSecret()), nil
	})
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return domain.UserID{}, fmt.Errorf("auth service: verify token: %w", ErrTokenExpired)
		}
		if errors.Is(err, ErrInvalidSigningMethod) {
			return domain.UserID{}, fmt.Errorf("auth service: verify token: %w", ErrInvalidSigningMethod)
		}
		return domain.UserID{}, fmt.Errorf("auth service: verify token: %v: %w", err, ErrInvalidToken)
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		idStr, ok := claims["sub"].(string)
		if !ok {
			return domain.UserID{}, fmt.Errorf("auth service: verify token: sub claim not found: %w", ErrInvalidToken)
		}

		id, err := domain.ParseUserID(idStr)
		if err != nil {
			return domain.UserID{}, fmt.Errorf("auth service: verify token: invalid id in sub: %v: %w", err, ErrInvalidToken)
		}
		return id, nil
	}

	return domain.UserID{}, fmt.Errorf("auth service: verify token: %w", ErrInvalidToken)
}
