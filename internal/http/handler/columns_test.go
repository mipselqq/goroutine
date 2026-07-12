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
		boardID            string
		inputBody          any
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name:      "Success",
			boardID:   validBoard.ID.String(),
			inputBody: map[string]string{"name": validColumn.Name.String(), "description": validColumn.Description.String()},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if name != validColumn.Name {
						t.Errorf("got name %v, want %v", name, validColumn.Name)
					}
					if description != validColumn.Description {
						t.Errorf("got description %v, want %v", description, validColumn.Description)
					}
					return validColumn, nil
				}
			},
			wantCode: http.StatusCreated,
			wantBody: map[string]any{
				"id":          validColumn.ID.String(),
				"boardId":     validColumn.BoardID.String(),
				"name":        validColumn.Name.String(),
				"description": validColumn.Description.String(),
				"position":    validColumn.Position.Int64(),
				"createdAt":   validColumn.CreatedAt.Format(testutil.TimeFormat),
				"updatedAt":   validColumn.UpdatedAt.Format(testutil.TimeFormat),
			},
		},
		{
			name:      "Invalid board id",
			boardID:   "not-a-uuid",
			inputBody: map[string]string{"name": "To Do"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid JSON",
			boardID:   validBoard.ID.String(),
			inputBody: "{\"name\":\"broken\"",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJSONError(),
		},
		{
			name:      "Invalid name",
			boardID:   validBoard.ID.String(),
			inputBody: map[string]string{"name": "   ", "description": validColumn.Description.String()},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("name", []string{"Name is too short"}),
		},
		{
			name:      "Description too long",
			boardID:   validBoard.ID.String(),
			inputBody: map[string]string{"name": validColumn.Name.String(), "description": strings.Repeat("a", 1025)},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("description", []string{"Description is too long"}),
		},
		{
			name:      "Missing context user",
			boardID:   validBoard.ID.String(),
			inputBody: map[string]string{"name": "To Do"},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenError(),
		},
		{
			name:      "Board not found",
			boardID:   validBoard.ID.String(),
			inputBody: map[string]string{"name": "To Do"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
					return domain.Column{}, service.ErrBoardNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: boardNotFoundError(),
		},
		{
			name:      "Internal error",
			boardID:   validBoard.ID.String(),
			inputBody: map[string]string{"name": "To Do"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
					return domain.Column{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalError(),
		},
		{
			name:      "Unexpected error",
			boardID:   validBoard.ID.String(),
			inputBody: map[string]string{"name": "To Do"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
					return domain.Column{}, errors.New("db exploded")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalError(),
		},
		{
			name:      "Body too large",
			boardID:   validBoard.ID.String(),
			inputBody: testutil.Big25KBJSON(),
			wantCode:  http.StatusRequestEntityTooLarge,
			wantBody:  payloadTooLargeError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/v1/boards/" + tt.boardID + "/columns"
			var req *http.Request
			if raw, ok := tt.inputBody.(string); ok {
				req = httptest.NewRequest(http.MethodPost, path, strings.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = testutil.NewJSONRequestAndRecorder(t, http.MethodPost, path, tt.inputBody)
			}

			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)
			req.SetPathValue("boardId", tt.boardID)

			rr := httptest.NewRecorder()
			mockColumns := NewMockColumnService(t)
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
			h.Create(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestColumns_ListByBoardID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	first := testutil.ValidColumn(validBoard.ID)
	second := testutil.ValidColumn(validBoard.ID)
	second.Position, _ = domain.NewColumnPosition(first.Position.Int64() + 1)

	tests := []struct {
		name               string
		boardID            string
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name:    "Success",
			boardID: validBoard.ID.String(),
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.ListByBoardIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
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
					"id":          first.ID.String(),
					"boardId":     first.BoardID.String(),
					"name":        first.Name.String(),
					"description": first.Description.String(),
					"position":    first.Position.Int64(),
					"createdAt":   first.CreatedAt.Format(testutil.TimeFormat),
					"updatedAt":   first.UpdatedAt.Format(testutil.TimeFormat),
				},
				{
					"id":          second.ID.String(),
					"boardId":     second.BoardID.String(),
					"name":        second.Name.String(),
					"description": second.Description.String(),
					"position":    second.Position.Int64(),
					"createdAt":   second.CreatedAt.Format(testutil.TimeFormat),
					"updatedAt":   second.UpdatedAt.Format(testutil.TimeFormat),
				},
			},
		},
		{
			name:     "Invalid board id",
			boardID:  "not-a-uuid",
			wantCode: http.StatusBadRequest,
			wantBody: validationError("boardId", []string{"Invalid board id"}),
		},
		{
			name:     "Missing context user",
			boardID:  validBoard.ID.String(),
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenError(),
		},
		{
			name:    "Board not found",
			boardID: validBoard.ID.String(),
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.ListByBoardIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
					return nil, service.ErrBoardNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: boardNotFoundError(),
		},
		{
			name:    "Internal error",
			boardID: validBoard.ID.String(),
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.ListByBoardIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
					return nil, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/v1/boards/" + tt.boardID + "/columns"
			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)
			req.SetPathValue("boardId", tt.boardID)

			rr := httptest.NewRecorder()
			mockColumns := NewMockColumnService(t)
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
			h.ListByBoardID(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestColumns_Update(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	updatedName, err := domain.NewColumnName("Renamed Column")
	if err != nil {
		t.Fatalf("NewColumnName() error = %v", err)
	}
	updatedColumn := validColumn
	updatedColumn.Name = updatedName
	updatedColumn.UpdatedAt = testutil.Fixed5mFromNow()

	updatedDescOnly, err := domain.NewColumnDescription("Updated description only")
	if err != nil {
		t.Fatalf("NewColumnDescription() error = %v", err)
	}
	updatedDescriptionOnlyColumn := validColumn
	updatedDescriptionOnlyColumn.Description = updatedDescOnly
	updatedDescriptionOnlyColumn.UpdatedAt = testutil.Fixed5mFromNow()

	emptyDescriptionColumn := validColumn
	emptyDescriptionColumn.Name = updatedName
	emptyDesc, errEmpty := domain.NewColumnDescription("")
	if errEmpty != nil {
		t.Fatalf("NewColumnDescription() error = %v", errEmpty)
	}
	emptyDescriptionColumn.Description = emptyDesc
	emptyDescriptionColumn.UpdatedAt = testutil.Fixed5mFromNow()

	tests := []struct {
		name               string
		boardID            string
		columnID           string
		inputBody          any
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name:      "Success (name update)",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": updatedName.String()},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
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
					if description != nil {
						t.Errorf("got description %+v, want nil", description)
					}
					return updatedColumn, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":          updatedColumn.ID.String(),
				"boardId":     updatedColumn.BoardID.String(),
				"name":        updatedColumn.Name.String(),
				"description": updatedColumn.Description.String(),
				"position":    updatedColumn.Position.Int64(),
				"createdAt":   updatedColumn.CreatedAt.Format(testutil.TimeFormat),
				"updatedAt":   updatedColumn.UpdatedAt.Format(testutil.TimeFormat),
			},
		},
		{
			name:      "Success (description only)",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"description": updatedDescOnly.String()},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					if name != nil {
						t.Errorf("got name %+v, want nil", name)
					}
					if description == nil || *description != updatedDescOnly {
						t.Errorf("got description %v, want %v", description, updatedDescOnly)
					}
					return updatedDescriptionOnlyColumn, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":          updatedDescriptionOnlyColumn.ID.String(),
				"boardId":     updatedDescriptionOnlyColumn.BoardID.String(),
				"name":        updatedDescriptionOnlyColumn.Name.String(),
				"description": updatedDescriptionOnlyColumn.Description.String(),
				"position":    updatedDescriptionOnlyColumn.Position.Int64(),
				"createdAt":   updatedDescriptionOnlyColumn.CreatedAt.Format(testutil.TimeFormat),
				"updatedAt":   updatedDescriptionOnlyColumn.UpdatedAt.Format(testutil.TimeFormat),
			},
		},
		{
			name:      "Success (empty body no-op)",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]any{},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					if name != nil {
						t.Errorf("got name %+v, want nil", name)
					}
					if description != nil {
						t.Errorf("got description %+v, want nil", description)
					}
					return validColumn, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":          validColumn.ID.String(),
				"boardId":     validColumn.BoardID.String(),
				"name":        validColumn.Name.String(),
				"description": validColumn.Description.String(),
				"position":    validColumn.Position.Int64(),
				"createdAt":   validColumn.CreatedAt.Format(testutil.TimeFormat),
				"updatedAt":   validColumn.UpdatedAt.Format(testutil.TimeFormat),
			},
		},
		{
			name:      "Empty description",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": updatedName.String(), "description": ""},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					if name == nil || description == nil {
						t.Errorf("got name %+v, description %+v, want non-nil, non-nil", name, description)
					}
					return emptyDescriptionColumn, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":          emptyDescriptionColumn.ID.String(),
				"boardId":     emptyDescriptionColumn.BoardID.String(),
				"name":        emptyDescriptionColumn.Name.String(),
				"description": emptyDescriptionColumn.Description.String(),
				"position":    emptyDescriptionColumn.Position.Int64(),
				"createdAt":   emptyDescriptionColumn.CreatedAt.Format(testutil.TimeFormat),
				"updatedAt":   emptyDescriptionColumn.UpdatedAt.Format(testutil.TimeFormat),
			},
		},
		{
			name:      "Invalid board id",
			boardID:   "not-a-uuid",
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid column id",
			boardID:   validBoard.ID.String(),
			columnID:  "not-a-uuid",
			inputBody: map[string]string{"name": "Renamed"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("columnId", []string{"Invalid column id"}),
		},
		{
			name:      "Invalid JSON",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: "{\"name\":\"broken\"",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJSONError(),
		},
		{
			name:      "Invalid name",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "   ", "description": validColumn.Description.String()},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("name", []string{"Name is too short"}),
		},
		{
			name:      "Invalid description",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": validColumn.Name.String(), "description": strings.Repeat("a", 1025)},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("description", []string{"Description is too long"}),
		},
		{
			name:      "Missing context user",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenError(),
		},
		{
			name:      "Column not found",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					return domain.Column{}, service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundError("columnId"),
		},
		{
			name:      "Internal error",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.UpdateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					return domain.Column{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalError(),
		},
		{
			name:      "Body too large",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: testutil.Big25KBJSON(),
			wantCode:  http.StatusRequestEntityTooLarge,
			wantBody:  payloadTooLargeError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/v1/boards/" + tt.boardID + "/columns/" + tt.columnID
			var req *http.Request
			if raw, ok := tt.inputBody.(string); ok {
				req = httptest.NewRequest(http.MethodPatch, path, strings.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = testutil.NewJSONRequestAndRecorder(t, http.MethodPatch, path, tt.inputBody)
			}

			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)

			req.SetPathValue("boardId", tt.boardID)
			req.SetPathValue("columnId", tt.columnID)

			rr := httptest.NewRecorder()
			mockColumns := NewMockColumnService(t)
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
			h.Update(rr, req)

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

	tests := []struct {
		name               string
		boardID            string
		columnID           string
		inputBody          any
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name:      "Success",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
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
			boardID:   "not-a-uuid",
			columnID:  validColumn.ID.String(),
			inputBody: map[string]int64{"targetPosition": 1},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid column id",
			boardID:   validBoard.ID.String(),
			columnID:  "not-a-uuid",
			inputBody: map[string]int64{"targetPosition": 1},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("columnId", []string{"Invalid column id"}),
		},
		{
			name:      "Invalid JSON",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: "{\"targetPosition\":",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJSONError(),
		},
		{
			name:      "Invalid target position",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]int64{"targetPosition": 0},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationError("targetPosition", []string{"Position is invalid"}),
		},
		{
			name:      "Missing context user",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]int64{"targetPosition": 1},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenError(),
		},
		{
			name:      "Index out of bounds",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]int64{"targetPosition": 10},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, service.ErrIndexOutOfBounds
				}
			},
			wantCode: http.StatusBadRequest,
			wantBody: validationError("targetPosition", []string{"Index out of bounds"}),
		},
		{
			name:      "Column not found",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]int64{"targetPosition": 1},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundError("columnId"),
		},
		{
			name:      "Internal error",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]int64{"targetPosition": 1},
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalError(),
		},
		{
			name:      "Body too large",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: testutil.Big25KBJSON(),
			wantCode:  http.StatusRequestEntityTooLarge,
			wantBody:  payloadTooLargeError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/v1/boards/" + tt.boardID + "/columns/" + tt.columnID + "/position"
			var req *http.Request
			if raw, ok := tt.inputBody.(string); ok {
				req = httptest.NewRequest(http.MethodPut, path, strings.NewReader(raw))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, _ = testutil.NewJSONRequestAndRecorder(t, http.MethodPut, path, tt.inputBody)
			}

			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)

			req.SetPathValue("boardId", tt.boardID)
			req.SetPathValue("columnId", tt.columnID)

			rr := httptest.NewRecorder()
			mockColumns := NewMockColumnService(t)
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
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

	tests := []struct {
		name               string
		boardID            string
		columnID           string
		context            context.Context
		setupColumnService func(t *testing.T, s *MockColumnService)
		wantCode           int
		wantBody           any
	}{
		{
			name:     "Success",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
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
			boardID:  "not-a-uuid",
			columnID: validColumn.ID.String(),
			wantCode: http.StatusBadRequest,
			wantBody: validationError("boardId", []string{"Invalid board id"}),
		},
		{
			name:     "Invalid column id",
			boardID:  validBoard.ID.String(),
			columnID: "not-a-uuid",
			wantCode: http.StatusBadRequest,
			wantBody: validationError("columnId", []string{"Invalid column id"}),
		},
		{
			name:     "Missing context user",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenError(),
		},
		{
			name:     "Column not found",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.DeleteFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error {
					return service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundError("columnId"),
		},
		{
			name:     "Internal error",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			setupColumnService: func(t *testing.T, s *MockColumnService) {
				s.DeleteFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) error {
					return service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalError(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/v1/boards/" + tt.boardID + "/columns/" + tt.columnID
			req := httptest.NewRequest(http.MethodDelete, path, http.NoBody)
			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)

			req.SetPathValue("boardId", tt.boardID)
			req.SetPathValue("columnId", tt.columnID)

			rr := httptest.NewRecorder()
			mockColumns := NewMockColumnService(t)
			if tt.setupColumnService != nil {
				tt.setupColumnService(t, mockColumns)
			}

			logger := testutil.NewLogger(t)
			h := handler.NewColumns(logger, mockColumns, httpschema.MustNewErrorResponder(logger, testutil.FixedNowStr))
			h.Delete(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			if tt.wantCode != http.StatusNoContent {
				testutil.AssertContentType(t, rr, "application/json")
			}
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}
