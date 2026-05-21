package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

type userTestCase struct {
	name             string
	inputBody        any
	context          context.Context
	setupAuthService func(t *testing.T, s *MockAuthService)
	wantCode         int
	wantBody         any
}

func TestUser_CreateTelegramLinkToken(t *testing.T) {
	t.Parallel()

	authorizedUserID := testutil.ValidUserID()

	tests := []userTestCase{
		{
			name:      "Success",
			inputBody: testutil.Big25KBJson(), // Body is ignored, no error on big payload
			setupAuthService: func(t *testing.T, s *MockAuthService) {
				s.CreateTelegramLinkTokenFunc = func(ctx context.Context, callerID domain.UserID) (domain.TelegramLinkToken, error) {
					if callerID != authorizedUserID {
						t.Errorf("got service call user ID %q, want %q as in context", authorizedUserID, callerID)
					}
					return testutil.ValidTelegramLinkToken(), nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]string{"token": testutil.ValidTelegramLinkToken().RevealSecret()},
		},
		{
			name:     "Missing context user ID",
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
		},
		{
			name:      "Internal error",
			inputBody: nil,
			setupAuthService: func(t *testing.T, s *MockAuthService) {
				s.CreateTelegramLinkTokenFunc = func(ctx context.Context, userID domain.UserID) (domain.TelegramLinkToken, error) {
					return domain.TelegramLinkToken{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name:      "Unexpected error",
			inputBody: nil,
			setupAuthService: func(t *testing.T, s *MockAuthService) {
				s.CreateTelegramLinkTokenFunc = func(ctx context.Context, userID domain.UserID) (domain.TelegramLinkToken, error) {
					return domain.TelegramLinkToken{}, errors.New("storage crash")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/v1/users/me/telegram/link", tt.inputBody)
			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, testutil.ValidUserID())
			}
			req = req.WithContext(ctx)

			mockAuth := &MockAuthService{}
			if tt.setupAuthService != nil {
				tt.setupAuthService(t, mockAuth)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewUser(logger, mockAuth, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.CreateTelegramLinkToken(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}
