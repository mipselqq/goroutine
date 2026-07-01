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
			s := service.NewUser(tr, tt.tokenFn)

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
