package service_test

import (
	"context"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/secrecy"
	"goroutine/internal/service"
	"goroutine/internal/testutil"

	"github.com/golang-jwt/jwt/v5"
)

// TODO: user functions to create valid values for tests
// to avoid potential global variables pollution
var (
	validEmailStr            = "test@example.com"
	email, _                 = domain.NewEmail(validEmailStr)
	userID                   = testutil.ParseUserID("018e1000-0000-7000-8000-000000000000")
	validPasswordStr         = "qwerty"
	password, _              = domain.NewPassword(validPasswordStr)
	validPasswordHash        = "$argon2id$v=19$m=65536,t=1,p=16$kUYJyX3h53cARKnKqFZxvQ$IXz2KOKbyVklgyVmz9ebJ1ffOgmcyMpn/GTUWsep5lk"
	validAnotherPasswordHash = "$argon2id$v=19$m=65536,t=3,p=4$bm90LXF3ZXJ0eQ$fSowp1Rof0fXhF+rXv2f6w"
	JWTSecret                = secrecy.SecretString("secret")
	JWTOpts                  = service.JWTOptions{
		JWTSecret:     JWTSecret,
		Exp:           time.Hour,
		SigningMethod: jwt.SigningMethodHS256,
	}

	validBoardNameStr        = "Test Board"
	boardName, _             = domain.NewBoardName(validBoardNameStr)
	validBoardDescriptionStr = "Test Board Description"
	boardDescription, _      = domain.NewBoardDescription(validBoardDescriptionStr)
)

type MockUserRepository struct {
	InsertFunc     func(ctx context.Context, email domain.Email, hash string) error
	GetByEmailFunc func(ctx context.Context, email domain.Email) (id domain.UserID, hash string, err error)
}

func (m *MockUserRepository) Insert(ctx context.Context, email domain.Email, hash string) error {
	return m.InsertFunc(ctx, email, hash)
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email domain.Email) (id domain.UserID, hash string, err error) {
	return m.GetByEmailFunc(ctx, email)
}

type MockBoardRepository struct {
	CreateFunc func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) error
}

func (m *MockBoardRepository) Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) error {
	return m.CreateFunc(ctx, ownerID, name, description)
}
