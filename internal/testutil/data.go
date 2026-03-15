package testutil

import (
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/secrecy"
	"goroutine/internal/service"

	"github.com/golang-jwt/jwt/v5"
)

func FixedTime() string { return "2026-01-01T00:00:00Z" }

func ParseUserID(s string) domain.UserID {
	u, err := domain.ParseUserID(s)
	if err != nil {
		panic(err)
	}
	return u
}

func ValidUserID() domain.UserID {
	return ParseUserID("018e1000-0000-7000-8000-000000000000")
}

func ValidEmail() domain.Email {
	e, _ := domain.NewEmail("test@example.com")
	return e
}

func ValidPassword() domain.Password {
	p, _ := domain.NewPassword("qwerty")
	return p
}

func ValidPasswordHash() string {
	return "$argon2id$v=19$m=65536,t=1,p=16$kUYJyX3h53cARKnKqFZxvQ$IXz2KOKbyVklgyVmz9ebJ1ffOgmcyMpn/GTUWsep5lk"
}

func AnotherValidPasswordHash() string {
	return "$argon2id$v=19$m=65536,t=3,p=4$bm90LXF3ZXJ0eQ$fSowp1Rof0fXhF+rXv2f6w"
}

func ValidBoardName() domain.BoardName {
	n, _ := domain.NewBoardName("Test Board")
	return n
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
