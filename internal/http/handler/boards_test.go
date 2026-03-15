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
	inputBody    any
	context      context.Context
	setupMock    func(s *MockBoards)
	expectedCode int
	expectedBody any
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
			inputBody: map[string]string{"name": name.String(), "description": description.String()},
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					if ownerID != userID {
						t.Errorf("expected ownerID %v, got %v", userID, ownerID)
					}
					return validBoard, nil
				}
			},
			expectedCode: http.StatusCreated,
			expectedBody: map[string]string{
				"id":          id.String(),
				"ownerId":     userID.String(),
				"name":        name.String(),
				"description": description.String(),
				"createdAt":   validBoard.CreatedAt.Format(time.RFC3339),
			},
		},
		{
			name:         "Empty name",
			inputBody:    map[string]string{"name": "", "description": description.String()},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "name", "issues": []string{"Name is too short"}},
				},
			},
		},
		{
			name:         "Description too long",
			inputBody:    map[string]string{"name": name.String(), "description": strings.Repeat("a", 1025)},
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "description", "issues": []string{"Description is too long"}},
				},
			},
		},
		{
			name:         "Invalid JSON",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q`, name.String(), description.String()), // missing closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: map[string]any{
				"code":      "VALIDATION_ERROR",
				"message":   "Some fields are invalid",
				"timestamp": testutil.FixedTime(),
				"details": []any{
					map[string]any{"field": "body", "issues": []string{"Invalid JSON body"}},
				},
			},
		},
		{
			name:      "Internal error",
			inputBody: map[string]string{"name": name.String(), "description": description.String()},
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedTime(),
			},
		},
		{
			name:      "Unknown error",
			inputBody: map[string]string{"name": name.String(), "description": description.String()},
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("unknown error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedTime(),
			},
		},
		{
			name:         "No context user ID",
			inputBody:    map[string]string{"name": name.String(), "description": description.String()},
			context:      context.Background(),
			expectedCode: http.StatusInternalServerError,
			expectedBody: map[string]any{
				"code":      "INTERNAL_SERVER_ERROR",
				"message":   "Internal server error",
				"timestamp": testutil.FixedTime(),
			},
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

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}
