package handler_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

type boardsTestCase struct {
	name         string
	inputBody    string
	context      context.Context
	setupMock    func(s *MockBoards)
	expectedCode int
	expectedBody string
}

func TestBoards_Create(t *testing.T) {
	t.Parallel()

	name := testutil.ValidBoardName()
	description := testutil.ValidBoardDescription()
	id := domain.NewBoardID()
	userID := testutil.ValidUserID()

	validBoard := domain.Board{
		ID:          id,
		OwnerID:     userID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}

	tests := []boardsTestCase{
		{
			name:      "Success",
			inputBody: fmt.Sprintf(`{"name": %q, "description": %q}`, name, description),
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					if ownerID != userID {
						t.Errorf("expected ownerID %v, got %v", userID, ownerID)
					}
					return validBoard, nil
				}
			},
			expectedCode: http.StatusCreated,
			expectedBody: fmt.Sprintf(
				`{"id":%q,"ownerId":%q,"name":%q,"description":%q,"createdAt":%q}`,
				id.String(),
				userID.String(),
				name.String(),
				description.String(),
				validBoard.CreatedAt.Format(time.RFC3339),
			),
		},
		{
			name:         "Empty name",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q}`, "", description),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"name","issues":["Name is too short"]}]}`, testutil.FixedTime()),
		},
		{
			name:         "Description too long",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q}`, name, strings.Repeat("a", 1025)),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"description","issues":["Description is too long"]}]}`, testutil.FixedTime()),
		},
		{
			name:         "Invalid JSON",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q`, name, description), // missing closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"body","issues":["Invalid JSON body"]}]}`, testutil.FixedTime()),
		},
		{
			name:      "Internal error",
			inputBody: fmt.Sprintf(`{"name": %q, "description": %q}`, name, description),
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, testutil.FixedTime()),
		},
		{
			name:      "Unknown error",
			inputBody: fmt.Sprintf(`{"name": %q, "description": %q}`, name, description),
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("unknown error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, testutil.FixedTime()),
		},
		{
			name:         "No context user ID",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q}`, name, description),
			context:      context.Background(),
			expectedCode: http.StatusInternalServerError,
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, testutil.FixedTime()),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/v1/boards", tt.inputBody)

			if tt.context != nil {
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, userID))
			}

			s := &MockBoards{}

			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTime))
			h.Create(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}
