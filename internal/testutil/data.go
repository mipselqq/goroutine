package testutil

import (
	"fmt"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/secrecy"
	"goroutine/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

func FixedTime() string { return "2026-01-01T00:00:00Z" }

func must[T any](fn func(string) (T, error), s string) T {
	v, err := fn(s)
	if err != nil {
		panic(fmt.Errorf("testutil: BUG: value is no longer valid: %w", err))
	}
	return v
}

func ValidUserID() domain.UserID {
	return must(domain.ParseUserID, "018e1000-0000-7000-8000-000000000000")
}

func ValidEmail() domain.Email {
	return must(domain.NewEmail, "test@example.com")
}

func ValidPassword() domain.Password {
	return must(domain.NewPassword, "qwerty")
}

func ValidPasswordHash() string {
	return "$argon2id$v=19$m=65536,t=1,p=16$kUYJyX3h53cARKnKqFZxvQ$IXz2KOKbyVklgyVmz9ebJ1ffOgmcyMpn/GTUWsep5lk"
}

func AnotherValidPasswordHash() string {
	return "$argon2id$v=19$m=65536,t=3,p=4$bm90LXF3ZXJ0eQ$fSowp1Rof0fXhF+rXv2f6w"
}

func ValidBoardName() domain.BoardName {
	return must(domain.NewBoardName, "Test Board")
}

func ValidBoardDescription() domain.BoardDescription {
	d, _ := domain.NewBoardDescription("Test Board Description")
	return d
}

func ValidJWTSecret() secrecy.SecretString {
	return secrecy.SecretString("secret")
}

func ValidJWTOptions() service.JWTOptions {
	return service.JWTOptions{
		JWTSecret:     ValidJWTSecret(),
		Exp:           time.Hour,
		SigningMethod: jwt.SigningMethodHS256,
	}
}
