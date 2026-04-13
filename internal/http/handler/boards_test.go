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
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))

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
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))

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
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Get(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestBoards_UpdateByID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	updatedValidBoard := testutil.UpdateValidBoard(t, &validBoard, "Updated Board Name", "Updated Board Description", testutil.FixedTime5mFromNow())
	updatedNameOnlyBoard := testutil.UpdateValidBoard(t, &validBoard, "Updated Board Name Only", validBoard.Description.String(), testutil.FixedTime5mFromNow())
	updatedDescriptionOnlyBoard := testutil.UpdateValidBoard(t, &validBoard, validBoard.Name.String(), "Updated Board Description Only", testutil.FixedTime5mFromNow())
	emptyDescriptionBoard := testutil.UpdateValidBoard(t, &validBoard, "Updated Board Name", "", testutil.FixedTime5mFromNow())

	okPath := "/v1/boards/" + validBoard.ID.String()

	tests := []boardsTestCase{
		{
			name: "Success (full update)",
			path: okPath,
			inputBody: map[string]string{
				"name":        updatedValidBoard.Name.String(),
				"description": updatedValidBoard.Description.String(),
			},
			setupMock: func(s *MockBoards) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if ownerID != validBoard.OwnerID || boardID != validBoard.ID {
						t.Errorf("unexpected ownerID %v boardID %v", ownerID, boardID)
					}
					if name == nil || *name != updatedValidBoard.Name {
						t.Errorf("unexpected name %+v", name)
					}
					if description == nil || *description != updatedValidBoard.Description {
						t.Errorf("unexpected description %+v", description)
					}
					return updatedValidBoard, nil
				}
			},
			expectedBody: map[string]string{
				"id":          updatedValidBoard.ID.String(),
				"ownerId":     updatedValidBoard.OwnerID.String(),
				"name":        updatedValidBoard.Name.String(),
				"description": updatedValidBoard.Description.String(),
				"createdAt":   updatedValidBoard.CreatedAt.Format(time.RFC3339),
				"updatedAt":   updatedValidBoard.UpdatedAt.Format(time.RFC3339),
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Empty name",
			path: okPath,
			inputBody: map[string]string{
				"name":        "",
				"description": validBoard.Description.String(),
			},
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name: "Success (name only)",
			path: okPath,
			inputBody: map[string]string{
				"name": updatedNameOnlyBoard.Name.String(),
			},
			setupMock: func(s *MockBoards) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name == nil || *name != updatedNameOnlyBoard.Name {
						t.Errorf("unexpected name %+v", name)
					}
					if description != nil {
						t.Errorf("expected nil description, got %+v", description)
					}
					return updatedNameOnlyBoard, nil
				}
			},
			expectedBody: map[string]any{
				"id":          updatedNameOnlyBoard.ID.String(),
				"ownerId":     updatedNameOnlyBoard.OwnerID.String(),
				"name":        updatedNameOnlyBoard.Name.String(),
				"description": updatedNameOnlyBoard.Description.String(),
				"createdAt":   updatedNameOnlyBoard.CreatedAt.Format(time.RFC3339),
				"updatedAt":   updatedNameOnlyBoard.UpdatedAt.Format(time.RFC3339),
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Success (description only)",
			path: okPath,
			inputBody: map[string]string{
				"description": updatedDescriptionOnlyBoard.Description.String(),
			},
			setupMock: func(s *MockBoards) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name != nil {
						t.Errorf("expected nil name, got %+v", name)
					}
					if description == nil || *description != updatedDescriptionOnlyBoard.Description {
						t.Errorf("unexpected description %+v", description)
					}
					return updatedDescriptionOnlyBoard, nil
				}
			},
			expectedBody: map[string]any{
				"id":          updatedDescriptionOnlyBoard.ID.String(),
				"ownerId":     updatedDescriptionOnlyBoard.OwnerID.String(),
				"name":        updatedDescriptionOnlyBoard.Name.String(),
				"description": updatedDescriptionOnlyBoard.Description.String(),
				"createdAt":   updatedDescriptionOnlyBoard.CreatedAt.Format(time.RFC3339),
				"updatedAt":   updatedDescriptionOnlyBoard.UpdatedAt.Format(time.RFC3339),
			},
			expectedCode: http.StatusOK,
		},
		{
			name:      "Success (null fields mean skip)",
			path:      okPath,
			inputBody: map[string]any{"name": nil, "description": nil},
			setupMock: func(s *MockBoards) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name != nil || description != nil {
						t.Errorf("expected nil pointers, got name=%+v description=%+v", name, description)
					}
					return validBoard, nil
				}
			},
			expectedBody: map[string]any{
				"id":          validBoard.ID.String(),
				"ownerId":     validBoard.OwnerID.String(),
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
				"createdAt":   validBoard.CreatedAt.Format(time.RFC3339),
				"updatedAt":   validBoard.UpdatedAt.Format(time.RFC3339),
			},
			expectedCode: http.StatusOK,
		},
		{
			name: "Empty description",
			path: okPath,
			inputBody: map[string]string{
				"name":        emptyDescriptionBoard.Name.String(),
				"description": "",
			},
			setupMock: func(s *MockBoards) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name == nil || description == nil {
						t.Errorf("expected non-nil pointers")
					}
					return emptyDescriptionBoard, nil
				}
			},
			expectedBody: map[string]any{
				"id":          emptyDescriptionBoard.ID.String(),
				"ownerId":     emptyDescriptionBoard.OwnerID.String(),
				"name":        emptyDescriptionBoard.Name.String(),
				"description": emptyDescriptionBoard.Description.String(),
				"createdAt":   emptyDescriptionBoard.CreatedAt.Format(time.RFC3339),
				"updatedAt":   emptyDescriptionBoard.UpdatedAt.Format(time.RFC3339),
			},
			expectedCode: http.StatusOK,
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
			inputBody: map[string]string{
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
			},
			setupMock: func(s *MockBoards) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, service.ErrBoardNotFound
				}
			},
			expectedCode: http.StatusNotFound,
			expectedBody: notFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: okPath,
			inputBody: map[string]string{
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
			},
			setupMock: func(s *MockBoards) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name: "Unknown error",
			path: okPath,
			inputBody: map[string]string{
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
			},
			setupMock: func(s *MockBoards) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("unknown")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name: "No context user ID",
			path: okPath,
			inputBody: map[string]string{
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
			},
			context:      context.Background(),
			expectedCode: http.StatusUnauthorized,
			expectedBody: unauthorizedTokenBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, rr := testutil.NewJSONRequestAndRecorder(t, http.MethodPatch, tt.path, tt.inputBody)
			if tt.context != nil {
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID))
			}
			req.SetPathValue("boardId", strings.TrimPrefix(tt.path, "/v1/boards/"))

			s := &MockBoards{}
			if tt.setupMock != nil {
				tt.setupMock(s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.UpdateByID(rr, req)

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
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Delete(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			if tt.expectedCode != http.StatusNoContent {
				testutil.AssertContentType(t, rr, "application/json")
			}
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}
