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

	"goroutine/internal/domain"
	"goroutine/internal/http/handler"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

type boardsTestCase struct {
	name              string
	inputBody         any
	context           context.Context
	setupBoardService func(t *testing.T, s *MockBoardService)
	wantCode          int
	wantBody          any
	path              string
}

const timeFormat = "2006-01-02T15:04:05.000Z07:00"

func TestBoards_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []boardsTestCase{
		{
			name:      "Success",
			inputBody: map[string]string{"name": validBoard.Name.String(), "description": validBoard.Description.String()},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("got ownerID %v, want %v", ownerID, validBoard.OwnerID)
					}
					return validBoard, nil
				}
			},
			wantCode: http.StatusCreated,
			wantBody: map[string]string{
				"id":          validBoard.ID.String(),
				"ownerId":     validBoard.OwnerID.String(),
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
				"createdAt":   validBoard.CreatedAt.Format(timeFormat),
				"updatedAt":   validBoard.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:      "Empty name",
			inputBody: map[string]string{"name": "", "description": validBoard.Description.String()},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name:      "Description too long",
			inputBody: map[string]string{"name": validBoard.Name.String(), "description": strings.Repeat("a", 1025)},
			wantCode:  http.StatusBadRequest,
			wantBody:  validationErrorBody("description", []string{"Description is too long"}),
		},
		{
			name:      "Invalid JSON",
			inputBody: json.RawMessage([]byte(fmt.Sprintf(`{"name": %q, "description": %q`, validBoard.Name.String(), validBoard.Description.String()))), // missing closing brace
			wantCode:  http.StatusBadRequest,
			wantBody:  invalidJsonBody(),
		},
		{
			name:      "No context user ID",
			inputBody: map[string]string{"name": validBoard.Name.String(), "description": validBoard.Description.String()},
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenBody(),
		},
		{
			name:      "Internal error",
			inputBody: map[string]string{"name": validBoard.Name.String(), "description": validBoard.Description.String()},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name:      "Unknown error",
			inputBody: map[string]string{"name": validBoard.Name.String(), "description": validBoard.Description.String()},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("unknown error")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
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

			s := &MockBoardService{}

			if tt.setupBoardService != nil {
				tt.setupBoardService(t, s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))

			h.Create(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
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
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("got ownerID %v, want %v", ownerID, validBoard.OwnerID)
					}

					return []domain.Board{validBoard}, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: []map[string]string{
				{
					"id":          validBoard.ID.String(),
					"ownerId":     validBoard.OwnerID.String(),
					"name":        validBoard.Name.String(),
					"description": validBoard.Description.String(),
					"createdAt":   validBoard.CreatedAt.Format(timeFormat),
					"updatedAt":   validBoard.UpdatedAt.Format(timeFormat),
				},
			},
		},
		{
			name:      "No context user ID",
			inputBody: "",
			context:   context.Background(),
			wantCode:  http.StatusUnauthorized,
			wantBody:  unauthorizedTokenBody(),
		},
		{
			name:      "Internal error",
			inputBody: "",
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					return nil, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name:      "Unknown error",
			inputBody: "",
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					return nil, errors.New("unknown error")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
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

			s := &MockBoardService{}

			if tt.setupBoardService != nil {
				tt.setupBoardService(t, s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))

			h.GetMany(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
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
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("got ownerID %v, want %v", ownerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got boardID %v, want %v", boardID, validBoard.ID)
					}
					return validBoard, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]string{
				"id":          validBoard.ID.String(),
				"ownerId":     validBoard.OwnerID.String(),
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
				"createdAt":   validBoard.CreatedAt.Format(timeFormat),
				"updatedAt":   validBoard.UpdatedAt.Format(timeFormat),
			},
		},
		{
			name:     "Invalid board id",
			path:     "/v1/boards/not-a-uuid",
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name: "Not found",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
					return domain.Board{}, service.ErrBoardNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: notFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
					return domain.Board{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name: "Unknown error",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error) {
					return domain.Board{}, errors.New("unknown")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name:     "No context user ID",
			path:     okPath,
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
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

			s := &MockBoardService{}
			if tt.setupBoardService != nil {
				tt.setupBoardService(t, s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Get(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}

func TestBoards_GetAggregate(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	firstColumn := testutil.ValidColumn(validBoard.ID)
	secondColumn := testutil.NewValidColumn(t, validBoard.ID, "Done", 2)
	firstTask := testutil.ValidTask(firstColumn.ID)
	secondTask := testutil.NewValidTask(t, firstColumn.ID, "Second task", "Second description", 2)
	doneTask := testutil.ValidTask(secondColumn.ID)
	okPath := "/v1/boards/" + validBoard.ID.String() + "/aggregate"

	aggregate := service.AggregateBoard{
		Board: validBoard,
		Columns: []service.AggregateColumn{
			{
				Column: firstColumn,
				Tasks:  []domain.Task{firstTask, secondTask},
			},
			{
				Column: secondColumn,
				Tasks:  []domain.Task{doneTask},
			},
		},
	}

	tests := []boardsTestCase{
		{
			name: "Success",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetAggregateFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("got ownerID %v, want %v", ownerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got boardID %v, want %v", boardID, validBoard.ID)
					}
					return aggregate, nil
				}
			},
			wantCode: http.StatusOK,
			wantBody: map[string]any{
				"id":          validBoard.ID.String(),
				"ownerId":     validBoard.OwnerID.String(),
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
				"createdAt":   validBoard.CreatedAt.Format(timeFormat),
				"updatedAt":   validBoard.UpdatedAt.Format(timeFormat),
				"columns": []map[string]any{
					{
						"id":        firstColumn.ID.String(),
						"boardId":   firstColumn.BoardID.String(),
						"name":      firstColumn.Name.String(),
						"position":  firstColumn.Position.Int64(),
						"createdAt": firstColumn.CreatedAt.Format(timeFormat),
						"updatedAt": firstColumn.UpdatedAt.Format(timeFormat),
						"tasks": []map[string]any{
							{
								"id":          firstTask.ID.String(),
								"columnId":    firstTask.ColumnID.String(),
								"name":        firstTask.Name.String(),
								"description": firstTask.Description.String(),
								"position":    firstTask.Position.Int64(),
								"createdAt":   firstTask.CreatedAt.Format(timeFormat),
								"updatedAt":   firstTask.UpdatedAt.Format(timeFormat),
							},
							{
								"id":          secondTask.ID.String(),
								"columnId":    secondTask.ColumnID.String(),
								"name":        secondTask.Name.String(),
								"description": secondTask.Description.String(),
								"position":    secondTask.Position.Int64(),
								"createdAt":   secondTask.CreatedAt.Format(timeFormat),
								"updatedAt":   secondTask.UpdatedAt.Format(timeFormat),
							},
						},
					},
					{
						"id":        secondColumn.ID.String(),
						"boardId":   secondColumn.BoardID.String(),
						"name":      secondColumn.Name.String(),
						"position":  secondColumn.Position.Int64(),
						"createdAt": secondColumn.CreatedAt.Format(timeFormat),
						"updatedAt": secondColumn.UpdatedAt.Format(timeFormat),
						"tasks": []map[string]any{
							{
								"id":          doneTask.ID.String(),
								"columnId":    doneTask.ColumnID.String(),
								"name":        doneTask.Name.String(),
								"description": doneTask.Description.String(),
								"position":    doneTask.Position.Int64(),
								"createdAt":   doneTask.CreatedAt.Format(timeFormat),
								"updatedAt":   doneTask.UpdatedAt.Format(timeFormat),
							},
						},
					},
				},
			},
		},
		{
			name:     "Invalid board id",
			path:     "/v1/boards/not-a-uuid/aggregate",
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name: "Board not found",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetAggregateFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error) {
					return service.AggregateBoard{}, service.ErrBoardNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: boardNotFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetAggregateFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error) {
					return service.AggregateBoard{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name: "Unknown error",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.GetAggregateFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error) {
					return service.AggregateBoard{}, errors.New("unknown")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name:     "No context user ID",
			path:     okPath,
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tt.path, http.NoBody)
			if tt.context != nil {
				// TODO: factor out this pattern
				req = req.WithContext(tt.context)
			} else {
				req = req.WithContext(context.WithValue(req.Context(), httpschema.ContextKeyUserID, validBoard.OwnerID))
			}
			req.SetPathValue("boardId", strings.TrimSuffix(strings.TrimPrefix(tt.path, "/v1/boards/"), "/aggregate"))

			rr := httptest.NewRecorder()

			s := &MockBoardService{}
			if tt.setupBoardService != nil {
				tt.setupBoardService(t, s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.GetAggregate(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
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
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("got ownerID %v, want %v", ownerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got boardID %v, want %v", boardID, validBoard.ID)
					}
					if name == nil || *name != updatedValidBoard.Name {
						t.Errorf("got name %+v, want %+v", name, updatedValidBoard.Name)
					}
					if description == nil || *description != updatedValidBoard.Description {
						t.Errorf("got description %+v, want %+v", description, updatedValidBoard.Description)
					}
					return updatedValidBoard, nil
				}
			},
			wantBody: map[string]string{
				"id":          updatedValidBoard.ID.String(),
				"ownerId":     updatedValidBoard.OwnerID.String(),
				"name":        updatedValidBoard.Name.String(),
				"description": updatedValidBoard.Description.String(),
				"createdAt":   updatedValidBoard.CreatedAt.Format(timeFormat),
				"updatedAt":   updatedValidBoard.UpdatedAt.Format(timeFormat),
			},
			wantCode: http.StatusOK,
		},
		{
			name: "Empty name",
			path: okPath,
			inputBody: map[string]string{
				"name":        "",
				"description": validBoard.Description.String(),
			},
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("name", []string{"Name is too short"}),
		},
		{
			name: "Success (name only)",
			path: okPath,
			inputBody: map[string]string{
				"name": updatedNameOnlyBoard.Name.String(),
			},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name == nil || *name != updatedNameOnlyBoard.Name {
						t.Errorf("got name %+v, want %+v", name, updatedNameOnlyBoard.Name)
					}
					if description != nil {
						t.Errorf("got description %+v, want nil", description)
					}
					return updatedNameOnlyBoard, nil
				}
			},
			wantBody: map[string]any{
				"id":          updatedNameOnlyBoard.ID.String(),
				"ownerId":     updatedNameOnlyBoard.OwnerID.String(),
				"name":        updatedNameOnlyBoard.Name.String(),
				"description": updatedNameOnlyBoard.Description.String(),
				"createdAt":   updatedNameOnlyBoard.CreatedAt.Format(timeFormat),
				"updatedAt":   updatedNameOnlyBoard.UpdatedAt.Format(timeFormat),
			},
			wantCode: http.StatusOK,
		},
		{
			name: "Success (description only)",
			path: okPath,
			inputBody: map[string]string{
				"description": updatedDescriptionOnlyBoard.Description.String(),
			},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name != nil {
						t.Errorf("got name %+v, want nil", name)
					}
					if description == nil || *description != updatedDescriptionOnlyBoard.Description {
						t.Errorf("got description %+v, want %+v", description, updatedDescriptionOnlyBoard.Description)
					}
					return updatedDescriptionOnlyBoard, nil
				}
			},
			wantBody: map[string]any{
				"id":          updatedDescriptionOnlyBoard.ID.String(),
				"ownerId":     updatedDescriptionOnlyBoard.OwnerID.String(),
				"name":        updatedDescriptionOnlyBoard.Name.String(),
				"description": updatedDescriptionOnlyBoard.Description.String(),
				"createdAt":   updatedDescriptionOnlyBoard.CreatedAt.Format(timeFormat),
				"updatedAt":   updatedDescriptionOnlyBoard.UpdatedAt.Format(timeFormat),
			},
			wantCode: http.StatusOK,
		},
		{
			name:      "Success (null fields mean skip)",
			path:      okPath,
			inputBody: map[string]any{"name": nil, "description": nil},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name != nil || description != nil {
						t.Errorf("got name %+v, description %+v, want nil, nil", name, description)
					}
					return validBoard, nil
				}
			},
			wantBody: map[string]any{
				"id":          validBoard.ID.String(),
				"ownerId":     validBoard.OwnerID.String(),
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
				"createdAt":   validBoard.CreatedAt.Format(timeFormat),
				"updatedAt":   validBoard.UpdatedAt.Format(timeFormat),
			},
			wantCode: http.StatusOK,
		},
		{
			name: "Empty description",
			path: okPath,
			inputBody: map[string]string{
				"name":        emptyDescriptionBoard.Name.String(),
				"description": "",
			},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name == nil || description == nil {
						t.Errorf("got name %+v, description %+v, want non-nil, non-nil", name, description)
					}
					return emptyDescriptionBoard, nil
				}
			},
			wantBody: map[string]any{
				"id":          emptyDescriptionBoard.ID.String(),
				"ownerId":     emptyDescriptionBoard.OwnerID.String(),
				"name":        emptyDescriptionBoard.Name.String(),
				"description": emptyDescriptionBoard.Description.String(),
				"createdAt":   emptyDescriptionBoard.CreatedAt.Format(timeFormat),
				"updatedAt":   emptyDescriptionBoard.UpdatedAt.Format(timeFormat),
			},
			wantCode: http.StatusOK,
		},
		{
			name:     "Invalid board id",
			path:     "/v1/boards/not-a-uuid",
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name: "Not found",
			path: okPath,
			inputBody: map[string]string{
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
			},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, service.ErrBoardNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: notFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: okPath,
			inputBody: map[string]string{
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
			},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name: "Unknown error",
			path: okPath,
			inputBody: map[string]string{
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
			},
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.UpdateByIDFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("unknown")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name: "No context user ID",
			path: okPath,
			inputBody: map[string]string{
				"name":        validBoard.Name.String(),
				"description": validBoard.Description.String(),
			},
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
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

			s := &MockBoardService{}
			if tt.setupBoardService != nil {
				tt.setupBoardService(t, s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.UpdateByID(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			testutil.AssertContentType(t, rr, "application/json")
			testutil.AssertResponseBody(t, rr, tt.wantBody)
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
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.DeleteFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
					if ownerID != validBoard.OwnerID {
						t.Errorf("got ownerID %v, want %v", ownerID, validBoard.OwnerID)
					}
					if boardID != validBoard.ID {
						t.Errorf("got boardID %v, want %v", boardID, validBoard.ID)
					}
					return nil
				}
			},
			wantCode: http.StatusNoContent,
			wantBody: nil,
		},
		{
			name:     "Invalid board id",
			path:     "/v1/boards/not-a-uuid",
			wantCode: http.StatusBadRequest,
			wantBody: validationErrorBody("boardId", []string{"Invalid board id"}),
		},
		{
			name: "Not found",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.DeleteFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
					return service.ErrBoardNotFound
				}
			},
			wantCode: http.StatusNotFound,
			wantBody: notFoundErrorBody(),
		},
		{
			name: "Internal error",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.DeleteFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
					return service.ErrInternal
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name: "Unknown error",
			path: okPath,
			setupBoardService: func(t *testing.T, s *MockBoardService) {
				s.DeleteFunc = func(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error {
					return errors.New("unknown")
				}
			},
			wantCode: http.StatusInternalServerError,
			wantBody: internalErrorBody(),
		},
		{
			name:     "No context user ID",
			path:     okPath,
			context:  context.Background(),
			wantCode: http.StatusUnauthorized,
			wantBody: unauthorizedTokenBody(),
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

			s := &MockBoardService{}
			if tt.setupBoardService != nil {
				tt.setupBoardService(t, s)
			}

			logger := testutil.NewTestLogger(t)
			h := handler.NewBoards(logger, s, httpschema.MustNewErrorResponder(logger, testutil.FixedTimeNowStr))
			h.Delete(rr, req)

			testutil.AssertStatusCode(t, rr, tt.wantCode)
			if tt.wantCode != http.StatusNoContent {
				testutil.AssertContentType(t, rr, "application/json")
			}
			testutil.AssertResponseBody(t, rr, tt.wantBody)
		})
	}
}
