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

func TestTasks_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	validTask := testutil.ValidTask(validColumn.ID)

	tests := []struct {
		name             string
		boardID          string
		columnID         string
		inputBody        any
		context          context.Context
		setupTaskService func(t *testing.T, s *MockTaskService)
		wantCode         int
		wantBody         any
	}{
		{
			name:     "Success",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			inputBody: map[string]string{
				"name":        validTask.Name.String(),
				"description": validTask.Description.String(),
			},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					if name != validTask.Name {
						t.Errorf("got name %v, want %v", name, validTask.Name)
					}
					if description != validTask.Description {
						t.Errorf("got description %v, want %v", description, validTask.Description)
					}
					return validTask, nil
				}
			},
			wantCode: http.StatusCreated,
			wantBody: map[string]any{
				"id":          validTask.ID.String(),
				"columnId":    validTask.ColumnID.String(),
				"name":        validTask.Name.String(),
				"description": validTask.Description.String(),
				"position":    validTask.Position.Int64(),
				"createdAt":   validTask.CreatedAt.Format(timeFormat),
				"updatedAt":   validTask.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:      "Invalid board id",
			boardID:   "not-a-uuid",
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "Name", "description": "Description"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid column id",
			boardID:   validBoard.ID.String(),
			columnID:  "not-a-uuid",
			inputBody: map[string]string{"name": "Name", "description": "Description"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:      "Invalid JSON",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: "{\"name\":\"broken\"",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJsonBody(),
		},
		{
			name:      "Invalid name",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "   ", "description": "ok"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name:      "Missing context user",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "ok", "description": "ok"},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenBody(),
		},
		{
			name:      "Column not found",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "ok", "description": "ok"},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					return domain.Task{}, service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundByFieldError("columnId"),
		},
		{
			name:      "Internal error",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "ok", "description": "ok"},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					return domain.Task{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name:      "Unexpected error",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			inputBody: map[string]string{"name": "ok", "description": "ok"},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.CreateFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					return domain.Task{}, errors.New("db exploded")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/v1/boards/" + tt.boardID + "/columns/" + tt.columnID + "/tasks"
			req := buildTaskRequest(t, http.MethodPost, path, tt.inputBody)
			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)
			req.SetPathValue("boardId", tt.boardID)
			req.SetPathValue("columnId", tt.columnID)

			rr := httptest.NewRecorder()
			mockTasks := &MockTaskService{}
			if tt.setupTaskService != nil {
				tt.setupTaskService(t, mockTasks)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewTasks(logger, mockTasks, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Create(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestTasks_List(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	first := testutil.ValidTask(validColumn.ID)
	second := testutil.ValidTask(validColumn.ID)
	second.Position = testutil.MustTaskPosition(t, first.Position.Int64()+1)

	tests := []struct {
		name             string
		boardID          string
		columnID         string
		context          context.Context
		setupTaskService func(t *testing.T, s *MockTaskService)
		wantCode         int
		wantBody         any
	}{
		{
			name:     "Success",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) ([]domain.Task, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					return []domain.Task{first, second}, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: []map[string]any{
				{
					"id":          first.ID.String(),
					"columnId":    first.ColumnID.String(),
					"name":        first.Name.String(),
					"description": first.Description.String(),
					"position":    first.Position.Int64(),
					"createdAt":   first.CreatedAt.Format(timeFormat),
					"updatedAt":   first.UpdatedAt.Format(timeFormat),
				},
				{
					"id":          second.ID.String(),
					"columnId":    second.ColumnID.String(),
					"name":        second.Name.String(),
					"description": second.Description.String(),
					"position":    second.Position.Int64(),
					"createdAt":   second.CreatedAt.Format(timeFormat),
					"updatedAt":   second.UpdatedAt.Format(timeFormat),
				},
			},
		},
		{
			name:     "Invalid board id",
			boardID:  "not-a-uuid",
			columnID: validColumn.ID.String(),
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:     "Invalid column id",
			boardID:  validBoard.ID.String(),
			columnID: "not-a-uuid",
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:     "Missing context user",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
		},
		{
			name:     "Column not found",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) ([]domain.Task, error) {
					return nil, service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundByFieldError("columnId"),
		},
		{
			name:     "Internal error",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.ListFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) ([]domain.Task, error) {
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

			path := "/v1/boards/" + tt.boardID + "/columns/" + tt.columnID + "/tasks"
			req := httptest.NewRequest(http.MethodGet, path, http.NoBody)
			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)
			req.SetPathValue("boardId", tt.boardID)
			req.SetPathValue("columnId", tt.columnID)

			rr := httptest.NewRecorder()
			mockTasks := &MockTaskService{}
			if tt.setupTaskService != nil {
				tt.setupTaskService(t, mockTasks)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewTasks(logger, mockTasks, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.List(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestTasks_UpdateByID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	validTask := testutil.ValidTask(validColumn.ID)
	updatedName, err := domain.NewTaskName("Renamed Task")
	if err != nil {
		t.Fatalf("NewTaskName() error = %v", err)
	}
	updatedDescription, err := domain.NewTaskDescription("Renamed Description")
	if err != nil {
		t.Fatalf("NewTaskDescription() error = %v", err)
	}
	updatedTask := validTask
	updatedTask.Name = updatedName
	updatedTask.Description = updatedDescription
	updatedTask.UpdatedAt = testutil.FixedTime5mFromNow()

	tests := []struct {
		name             string
		boardID          string
		columnID         string
		taskID           string
		inputBody        any
		context          context.Context
		setupTaskService func(t *testing.T, s *MockTaskService)
		wantCode         int
		wantBody         any
	}{
		{
			name:      "Success (name and description update)",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]string{"name": updatedName.String(), "description": updatedDescription.String()},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					if taskID != validTask.ID {
						t.Errorf("got task id %v, want %v", taskID, validTask.ID)
					}
					if name == nil || *name != updatedName {
						t.Errorf("got name %v, want %v", name, updatedName)
					}
					if description == nil || *description != updatedDescription {
						t.Errorf("got description %v, want %v", description, updatedDescription)
					}
					return updatedTask, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":          updatedTask.ID.String(),
				"columnId":    updatedTask.ColumnID.String(),
				"name":        updatedTask.Name.String(),
				"description": updatedTask.Description.String(),
				"position":    updatedTask.Position.Int64(),
				"createdAt":   updatedTask.CreatedAt.Format(timeFormat),
				"updatedAt":   updatedTask.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:      "Success (empty body no-op)",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					if name != nil {
						t.Errorf("got name %+v, want nil", name)
					}
					if description != nil {
						t.Errorf("got description %+v, want nil", description)
					}
					return validTask, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":          validTask.ID.String(),
				"columnId":    validTask.ColumnID.String(),
				"name":        validTask.Name.String(),
				"description": validTask.Description.String(),
				"position":    validTask.Position.Int64(),
				"createdAt":   validTask.CreatedAt.Format(timeFormat),
				"updatedAt":   validTask.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:      "Invalid board id",
			boardID:   "not-a-uuid",
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid column id",
			boardID:   validBoard.ID.String(),
			columnID:  "not-a-uuid",
			taskID:    validTask.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:      "Invalid task id",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    "not-a-uuid",
			inputBody: map[string]string{"name": "Renamed"},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("taskId", []string{"Invalid task id"}),
		},
		{
			name:      "Invalid JSON",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: "{\"name\":\"broken\"",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJsonBody(),
		},
		{
			name:      "Invalid name",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]string{"name": "   "},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name:      "Missing context user",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenBody(),
		},
		{
			name:      "Task not found",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					return domain.Task{}, service.ErrTaskNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: taskNotFoundErrorBody(),
		},
		{
			name:      "Internal error",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]string{"name": "Renamed"},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.UpdateByIDFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					return domain.Task{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/v1/boards/" + tt.boardID + "/columns/" + tt.columnID + "/tasks/" + tt.taskID
			req := buildTaskRequest(t, http.MethodPatch, path, tt.inputBody)
			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)
			req.SetPathValue("boardId", tt.boardID)
			req.SetPathValue("columnId", tt.columnID)
			req.SetPathValue("taskId", tt.taskID)

			rr := httptest.NewRecorder()
			mockTasks := &MockTaskService{}
			if tt.setupTaskService != nil {
				tt.setupTaskService(t, mockTasks)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewTasks(logger, mockTasks, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.UpdateByID(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestTasks_Move(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	validTask := testutil.ValidTask(validColumn.ID)
	targetColumn := testutil.NewValidColumn(t, validBoard.ID, "Done", 2)
	targetPosition := testutil.MustTaskPosition(t, 2)

	tests := []struct {
		name             string
		boardID          string
		columnID         string
		taskID           string
		inputBody        any
		context          context.Context
		setupTaskService func(t *testing.T, s *MockTaskService)
		wantCode         int
		wantBody         any
	}{
		{
			name:     "Success",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			taskID:   validTask.ID.String(),
			inputBody: map[string]any{
				"targetColumnId": targetColumn.ID.String(),
				"targetPosition": targetPosition.Int64(),
			},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					if taskID != validTask.ID {
						t.Errorf("got task id %v, want %v", taskID, validTask.ID)
					}
					if gotTargetColumnID != targetColumn.ID {
						t.Errorf("got target column id %v, want %v", gotTargetColumnID, targetColumn.ID)
					}
					if gotTargetPosition != targetPosition {
						t.Errorf("got target position %v, want %v", gotTargetPosition, targetPosition)
					}
					return targetColumn.ID, targetPosition, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"columnId": targetColumn.ID.String(),
				"position": targetPosition.Int64(),
			},
		},
		{
			name:      "Invalid board id",
			boardID:   "not-a-uuid",
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 1},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:      "Invalid column id",
			boardID:   validBoard.ID.String(),
			columnID:  "not-a-uuid",
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 1},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:      "Invalid task id",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    "not-a-uuid",
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 1},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("taskId", []string{"Invalid task id"}),
		},
		{
			name:      "Invalid JSON",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: "{\"targetPosition\":",
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJsonBody(),
		},
		{
			name:      "Invalid target column id",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": "not-a-uuid", "targetPosition": 1},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("targetColumnId", []string{"Invalid target column id"}),
		},
		{
			name:      "Invalid target position",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 0},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("targetPosition", []string{"Position is invalid"}),
		},
		{
			name:      "Missing context user",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 1},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenBody(),
		},
		{
			name:      "Index out of bounds",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 10},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					return domain.ColumnID{}, domain.TaskPosition{}, service.ErrIndexOutOfBounds
				}
			},
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("targetPosition", []string{"Index out of bounds"}),
		},
		{
			name:      "Task not found",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 1},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					return domain.ColumnID{}, domain.TaskPosition{}, service.ErrTaskNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: taskNotFoundErrorBody(),
		},
		{
			name:      "Target column not found",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 1},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					return domain.ColumnID{}, domain.TaskPosition{}, service.ErrColumnNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: columnNotFoundByFieldError("targetColumnId"),
		},
		{
			name:      "Internal error",
			boardID:   validBoard.ID.String(),
			columnID:  validColumn.ID.String(),
			taskID:    validTask.ID.String(),
			inputBody: map[string]any{"targetColumnId": targetColumn.ID.String(), "targetPosition": 1},
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.MoveFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					return domain.ColumnID{}, domain.TaskPosition{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			path := "/v1/boards/" + tt.boardID + "/columns/" + tt.columnID + "/tasks/" + tt.taskID + "/position"
			req := buildTaskRequest(t, http.MethodPut, path, tt.inputBody)
			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)
			req.SetPathValue("boardId", tt.boardID)
			req.SetPathValue("columnId", tt.columnID)
			req.SetPathValue("taskId", tt.taskID)

			rr := httptest.NewRecorder()
			mockTasks := &MockTaskService{}
			if tt.setupTaskService != nil {
				tt.setupTaskService(t, mockTasks)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewTasks(logger, mockTasks, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Move(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestTasks_Delete(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	validTask := testutil.ValidTask(validColumn.ID)

	tests := []struct {
		name             string
		boardID          string
		columnID         string
		taskID           string
		context          context.Context
		setupTaskService func(t *testing.T, s *MockTaskService)
		wantCode         int
		wantBody         any
	}{
		{
			name:     "Success",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			taskID:   validTask.ID.String(),
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.DeleteFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
					if callerID != validBoard.OwnerID {
						t.Errorf("got caller id %v, want %v", callerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					if taskID != validTask.ID {
						t.Errorf("got task id %v, want %v", taskID, validTask.ID)
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
			taskID:   validTask.ID.String(),
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name:     "Invalid column id",
			boardID:  validBoard.ID.String(),
			columnID: "not-a-uuid",
			taskID:   validTask.ID.String(),
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("columnId", []string{"Invalid column id"}),
		},
		{
			name:     "Invalid task id",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			taskID:   "not-a-uuid",
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("taskId", []string{"Invalid task id"}),
		},
		{
			name:     "Missing context user",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			taskID:   validTask.ID.String(),
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
		},
		{
			name:     "Task not found",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			taskID:   validTask.ID.String(),
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.DeleteFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
					return service.ErrTaskNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: taskNotFoundErrorBody(),
		},
		{
			name:     "Internal error",
			boardID:  validBoard.ID.String(),
			columnID: validColumn.ID.String(),
			taskID:   validTask.ID.String(),
			setupTaskService: func(t *testing.T, s *MockTaskService) {
				s.DeleteFunc = func(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
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

			path := "/v1/boards/" + tt.boardID + "/columns/" + tt.columnID + "/tasks/" + tt.taskID
			req := httptest.NewRequest(http.MethodDelete, path, http.NoBody)
			ctx := tt.context
			if ctx == nil {
				ctx = context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID)
			}
			req = req.WithContext(ctx)
			req.SetPathValue("boardId", tt.boardID)
			req.SetPathValue("columnId", tt.columnID)
			req.SetPathValue("taskId", tt.taskID)

			rr := httptest.NewRecorder()
			mockTasks := &MockTaskService{}
			if tt.setupTaskService != nil {
				tt.setupTaskService(t, mockTasks)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewTasks(logger, mockTasks, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Delete(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			if tt.wantCode != http.StatusNoContent {
				testutil.AssertContentType(t, rr, "application/json")
			}
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func buildTaskRequest(t *testing.T, method, path string, body any) *http.Request {
	t.Helper()

	if raw, ok := body.(string); ok {
		req := httptest.NewRequest(method, path, strings.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		return req
	}
	req, _ := testutil.NewJSONRequestAndRecorder(t, method, path, body)
	return req
}
