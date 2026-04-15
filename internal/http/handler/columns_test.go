package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestColumns_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)

	tests := []struct {
		name         string
		path         string
		inputBody    any
		context      context.Context
		setupMock    func(s *MockColumns)
		expectedCode int
		expectedBody any
	}{
		{
			name:      "Success",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": validColumn.Name.String()},
			setupMock: func(s *MockColumns) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("expected caller id %v, got %v", validBoard.OwnerID, callerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, boardID)
					}
					if name != validColumn.Name {
						t.Errorf("expected name %v, got %v", validColumn.Name, name)
					}
					return validColumn, nil
				}
			},
			expectedCode: http.StatusCreated,
			expectedBody: map[string]any{
				"id":        validColumn.ID.String(),
				"boardId":   validColumn.BoardID.String(),
				"name":      validColumn.Name.String(),
				"position":  float64(validColumn.Position.Int64()),
				"createdAt": validColumn.CreatedAt.Format(timeFormat),
				"updatedAt": validColumn.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:         "Invalid board id",
			path:         "/v1/boards/not-a-uuid/columns",
			inputBody:    map[string]string{"name": "To Do"},
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:         "Invalid JSON",
			path:         "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody:    "{\"name\":\"broken\"",
			expectedCode: http.StatusBadRequest,
			expectedBody: invalidJsonBody(),
		},
		{
			name:         "Invalid name",
			path:         "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody:    map[string]string{"name": "   "},
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name:         "Missing context user",
			path:         "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody:    map[string]string{"name": "To Do"},
			context:      context.Background(),
			expectedCode: http.StatusUnauthorized,
			expectedBody: unauthorizedTokenBody(),
		},
		{
			name:      "Board not found",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": "To Do"},
			setupMock: func(s *MockColumns) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, service.ErrBoardNotFound
				}
			},
			expectedCode: http.StatusNotFound,
			expectedBody: boardNotFoundErrorBody(),
		},
		{
			name:      "Internal error",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": "To Do"},
			setupMock: func(s *MockColumns) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
		{
			name:      "Unexpected error",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": "To Do"},
			setupMock: func(s *MockColumns) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, errors.New("db exploded")
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var req *http.Request
			if raw, ok := tt.inputBody.(string); ok {
				req = httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = testutil.NewJSONRequestAndRecorder(t, http.MethodPost, tt.path, tt.inputBody)
			}

			if tt.context != nil {
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID))
			}
			req.SetPathValue("boardId", strings.TrimPrefix(strings.TrimSuffix(tt.path, "/columns"), "/v1/boards/"))

			rr := httptest.NewRecorder()
			mockColumns := &MockColumns{}
			if tt.setupMock != nil {
				tt.setupMock(mockColumns)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Create(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestColumns_List(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	first := testutil.ValidColumn(validBoard.ID)
	second := testutil.ValidColumn(validBoard.ID)
	second.Position, _ = domain.NewColumnPosition(first.Position.Int64() + 1)

	tests := []struct {
		name         string
		path         string
		context      context.Context
		setupMock    func(s *MockColumns)
		expectedCode int
		expectedBody any
	}{
		{
			name: "Success",
			path: "/v1/boards/" + validBoard.ID.String() + "/columns",
			setupMock: func(s *MockColumns) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("expected caller id %v, got %v", validBoard.OwnerID, callerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, boardID)
					}
					return []domain.Column{first, second}, nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: []map[string]any{
				{
					"id":        first.ID.String(),
					"boardId":   first.BoardID.String(),
					"name":      first.Name.String(),
					"position":  float64(first.Position.Int64()),
					"createdAt": first.CreatedAt.Format(timeFormat),
					"updatedAt": first.UpdatedAt.Format(timeFormat),
				},
				{
					"id":        second.ID.String(),
					"boardId":   second.BoardID.String(),
					"name":      second.Name.String(),
					"position":  float64(second.Position.Int64()),
					"createdAt": second.CreatedAt.Format(timeFormat),
					"updatedAt": second.UpdatedAt.Format(timeFormat),
				},
			},
		},
		{
			name:         "Invalid board id",
			path:         "/v1/boards/not-a-uuid/columns",
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:         "Missing context user",
			path:         "/v1/boards/" + validBoard.ID.String() + "/columns",
			context:      context.Background(),
			expectedCode: http.StatusUnauthorized,
			expectedBody: unauthorizedTokenBody(),
		},
		{
			name: "Board not found",
			path: "/v1/boards/" + validBoard.ID.String() + "/columns",
			setupMock: func(s *MockColumns) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
					return nil, service.ErrBoardNotFound
				}
			},
			expectedCode: http.StatusNotFound,
			expectedBody: boardNotFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: "/v1/boards/" + validBoard.ID.String() + "/columns",
			setupMock: func(s *MockColumns) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
					return nil, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
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
			req.SetPathValue("boardId", strings.TrimPrefix(strings.TrimSuffix(tt.path, "/columns"), "/v1/boards/"))

			rr := httptest.NewRecorder()
			mockColumns := &MockColumns{}
			if tt.setupMock != nil {
				tt.setupMock(mockColumns)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.List(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}

func TestColumns_UpdateByID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	updatedName, err := domain.NewColumnName("Renamed Column")
	if err != nil {
		t.Fatalf("NewColumnName: %v", err)
	}
	updatedColumn := validColumn
	updatedColumn.Name = updatedName
	updatedColumn.UpdatedAt = testutil.FixedTime5mFromNow()

	okPath := "/v1/boards/" + validBoard.ID.String() + "/columns/" + validColumn.ID.String()

	tests := []struct {
		name         string
		path         string
		inputBody    any
		context      context.Context
		setupMock    func(s *MockColumns)
		expectedCode int
		expectedBody any
	}{
		{
			name:      "Success (name update)",
			path:      okPath,
			inputBody: map[string]string{"name": updatedName.String()},
			// FIXME: setupMock receives main t, and not one from a subtest.
			// This is a critical bug across whole test suite that makes even failed subtests pass,
			// and the main one will fail with errors which origin is unclear (which subtest failed?)
			// All the setupMock blocks from app test suite must be rewritten to accept a subtest's t.
			// Created https://github.com/mipselqq/goroutine/issues/148
			setupMock: func(s *MockColumns) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("expected caller id %v, got %v", validBoard.OwnerID, callerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, boardID)
					}
					if columnID != validColumn.ID {
						t.Errorf("expected column id %v, got %v", validColumn.ID, columnID)
					}
					if name == nil || *name != updatedName {
						t.Errorf("expected updated name %v, got %+v", updatedName, name)
					}
					return updatedColumn, nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]any{
				"id":        updatedColumn.ID.String(),
				"boardId":   updatedColumn.BoardID.String(),
				"name":      updatedColumn.Name.String(),
				"position":  float64(updatedColumn.Position.Int64()),
				"createdAt": updatedColumn.CreatedAt.Format(timeFormat),
				"updatedAt": updatedColumn.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:      "Success (empty body no-op)",
			path:      okPath,
			inputBody: map[string]any{},
			setupMock: func(s *MockColumns) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
					if name != nil {
						t.Errorf("expected nil name for no-op patch, got %+v", name)
					}
					return validColumn, nil
				}
			},
			expectedCode: http.StatusOK,
			expectedBody: map[string]any{
				"id":        validColumn.ID.String(),
				"boardId":   validColumn.BoardID.String(),
				"name":      validColumn.Name.String(),
				"position":  float64(validColumn.Position.Int64()),
				"createdAt": validColumn.CreatedAt.Format(timeFormat),
				"updatedAt": validColumn.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:         "Invalid board id",
			path:         "/v1/boards/not-a-uuid/columns/" + validColumn.ID.String(),
			inputBody:    map[string]string{"name": "Renamed"},
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:         "Invalid column id",
			path:         "/v1/boards/" + validBoard.ID.String() + "/columns/not-a-uuid",
			inputBody:    map[string]string{"name": "Renamed"},
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:         "Invalid JSON",
			path:         okPath,
			inputBody:    "{\"name\":\"broken\"",
			expectedCode: http.StatusBadRequest,
			expectedBody: invalidJsonBody(),
		},
		{
			name:         "Invalid name",
			path:         okPath,
			inputBody:    map[string]string{"name": "   "},
			expectedCode: http.StatusBadRequest,
			expectedBody: validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name:         "Missing context user",
			path:         okPath,
			inputBody:    map[string]string{"name": "Renamed"},
			context:      context.Background(),
			expectedCode: http.StatusUnauthorized,
			expectedBody: unauthorizedTokenBody(),
		},
		{
			name:      "Column not found",
			path:      okPath,
			inputBody: map[string]string{"name": "Renamed"},
			setupMock: func(s *MockColumns) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, service.ErrColumnNotFound
				}
			},
			expectedCode: http.StatusNotFound,
			expectedBody: columnNotFoundErrorBody(),
		},
		{
			name:      "Internal error",
			path:      okPath,
			inputBody: map[string]string{"name": "Renamed"},
			setupMock: func(s *MockColumns) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, service.ErrInternal
				}
			},
			expectedCode: http.StatusInternalServerError,
			expectedBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var req *http.Request
			if raw, ok := tt.inputBody.(string); ok {
				req = httptest.NewRequest(http.MethodPatch, tt.path, strings.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = testutil.NewJSONRequestAndRecorder(t, http.MethodPatch, tt.path, tt.inputBody)
			}

			if tt.context != nil {
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID))
			}

			boardAndColumn := strings.TrimPrefix(tt.path, "/v1/boards/")
			parts := strings.Split(boardAndColumn, "/columns/")
			if len(parts) == 2 {
				req.SetPathValue("boardId", parts[0])
				req.SetPathValue("columnId", parts[1])
			}

			rr := httptest.NewRecorder()
			mockColumns := &MockColumns{}
			if tt.setupMock != nil {
				tt.setupMock(mockColumns)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.UpdateByID(rr, req)

			testutil.AssertStatusCode(t, rr, tt.expectedCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.expectedBody)
		})
	}
}
