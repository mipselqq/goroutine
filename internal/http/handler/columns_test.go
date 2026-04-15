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
