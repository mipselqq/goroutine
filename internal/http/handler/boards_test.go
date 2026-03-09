package handler_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
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
	setupMock    func(s *MockBoards)
	expectedCode int
	expectedBody string
}

const (
	name        string = "My Todo Name"
	description string = "My Todo Description"
)

func TestBoards_Create(t *testing.T) {
	t.Parallel()

	name, _ := domain.NewBoardName(name)
	description, _ := domain.NewBoardDescription(description)
	id := domain.NewBoardID()
	userID := domain.NewUserID()

	validBoard := domain.Board{
		ID:          id,
		OwnerID:     userID,
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
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
			expectedCode: http.StatusOK,
			expectedBody: fmt.Sprintf(
				`{"id":%q,"ownerId":%q,"name":%q,"description":%q,"createdAt":%q,"updatedAt":%q}`,
				id.String(),
				userID.String(),
				name.String(),
				description.String(),
				validBoard.CreatedAt.Format(time.RFC3339),
				validBoard.UpdatedAt.Format(time.RFC3339),
			),
		},
		{
			name:         "Empty name",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q}`, "", description),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"name","issues":["Name is too short"]}]}`, FixedTime),
		},
		{
			name:         "Name too short",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q}`, "a", description),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"name","issues":["Name is too short"]}]}`, FixedTime),
		},
		{
			name:      "Empty description",
			inputBody: fmt.Sprintf(`{"name": %q, "description": %q}`, name, ""),
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{
						ID:          id,
						OwnerID:     ownerID,
						Name:        name,
						Description: description,
						CreatedAt:   validBoard.CreatedAt,
						UpdatedAt:   validBoard.UpdatedAt,
					}, nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: fmt.Sprintf(
				`{"id":%q,"ownerId":%q,"name":%q,"description":%q,"createdAt":%q,"updatedAt":%q}`,
				id.String(),
				userID.String(),
				name.String(),
				"",
				validBoard.CreatedAt.Format(time.RFC3339),
				validBoard.UpdatedAt.Format(time.RFC3339),
			),
		},
		{
			name:         "Description too long",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q}`, name, strings.Repeat("a", 1025)),
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"description","issues":["Description is too long"]}]}`, FixedTime),
		},
		{
			name:         "Invalid JSON",
			inputBody:    fmt.Sprintf(`{"name": %q, "description": %q`, name, description), // missing closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: fmt.Sprintf(`{"code":"VALIDATION_ERROR","message":"Some fields are invalid","timestamp":%q,"details":[{"field":"body","issues":["Invalid JSON body"]}]}`, FixedTime),
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
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, FixedTime),
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
			expectedBody: fmt.Sprintf(`{"code":"INTERNAL_SERVER_ERROR","message":"Internal server error","timestamp":%q}`, FixedTime),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodPost, "/v1/boards", bytes.NewBuffer([]byte(tt.inputBody)))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), httpschema.ContextKeyUserID, userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()
			s := &MockBoards{}

			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, MockTime))
			h.Create(rr, req)

			if rr.Code != tt.expectedCode {
				t.Errorf("expected status %d, got %d", tt.expectedCode, rr.Code)
			}

			contentType := rr.Header().Get("Content-Type")
			mediaType, _, err := mime.ParseMediaType(contentType)
			if err != nil {
				t.Fatalf("Failed to parse MIME %q", contentType)
			}
			if mediaType != ExpectedMime {
				t.Errorf("Expected %q, got %q", ExpectedMime, mediaType)
			}

			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestBoards_Create_NoContextUserID(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/boards", bytes.NewBuffer([]byte(`{"name": "My Todo Name", "description": "My Todo Description"}`)))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	s := &MockBoards{}
	logger := testutil.NewTestLogger(t)
	h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, MockTime))
	h.Create(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rr.Code)
	}

	// TODO: factor out to helpers_test/AssertContentType
	contentType := rr.Header().Get("Content-Type")
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		t.Fatalf("Failed to parse MIME %q", contentType)
	}
	if mediaType != ExpectedMime {
		t.Errorf("Expected %q, got %q", ExpectedMime, mediaType)
	}
}
