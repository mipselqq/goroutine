package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	inputBody    any
	context      context.Context
	setupMock    func(s *MockBoards)
	expectedCode int
	expectedBody any
	path         string
}

func TestBoards_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []boardsTestCase{
		{
			name:      "Success",
			inputBody: map[string]string{"name": validBoard.Name.String(), "description": validBoard.Description.String()},
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("expected ownerID %v, got %v", validBoard.OwnerID, ownerID)
					}
					return validBoard, nil
				}
			},
			expectedCode: http.StatusCreated,
			expectedBody: map[string]string{
				"id":          validBoard.ID.String(),
				"ownerId":     validBoard.OwnerID.String(),
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
				"createdAt":   validBoard.CreatedAt.Format(time.RFC3339),
				"updatedAt":   validBoard.UpdatedAt.Format(time.RFC3339),
			},
		},
		{
			name:         "Empty name",
			inputBody:    map[string]string{"name": "", "description": validBoard.Description.String()},
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name:         "Description too long",
			inputBody:    map[string]string{"name": validBoard.Name.String(), "description": strings.Repeat("a", 1025)},
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("description", []string{"Description is too long"}),
		},
		{
			name:         "Invalid JSON",
			inputBody:    json.RawMessage([]byte(fmt.Sprintf(`{"name": %q, "description": %q`, validBoard.Name.String(), validBoard.Description.String()))), // missing closing brace
			expectedCode: http.StatusBadRequest,
			expectedBody: invalidJsonBody(),
		},
		{
			name:         "No context user ID",
			inputBody:    map[string]string{"name": validBoard.Name.String(), "description": validBoard.Description.String()},
			context:      context.Background(),
			expectedCode: http.StatusUnauthorized,
			expectedBody: unauthorizedTokenBody(),
		},
		{
			name:      "Internal error",
			inputBody: map[string]string{"name": validBoard.Name.String(), "description": validBoard.Description.String()},
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name:      "Unknown error",
			inputBody: map[string]string{"name": validBoard.Name.String(), "description": validBoard.Description.String()},
			setupMock: func(s *MockBoards) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("unknown error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPost, "/v1/boards", tt.inputBody)

			if tt.context != nil {
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID))
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

func TestBoards_Get(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []boardsTestCase{
		{
			name:      "Success",
			inputBody: "",
			setupMock: func(s *MockBoards) {
				s.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("expected ownerID %v, got %v", validBoard.OwnerID, ownerID)
					}

					return []domain.Board{validBoard}, nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: []map[string]string{
				{
					"id":          validBoard.ID.String(),
					"ownerId":     validBoard.OwnerID.String(),
					"name":        validBoard.Name.String(),
					"description": validBoard.Description.String(),
					"createdAt":   validBoard.CreatedAt.Format(time.RFC3339),
					"updatedAt":   validBoard.UpdatedAt.Format(time.RFC3339),
				},
			},
		},
		{
			name:         "No context user ID",
			inputBody:    "",
			context:      context.Background(),
			expectedCode: http.StatusUnauthorized,
			expectedBody: unauthorizedTokenBody(),
		},
		{
			name:      "Internal error",
			inputBody: "",
			setupMock: func(s *MockBoards) {
				s.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					return nil, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name:      "Unknown error",
			inputBody: "",
			setupMock: func(s *MockBoards) {
				s.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					return nil, errors.New("unknown error")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodGet, "/v1/boards", tt.inputBody)

			if tt.context != nil {
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID))
			}

			s := &MockBoards{}

			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTime))

			h.GetMany(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestBoards_GetByID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	okPath := "/v1/boards/" + validBoard.ID.String()

	tests := []boardsTestCase{
		{
			name: "Success",
			path: okPath,
			setupMock: func(s *MockBoards) {
				s.GetFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
					if ownerID != validBoard.OwnerID || boardID != validBoard.ID {
						t.Errorf("unexpected ownerID %v boardID %v", ownerID, boardID)
					}
					return validBoard, nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]string{
				"id":          validBoard.ID.String(),
				"ownerId":     validBoard.OwnerID.String(),
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
				"createdAt":   validBoard.CreatedAt.Format(time.RFC3339),
				"updatedAt":   validBoard.UpdatedAt.Format(time.RFC3339),
			},
		},
		{
			name:         "Invalid board id",
			path:         "/v1/boards/not-a-uuid",
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name: "Not found",
			path: okPath,
			setupMock: func(s *MockBoards) {
				s.GetFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
					return domain.Board{}, service.ErrBoardNotFound
				}
			},
			expectedCode: http.StatusNotFound,
			expectedBody: notFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: okPath,
			setupMock: func(s *MockBoards) {
				s.GetFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
					return domain.Board{}, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name: "Unknown error",
			path: okPath,
			setupMock: func(s *MockBoards) {
				s.GetFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
					return domain.Board{}, errors.New("unknown")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name:         "No context user ID",
			path:         okPath,
			context:      context.Background(),
			expectedCode: http.StatusUnauthorized,
			expectedBody: unauthorizedTokenBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			if tt.context != nil {
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID))
			}
			req.SetPathValue("boardId", strings.TrimPrefix(tt.path, "/v1/boards/"))

			rr := httptest.NewRecorder()

			s := &MockBoards{}
			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTime))
			h.Get(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestBoards_Delete(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	okPath := "/v1/boards/" + validBoard.ID.String()

	tests := []boardsTestCase{
		{
			name: "Success",
			path: okPath,
			setupMock: func(s *MockBoards) {
				s.DeleteFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
					if ownerID != validBoard.OwnerID || boardID != validBoard.ID {
						t.Errorf("unexpected ownerID %v boardID %v", ownerID, boardID)
					}
					return nil
				}
			},
			expectedCode: http.StatusNoContent,
			expectedBody: nil,
		},
		{
			name:         "Invalid board id",
			path:         "/v1/boards/not-a-uuid",
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name: "Not found",
			path: okPath,
			setupMock: func(s *MockBoards) {
				s.DeleteFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
					return service.ErrBoardNotFound
				}
			},
			expectedCode: http.StatusNotFound,
			expectedBody: notFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: okPath,
			setupMock: func(s *MockBoards) {
				s.DeleteFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
					return service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name: "Unknown error",
			path: okPath,
			setupMock: func(s *MockBoards) {
				s.DeleteFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
					return errors.New("unknown")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name:         "No context user ID",
			path:         okPath,
			context:      context.Background(),
			expectedCode: http.StatusUnauthorized,
			expectedBody: unauthorizedTokenBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodDelete, tt.path, http.NoBody)
			if tt.context != nil {
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID))
			}
			req.SetPathValue("boardId", strings.TrimPrefix(tt.path, "/v1/boards/"))

			rr := httptest.NewRecorder()

			s := &MockBoards{}
			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTime))
			h.Delete(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			if tt.expectedCode != http.StatusNoContent {
				testutil.AssertContentType(t, rr, "application/json")
			}
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}
