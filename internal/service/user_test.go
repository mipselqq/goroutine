package service_test

import (
	"context"
	"errors"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestUser_CreateTelegramLinkToken(t *testing.T) {
	t.Parallel()

	wantUserID := testutil.ValidUserID()
	wantToken := testutil.ValidTelegramLinkToken()

	tests := []struct {
		name           string
		tokenFn        func() domain.TelegramLinkToken
		setupTokenRepo func(r *MockTelegramTokenRepository)
		wantErr        error
	}{
		{
			name: "Success",
			tokenFn: func() domain.TelegramLinkToken {
				return wantToken
			},
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.InsertLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, userID domain.UserID) error {
					if token != wantToken {
						t.Errorf("got token %v, want %v", token, wantToken)
					}
					if userID != wantUserID {
						t.Errorf("got userID %v, want %v", userID, wantUserID)
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name: "Internal error",
			tokenFn: func() domain.TelegramLinkToken {
				return wantToken
			},
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.InsertLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, uID domain.UserID) error {
					return repository.ErrInternal
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name: "Token already exists",
			tokenFn: func() domain.TelegramLinkToken {
				return wantToken
			},
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.InsertLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, uID domain.UserID) error {
					return repository.ErrKeyExists
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name: "Unexpected error",
			tokenFn: func() domain.TelegramLinkToken {
				return wantToken
			},
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.InsertLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken, uID domain.UserID) error {
					return errors.New("db exploded")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tr := &MockTelegramTokenRepository{}
			tt.setupTokenRepo(tr)
			s := service.NewUser(&MockUserRepository{}, tr, tt.tokenFn)

			gotToken, err := s.CreateTelegramLinkToken(context.Background(), wantUserID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}

			if tt.wantErr == nil && gotToken != wantToken {
				t.Errorf("got token %v, want %v", gotToken, wantToken)
			}
		})
	}
}

func TestUser_LinkTelegramByToken(t *testing.T) {
	t.Parallel()

	wantUserID := testutil.ValidUserID()
	wantToken := testutil.ValidTelegramLinkToken()
	wantChatID := testutil.ValidTelegramChatID()
	wantUsername := testutil.ValidTelegramUsername()

	tests := []struct {
		name           string
		setupTokenRepo func(r *MockTelegramTokenRepository)
		setupUserRepo  func(r *MockUserRepository)
		wantErr        error
	}{
		{
			name: "Success",
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.ConsumeTelegramLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
					if token != wantToken {
						t.Errorf("got token %v, want %v", token, wantToken)
					}
					return wantUserID, nil
				}
			},
			setupUserRepo: func(r *MockUserRepository) {
				r.UpdateTelegramInfoFunc = func(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					if userID != wantUserID {
						t.Errorf("got userID %v, want %v", userID, wantUserID)
					}
					if chatID != wantChatID {
						t.Errorf("got chatID %v, want %v", chatID, wantChatID)
					}
					if username != wantUsername {
						t.Errorf("got username %v, want %v", username, wantUsername)
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name: "Token not found",
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.ConsumeTelegramLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
					return domain.UserID{}, repository.ErrKeyNotFound
				}
			},
			setupUserRepo: func(r *MockUserRepository) {},
			wantErr:       service.ErrTelegramLinkTokenNotFound,
		},
		{
			name: "Consume internal error",
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.ConsumeTelegramLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
					return domain.UserID{}, repository.ErrInternal
				}
			},
			setupUserRepo: func(r *MockUserRepository) {},
			wantErr:       service.ErrInternal,
		},
		{
			name: "User not found",
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.ConsumeTelegramLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
					return wantUserID, nil
				}
			},
			setupUserRepo: func(r *MockUserRepository) {
				r.UpdateTelegramInfoFunc = func(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					return repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrUserNotFound,
		},
		{
			name: "Update internal error",
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.ConsumeTelegramLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
					return wantUserID, nil
				}
			},
			setupUserRepo: func(r *MockUserRepository) {
				r.UpdateTelegramInfoFunc = func(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					return repository.ErrInternal
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tokenRepo := &MockTelegramTokenRepository{}
			userRepo := &MockUserRepository{}

			tt.setupTokenRepo(tokenRepo)
			tt.setupUserRepo(userRepo)

			s := service.NewUser(userRepo, tokenRepo, nil)

			err := s.LinkTelegramByToken(context.Background(), wantToken, wantChatID, wantUsername)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}
