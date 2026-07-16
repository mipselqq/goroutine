package service_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/service"
	"goroutine/internal/template"
	"goroutine/internal/testutil"
)

func TestTask_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	validTask := testutil.ValidTask(validColumn.ID)
	validName := validTask.Name
	validDescription := validTask.Description

	tests := []struct {
		name            string
		callerID        domain.UserID
		setupBoardRepo  func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo func(t *testing.T, r *MockColumnRepository)
		setupTaskRepo   func(t *testing.T, r *MockTaskRepository)
		wantNotifs      []fmt.Stringer
		wantErr         error
		wantTask        domain.Task
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("got board id %v, want %v", id, validBoard.ID)
					}
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.CreateFunc = func(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					if name != validName {
						t.Errorf("got name %v, want %v", name, validName)
					}
					if description != validDescription {
						t.Errorf("got description %v, want %v", description, validDescription)
					}
					return validTask, nil
				}
			},
			wantNotifs: []fmt.Stringer{template.TaskCreateNotif{Name: validTask.Name}},
			wantTask:   validTask,
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.CreateFunc = func(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.CreateFunc = func(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column not found",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return domain.Column{}, repository.ErrRowNotFound
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.CreateFunc = func(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column belongs to another board",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					otherBoardColumn := validColumn
					otherBoardColumn.BoardID = domain.NewBoardID()
					return otherBoardColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.CreateFunc = func(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Create internal error",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.CreateFunc = func(ctx context.Context, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error) {
					return domain.Task{}, errors.New("insert failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := NewMockBoardRepository(t)
			columnRepo := NewMockColumnRepository(t)
			taskRepo := NewMockTaskRepository(t)
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)
			tt.setupTaskRepo(t, taskRepo)

			notifService := NewMockUserNotif(t, tt.callerID, tt.wantNotifs...)
			s := service.NewTask(taskRepo, boardRepo, columnRepo, notifService)
			got, err := s.Create(context.Background(), tt.callerID, validBoard.ID, validColumn.ID, validName, validDescription)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantTask, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("Create() task mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestTask_ListByColumnID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	first := testutil.ValidTask(validColumn.ID)
	second := testutil.NewValidTask(t, validColumn.ID, "Second", "second", 2)

	tests := []struct {
		name            string
		callerID        domain.UserID
		setupBoardRepo  func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo func(t *testing.T, r *MockColumnRepository)
		setupTaskRepo   func(t *testing.T, r *MockTaskRepository)
		wantErr         error
		wantTasks       []domain.Task
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.ListByColumnIDFunc = func(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					return []domain.Task{first, second}, nil
				}
			},
			wantTasks: []domain.Task{first, second},
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.ListByColumnIDFunc = func(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
					t.Fatalf("got call, want no call")
					return nil, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "No access",
			callerID: domain.NewUserID(),
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.ListByColumnIDFunc = func(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
					t.Fatalf("got call, want no call")
					return nil, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column not found",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return domain.Column{}, repository.ErrRowNotFound
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.ListByColumnIDFunc = func(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
					t.Fatalf("got call, want no call")
					return nil, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column belongs to another board",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					otherBoardColumn := validColumn
					otherBoardColumn.BoardID = domain.NewBoardID()
					return otherBoardColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.ListByColumnIDFunc = func(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
					t.Fatalf("got call, want no call")
					return nil, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Repository error",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.ListByColumnIDFunc = func(ctx context.Context, columnID domain.ColumnID) ([]domain.Task, error) {
					return nil, errors.New("db failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := NewMockBoardRepository(t)
			columnRepo := NewMockColumnRepository(t)
			taskRepo := NewMockTaskRepository(t)
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)
			tt.setupTaskRepo(t, taskRepo)

			s := service.NewTask(taskRepo, boardRepo, columnRepo, MockUserNotif{})
			got, err := s.ListByColumnID(context.Background(), tt.callerID, validBoard.ID, validColumn.ID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantTasks, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("ListByColumnID() tasks mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestTask_Update(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	validTask := testutil.ValidTask(validColumn.ID)
	updatedTask := testutil.UpdateValidTask(t, &validTask, "Renamed", "Renamed description", testutil.FixedNow())
	updatedName := updatedTask.Name
	updatedDescription := updatedTask.Description

	tests := []struct {
		name             string
		callerID         domain.UserID
		taskID           domain.TaskID
		patchName        *domain.TaskName
		patchDescription *domain.TaskDescription
		setupBoardRepo   func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo  func(t *testing.T, r *MockColumnRepository)
		setupTaskRepo    func(t *testing.T, r *MockTaskRepository)
		wantNotifs       []fmt.Stringer
		wantErr          error
		wantTask         domain.Task
	}{
		{
			name:             "Success",
			callerID:         validBoard.OwnerID,
			taskID:           validTask.ID,
			patchName:        &updatedName,
			patchDescription: &updatedDescription,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					if taskID != validTask.ID {
						t.Errorf("got task id %v, want %v", taskID, validTask.ID)
					}
					return validTask, nil
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
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
			wantNotifs: []fmt.Stringer{
				template.TaskRenameNotif{Source: validTask.Name, Target: updatedTask.Name},
				template.TaskDescriptionUpdateNotif{Name: updatedTask.Name},
			},
			wantTask: updatedTask,
		},
		{
			name:     "Success no-op patch",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantTask: validTask,
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:     "Column not found",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return domain.Column{}, repository.ErrRowNotFound
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:     "Column belongs to another board",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					otherBoardColumn := validColumn
					otherBoardColumn.BoardID = domain.NewBoardID()
					return otherBoardColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:     "Task not found",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return domain.Task{}, repository.ErrRowNotFound
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:     "Task does not belong to column",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					otherColumnTask := validTask
					otherColumnTask.ColumnID = domain.NewColumnID()
					return otherColumnTask, nil
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:      "Update internal error",
			callerID:  validBoard.OwnerID,
			taskID:    validTask.ID,
			patchName: &updatedName,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.UpdateFunc = func(ctx context.Context, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error) {
					return domain.Task{}, errors.New("update failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := NewMockBoardRepository(t)
			columnRepo := NewMockColumnRepository(t)
			taskRepo := NewMockTaskRepository(t)
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)
			tt.setupTaskRepo(t, taskRepo)

			notifService := NewMockUserNotif(t, tt.callerID, tt.wantNotifs...)
			s := service.NewTask(taskRepo, boardRepo, columnRepo, notifService)
			got, err := s.Update(context.Background(), tt.callerID, validBoard.ID, validColumn.ID, tt.taskID, tt.patchName, tt.patchDescription)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantTask, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("Update() task mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestTask_Move(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	targetColumn := testutil.NewValidColumn(t, validBoard.ID, "Done", 2)
	validTask := testutil.ValidTask(validColumn.ID)
	targetPosition := testutil.NewValidTaskPosition(t, 2)

	tests := []struct {
		name            string
		callerID        domain.UserID
		taskID          domain.TaskID
		targetColumnID  domain.ColumnID
		setupBoardRepo  func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo func(t *testing.T, r *MockColumnRepository)
		setupTaskRepo   func(t *testing.T, r *MockTaskRepository)
		wantNotifs      []fmt.Stringer
		wantErr         error
		wantColumn      domain.ColumnID
		wantPosition    domain.TaskPosition
	}{
		{
			name:           "Success same column",
			callerID:       validBoard.OwnerID,
			taskID:         validTask.ID,
			targetColumnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if currentColumnID != validColumn.ID {
						t.Errorf("got current column id %v, want %v", currentColumnID, validColumn.ID)
					}
					if taskID != validTask.ID {
						t.Errorf("got task id %v, want %v", taskID, validTask.ID)
					}
					if gotTargetColumnID != validColumn.ID {
						t.Errorf("got target column id %v, want %v", gotTargetColumnID, validColumn.ID)
					}
					if gotTargetPosition != targetPosition {
						t.Errorf("got target position %v, want %v", gotTargetPosition, targetPosition)
					}
					return validColumn.ID, targetPosition, nil
				}
			},
			wantNotifs: []fmt.Stringer{template.TaskMoveNotif{
				SourceColumnID: validColumn.ID,
				TargetColumnID: validColumn.ID,
				SourcePosition: validTask.Position,
				TargetPosition: targetPosition,
			}},
			wantColumn:   validColumn.ID,
			wantPosition: targetPosition,
		},
		{
			name:           "Success cross column",
			callerID:       validBoard.OwnerID,
			taskID:         validTask.ID,
			targetColumnID: targetColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					switch columnID {
					case validColumn.ID:
						return validColumn, nil
					case targetColumn.ID:
						return targetColumn, nil
					default:
						t.Errorf("got column id %v, want %v or %v", columnID, validColumn.ID, targetColumn.ID)
						return domain.Column{}, nil
					}
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					if gotTargetColumnID != targetColumn.ID {
						t.Errorf("got target column id %v, want %v", gotTargetColumnID, targetColumn.ID)
					}
					return targetColumn.ID, targetPosition, nil
				}
			},
			wantNotifs: []fmt.Stringer{template.TaskMoveNotif{
				SourceColumnID: validColumn.ID,
				TargetColumnID: targetColumn.ID,
				SourcePosition: validTask.Position,
				TargetPosition: targetPosition,
			}},
			wantColumn:   targetColumn.ID,
			wantPosition: targetPosition,
		},
		{
			name:           "Target column not found",
			callerID:       validBoard.OwnerID,
			taskID:         validTask.ID,
			targetColumnID: targetColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					if columnID == validColumn.ID {
						return validColumn, nil
					}
					return domain.Column{}, repository.ErrRowNotFound
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					t.Fatalf("got call, want no call")
					return domain.ColumnID{}, domain.TaskPosition{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:           "Target column belongs to another board",
			callerID:       validBoard.OwnerID,
			taskID:         validTask.ID,
			targetColumnID: targetColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					if columnID == validColumn.ID {
						return validColumn, nil
					}
					otherBoardColumn := targetColumn
					otherBoardColumn.BoardID = domain.NewBoardID()
					return otherBoardColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					t.Fatalf("got call, want no call")
					return domain.ColumnID{}, domain.TaskPosition{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:           "Index out of bounds",
			callerID:       validBoard.OwnerID,
			taskID:         validTask.ID,
			targetColumnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					return domain.ColumnID{}, domain.TaskPosition{}, repository.ErrIndexOutOfBounds
				}
			},
			wantErr: service.ErrIndexOutOfBounds,
		},
		{
			name:           "Move row not found",
			callerID:       validBoard.OwnerID,
			taskID:         validTask.ID,
			targetColumnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					return domain.ColumnID{}, domain.TaskPosition{}, repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:           "Task not found",
			callerID:       validBoard.OwnerID,
			taskID:         validTask.ID,
			targetColumnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return domain.Task{}, repository.ErrRowNotFound
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					t.Fatalf("got call, want no call")
					return domain.ColumnID{}, domain.TaskPosition{}, nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:           "Move internal error",
			callerID:       validBoard.OwnerID,
			taskID:         validTask.ID,
			targetColumnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, currentColumnID domain.ColumnID, taskID domain.TaskID, gotTargetColumnID domain.ColumnID, gotTargetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error) {
					return domain.ColumnID{}, domain.TaskPosition{}, errors.New("move failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := NewMockBoardRepository(t)
			columnRepo := NewMockColumnRepository(t)
			taskRepo := NewMockTaskRepository(t)
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)
			tt.setupTaskRepo(t, taskRepo)

			notifService := NewMockUserNotif(t, tt.callerID, tt.wantNotifs...)
			s := service.NewTask(taskRepo, boardRepo, columnRepo, notifService)
			gotColumn, gotPosition, err := s.Move(context.Background(), tt.callerID, validBoard.ID, validColumn.ID, tt.taskID, tt.targetColumnID, targetPosition)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if gotColumn != tt.wantColumn {
					t.Errorf("Move() column = %v, want %v", gotColumn, tt.wantColumn)
				}
				if gotPosition != tt.wantPosition {
					t.Errorf("Move() position = %v, want %v", gotPosition, tt.wantPosition)
				}
			}
		})
	}
}

func TestTask_Delete(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	validTask := testutil.ValidTask(validColumn.ID)

	tests := []struct {
		name            string
		callerID        domain.UserID
		taskID          domain.TaskID
		setupBoardRepo  func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo func(t *testing.T, r *MockColumnRepository)
		setupTaskRepo   func(t *testing.T, r *MockTaskRepository)
		wantNotifs      []fmt.Stringer
		wantErr         error
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
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
			wantNotifs: []fmt.Stringer{template.TaskDeleteNotif{Name: validTask.Name}},
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
					t.Fatalf("got call, want no call")
					return nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					t.Fatalf("got call, want no call")
					return domain.Task{}, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
					t.Fatalf("got call, want no call")
					return nil
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:     "Task not found",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
					return repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrTaskNotFound,
		},
		{
			name:     "Delete internal error",
			callerID: validBoard.OwnerID,
			taskID:   validTask.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
			},
			setupTaskRepo: func(t *testing.T, r *MockTaskRepository) {
				r.GetFunc = func(ctx context.Context, taskID domain.TaskID) (domain.Task, error) {
					return validTask, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error {
					return errors.New("delete failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := NewMockBoardRepository(t)
			columnRepo := NewMockColumnRepository(t)
			taskRepo := NewMockTaskRepository(t)
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)
			tt.setupTaskRepo(t, taskRepo)

			notifService := NewMockUserNotif(t, tt.callerID, tt.wantNotifs...)
			s := service.NewTask(taskRepo, boardRepo, columnRepo, notifService)
			err := s.Delete(context.Background(), tt.callerID, validBoard.ID, validColumn.ID, tt.taskID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}
