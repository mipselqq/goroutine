package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/service"
	"goroutine/internal/template"
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

			tr := NewMockTelegramTokenRepository(t)
			tt.setupTokenRepo(tr)
			s := service.NewUser(NewMockUserRepository(t), tr, nil, tt.tokenFn)

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
		setupNotif     func(n *MockTelegramLinkNotif)
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
			setupNotif: func(n *MockTelegramLinkNotif) {
				n.NotifChatFunc = func(ctx context.Context, chatID domain.TelegramChatID, notification fmt.Stringer) error {
					if chatID != wantChatID {
						t.Errorf("got chatID %v, want %v", chatID, wantChatID)
					}
					wantNotif := template.TelegramLinkedNotif{}
					if notification != wantNotif {
						t.Errorf("got notification %T, want %T", notification, wantNotif)
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
			setupNotif: func(n *MockTelegramLinkNotif) {
				n.NotifChatFunc = func(ctx context.Context, chatID domain.TelegramChatID, notification fmt.Stringer) error {
					wantNotif := template.TelegramLinkTokenExpiredNotif{}
					if notification != wantNotif {
						t.Errorf("got notification %T, want %T", notification, wantNotif)
					}
					return nil
				}
			},
			wantErr: service.ErrTelegramLinkTokenNotFound,
		},
		{
			name: "Consume internal error",
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.ConsumeTelegramLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
					return domain.UserID{}, repository.ErrInternal
				}
			},
			setupUserRepo: func(r *MockUserRepository) {},
			setupNotif: func(n *MockTelegramLinkNotif) {
				n.NotifChatFunc = func(ctx context.Context, chatID domain.TelegramChatID, notification fmt.Stringer) error {
					wantNotif := template.TelegramLinkFailedNotif{}
					if notification != wantNotif {
						t.Errorf("got notification %T, want %T", notification, wantNotif)
					}
					return nil
				}
			},
			wantErr: service.ErrInternal,
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
			setupNotif: func(n *MockTelegramLinkNotif) {
				n.NotifChatFunc = func(ctx context.Context, chatID domain.TelegramChatID, notification fmt.Stringer) error {
					wantNotif := template.TelegramUserNotFoundNotif{}
					if notification != wantNotif {
						t.Errorf("got notification %T, want %T", notification, wantNotif)
					}
					return nil
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
			setupNotif: func(n *MockTelegramLinkNotif) {
				n.NotifChatFunc = func(ctx context.Context, chatID domain.TelegramChatID, notification fmt.Stringer) error {
					wantNotif := template.TelegramLinkFailedNotif{}
					if notification != wantNotif {
						t.Errorf("got notification %T, want %T", notification, wantNotif)
					}
					return nil
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name: "Notif error",
			setupTokenRepo: func(r *MockTelegramTokenRepository) {
				r.ConsumeTelegramLinkTokenFunc = func(ctx context.Context, token domain.TelegramLinkToken) (domain.UserID, error) {
					return wantUserID, nil
				}
			},
			setupUserRepo: func(r *MockUserRepository) {
				r.UpdateTelegramInfoFunc = func(ctx context.Context, userID domain.UserID, chatID domain.TelegramChatID, username domain.TelegramUsername) error {
					return nil
				}
			},
			setupNotif: func(n *MockTelegramLinkNotif) {
				n.NotifChatFunc = func(ctx context.Context, chatID domain.TelegramChatID, notification fmt.Stringer) error {
					wantNotif := template.TelegramLinkedNotif{}
					if notification != wantNotif {
						t.Errorf("got notification %T, want %T", notification, wantNotif)
					}
					return errors.New("telegram unavailable")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tokenRepo := NewMockTelegramTokenRepository(t)
			userRepo := NewMockUserRepository(t)
			notifService := MockTelegramLinkNotif{}

			tt.setupTokenRepo(tokenRepo)
			tt.setupUserRepo(userRepo)
			tt.setupNotif(&notifService)

			s := service.NewUser(userRepo, tokenRepo, &notifService, nil)

			err := s.LinkTelegramByToken(context.Background(), wantToken, wantChatID, wantUsername)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if notifService.NotifChatCalls != 1 {
				t.Errorf("got %d notifications, want 1", notifService.NotifChatCalls)
			}
		})
	}
}
