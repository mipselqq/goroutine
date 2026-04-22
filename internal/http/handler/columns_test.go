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
		name               string
		path               string
		inputBody          any
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name:      "Success",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": validColumn.Name.String()},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if name != validColumn.Name {
						t.Errorf("got name %v, want %v", name, validColumn.Name)
					}
					return validColumn, nil
				}
			},
			wantCode: http.StatusCreated,
			wantBody: map[string]any{
				"id":        validColumn.ID.String(),
				"boardId":   validColumn.BoardID.String(),
				"name":      validColumn.Name.String(),
				"position":  validColumn.Position.Int64(),
				"createdAt": validColumn.CreatedAt.Format(timeFormat),
				"updatedAt": validColumn.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:      "Invalid board id",
			path:      "/v1/boards/not-a-uuid/columns",
			inputBody: map[string]string{"name": "To Do"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid JSON",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: "{\"name\":\"broken\"",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJsonBody(),
		},
		{
			name:      "Invalid name",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": "   "},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name:      "Missing context user",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": "To Do"},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenBody(),
		},
		{
			name:      "Board not found",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": "To Do"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, service.ErrBoardNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: boardNotFoundErrorBody(),
		},
		{
			name:      "Internal error",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": "To Do"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name:      "Unexpected error",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns",
			inputBody: map[string]string{"name": "To Do"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, errors.New("db exploded")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
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
			mockColumns := &MockColumnService{}
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Create(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
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
		name               string
		path               string
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name: "Success",
			path: "/v1/boards/" + validBoard.ID.String() + "/columns",
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					return []domain.Column{first, second}, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: []map[string]any{
				{
					"id":        first.ID.String(),
					"boardId":   first.BoardID.String(),
					"name":      first.Name.String(),
					"position":  first.Position.Int64(),
					"createdAt": first.CreatedAt.Format(timeFormat),
					"updatedAt": first.UpdatedAt.Format(timeFormat),
				},
				{
					"id":        second.ID.String(),
					"boardId":   second.BoardID.String(),
					"name":      second.Name.String(),
					"position":  second.Position.Int64(),
					"createdAt": second.CreatedAt.Format(timeFormat),
					"updatedAt": second.UpdatedAt.Format(timeFormat),
				},
			},
		},
		{
			name:     "Invalid board id",
			path:     "/v1/boards/not-a-uuid/columns",
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:     "Missing context user",
			path:     "/v1/boards/" + validBoard.ID.String() + "/columns",
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
		},
		{
			name: "Board not found",
			path: "/v1/boards/" + validBoard.ID.String() + "/columns",
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
					return nil, service.ErrBoardNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: boardNotFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: "/v1/boards/" + validBoard.ID.String() + "/columns",
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
					return nil, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
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
			mockColumns := &MockColumnService{}
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.List(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestColumns_UpdateByID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	updatedName, err := domain.NewColumnName("Renamed Column")
	if err != nil {
		t.Fatalf("NewColumnName() error = %v", err)
	}
	updatedColumn := validColumn
	updatedColumn.Name = updatedName
	updatedColumn.UpdatedAt = testutil.FixedTime5mFromNow()

	okPath := "/v1/boards/" + validBoard.ID.String() + "/columns/" + validColumn.ID.String()

	tests := []struct {
		name               string
		path               string
		inputBody          any
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name:      "Success (name update)",
			path:      okPath,
			inputBody: map[string]string{"name": updatedName.String()},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					if name == nil || *name != updatedName {
						t.Errorf("got name %v, want %v", name, updatedName)
					}
					return updatedColumn, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":        updatedColumn.ID.String(),
				"boardId":   updatedColumn.BoardID.String(),
				"name":      updatedColumn.Name.String(),
				"position":  updatedColumn.Position.Int64(),
				"createdAt": updatedColumn.CreatedAt.Format(timeFormat),
				"updatedAt": updatedColumn.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:      "Success (empty body no-op)",
			path:      okPath,
			inputBody: map[string]any{},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
					if name != nil {
						t.Errorf("got name %+v, want nil", name)
					}
					return validColumn, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":        validColumn.ID.String(),
				"boardId":   validColumn.BoardID.String(),
				"name":      validColumn.Name.String(),
				"position":  validColumn.Position.Int64(),
				"createdAt": validColumn.CreatedAt.Format(timeFormat),
				"updatedAt": validColumn.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:      "Invalid board id",
			path:      "/v1/boards/not-a-uuid/columns/" + validColumn.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid column id",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns/not-a-uuid",
			inputBody: map[string]string{"name": "Renamed"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:      "Invalid JSON",
			path:      okPath,
			inputBody: "{\"name\":\"broken\"",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJsonBody(),
		},
		{
			name:      "Invalid name",
			path:      okPath,
			inputBody: map[string]string{"name": "   "},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name:      "Missing context user",
			path:      okPath,
			inputBody: map[string]string{"name": "Renamed"},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenBody(),
		},
		{
			name:      "Column not found",
			path:      okPath,
			inputBody: map[string]string{"name": "Renamed"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundByFieldError("columnId"),
		},
		{
			name:      "Internal error",
			path:      okPath,
			inputBody: map[string]string{"name": "Renamed"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName) (domain.Column, error) {
					return domain.Column{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
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
			mockColumns := &MockColumnService{}
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.UpdateByID(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestColumns_Move(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	targetPosition, err := domain.NewColumnPosition(2)
	if err != nil {
		t.Fatalf("NewColumnPosition() error = %v", err)
	}

	okPath := "/v1/boards/" + validBoard.ID.String() + "/columns/" + validColumn.ID.String() + "/position"

	tests := []struct {
		name               string
		path               string
		inputBody          any
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name:      "Success",
			path:      okPath,
			inputBody: map[string]int64{"targetPosition": targetPosition.Int64()},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, gotTargetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					if gotTargetPosition != targetPosition {
						t.Errorf("got target position %v, want %v", gotTargetPosition, targetPosition)
					}
					return targetPosition, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"position": targetPosition.Int64(),
			},
		},
		{
			name:      "Invalid board id",
			path:      "/v1/boards/not-a-uuid/columns/" + validColumn.ID.String() + "/position",
			inputBody: map[string]int64{"targetPosition": 1},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid column id",
			path:      "/v1/boards/" + validBoard.ID.String() + "/columns/not-a-uuid/position",
			inputBody: map[string]int64{"targetPosition": 1},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:      "Invalid JSON",
			path:      okPath,
			inputBody: "{\"targetPosition\":",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJsonBody(),
		},
		{
			name:      "Invalid target position",
			path:      okPath,
			inputBody: map[string]int64{"targetPosition": 0},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("targetPosition", []string{"Position is invalid"}),
		},
		{
			name:      "Missing context user",
			path:      okPath,
			inputBody: map[string]int64{"targetPosition": 1},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenBody(),
		},
		{
			name:      "Index out of bounds",
			path:      okPath,
			inputBody: map[string]int64{"targetPosition": 10},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, service.ErrIndexOutOfBounds
				}
			},
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("targetPosition", []string{"Index out of bounds"}),
		},
		{
			name:      "Column not found",
			path:      okPath,
			inputBody: map[string]int64{"targetPosition": 1},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundByFieldError("columnId"),
		},
		{
			name:      "Internal error",
			path:      okPath,
			inputBody: map[string]int64{"targetPosition": 1},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var req *http.Request
			if raw, ok := tt.inputBody.(string); ok {
				req = httptest.NewRequest(http.MethodPut, tt.path, strings.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = testutil.NewJSONRequestAndRecorder(t, http.MethodPut, tt.path, tt.inputBody)
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
				req.SetPathValue("columnId", strings.TrimSuffix(parts[1], "/position"))
			}

			rr := httptest.NewRecorder()
			mockColumns := &MockColumnService{}
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Move(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestColumns_Delete(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	okPath := "/v1/boards/" + validBoard.ID.String() + "/columns/" + validColumn.ID.String()

	tests := []struct {
		name               string
		path               string
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name: "Success",
			path: okPath,
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.DeleteFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					return nil
				}
			},
			wantCode: http.StatusNoContent,
			wantBody: nil,
		},
		{
			name:     "Invalid board id",
			path:     "/v1/boards/not-a-uuid/columns/" + validColumn.ID.String(),
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:     "Invalid column id",
			path:     "/v1/boards/" + validBoard.ID.String() + "/columns/not-a-uuid",
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:     "Missing context user",
			path:     okPath,
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
		},
		{
			name: "Column not found",
			path: okPath,
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.DeleteFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error {
					return service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundByFieldError("columnId"),
		},
		{
			name: "Internal error",
			path: okPath,
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.DeleteFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error {
					return service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
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

			boardAndColumn := strings.TrimPrefix(tt.path, "/v1/boards/")
			parts := strings.Split(boardAndColumn, "/columns/")
			if len(parts) == 2 {
				req.SetPathValue("boardId", parts[0])
				req.SetPathValue("columnId", parts[1])
			}

			rr := httptest.NewRecorder()
			mockColumns := &MockColumnService{}
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Delete(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			if tt.wantCode != http.StatusNoContent {
				testutil.AssertContentType(t, rr, "application/json")
			}
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}
