package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestColumn_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validName, err := domain.NewColumnName("In Progress")
	if err != nil {
		t.Fatalf("NewColumnName: %v", err)
	}
	validPosition, err := domain.NewColumnPosition(3)
	if err != nil {
		t.Fatalf("NewColumnPosition: %v", err)
	}
	validColumn := domain.Column{
		ID:        domain.NewColumnID(),
		BoardID:   validBoard.ID,
		Name:      validName,
		Position:  validPosition,
		CreatedAt: testutil.FixedTimeNow(),
		UpdatedAt: testutil.FixedTimeNow(),
	}

	tests := []struct {
		name        string
		callerID    domain.UserID
		setupBoards func(r *MockBoardRepository)
		setupColumn func(r *MockColumnRepository)
		expectedErr error
		wantColumn  domain.Column
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, id)
					}
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.CreateFunc = func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, createdAt, updatedAt time.Time) (domain.Column, error) {
					if boardID != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, boardID)
					}
					if name != validName {
						t.Errorf("expected name %v, got %v", validName, name)
					}
					if !createdAt.Equal(testutil.FixedTimeNow()) || !updatedAt.Equal(testutil.FixedTimeNow()) {
						t.Errorf("expected service timestamp %v, got created=%v updated=%v", testutil.FixedTimeNow(), createdAt, updatedAt)
					}
					return validColumn, nil
				}
			},
			wantColumn: validColumn,
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.CreateFunc = func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, createdAt, updatedAt time.Time) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.CreateFunc = func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, createdAt, updatedAt time.Time) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Create internal error",
			callerID: validBoard.OwnerID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.CreateFunc = func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, createdAt, updatedAt time.Time) (domain.Column, error) {
					return domain.Column{}, errors.New("insert failed")
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoards(boardRepo)
			tt.setupColumn(columnRepo)

			s := service.NewColumn(columnRepo, boardRepo, testutil.FixedTimeNow)
			got, err := s.Create(context.Background(), tt.callerID, validBoard.ID, validName)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && !reflect.DeepEqual(tt.wantColumn, got) {
				t.Errorf("Create() column = %#v, want %#v", got, tt.wantColumn)
			}
		})
	}
}

func TestColumn_List(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	first := testutil.ValidColumn(validBoard.ID)
	second := testutil.ValidColumn(validBoard.ID)
	second.Position, _ = domain.NewColumnPosition(first.Position.Int64() + 1)

	tests := []struct {
		name        string
		callerID    domain.UserID
		setupBoards func(r *MockBoardRepository)
		setupColumn func(r *MockColumnRepository)
		expectedErr error
		wantColumns []domain.Column
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.ListByBoardIDFunc = func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
					if boardID != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, boardID)
					}
					return []domain.Column{first, second}, nil
				}
			},
			wantColumns: []domain.Column{first, second},
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.ListByBoardIDFunc = func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
					t.Fatalf("should not be called")
					return nil, nil
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "No access",
			callerID: domain.NewUserID(),
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.ListByBoardIDFunc = func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
					t.Fatalf("should not be called")
					return nil, nil
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Repository error",
			callerID: validBoard.OwnerID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.ListByBoardIDFunc = func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
					return nil, errors.New("db failed")
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoards(boardRepo)
			tt.setupColumn(columnRepo)

			s := service.NewColumn(columnRepo, boardRepo, testutil.FixedTimeNow)
			got, err := s.List(context.Background(), tt.callerID, validBoard.ID)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && !reflect.DeepEqual(tt.wantColumns, got) {
				t.Errorf("List() columns = %#v, want %#v", got, tt.wantColumns)
			}
		})
	}
}

func TestColumn_UpdateByID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	updatedName, err := domain.NewColumnName("Renamed")
	if err != nil {
		t.Fatalf("NewColumnName: %v", err)
	}
	updatedColumn := validColumn
	updatedColumn.Name = updatedName
	updatedColumn.UpdatedAt = testutil.FixedTimeNow()

	tests := []struct {
		name        string
		callerID    domain.UserID
		columnID    domain.ColumnID
		patchName   *domain.ColumnName
		setupBoards func(r *MockBoardRepository)
		setupColumn func(r *MockColumnRepository)
		expectedErr error
		wantColumn  domain.Column
	}{
		{
			name:      "Success",
			callerID:  validBoard.OwnerID,
			columnID:  validColumn.ID,
			patchName: &updatedName,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.GetByIDFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					if columnID != validColumn.ID {
						t.Errorf("expected column id %v, got %v", validColumn.ID, columnID)
					}
					return validColumn, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, updatedAt time.Time) (domain.Column, error) {
					if boardID != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, boardID)
					}
					if columnID != validColumn.ID {
						t.Errorf("expected column id %v, got %v", validColumn.ID, columnID)
					}
					if name == nil || *name != updatedName {
						t.Errorf("expected name %v, got %+v", updatedName, name)
					}
					if !updatedAt.Equal(testutil.FixedTimeNow()) {
						t.Errorf("expected updatedAt %v, got %v", testutil.FixedTimeNow(), updatedAt)
					}
					return updatedColumn, nil
				}
			},
			wantColumn: updatedColumn,
		},
		{
			name:     "Success no-op patch",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.GetByIDFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, updatedAt time.Time) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
			},
			wantColumn: validColumn,
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.GetByIDFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, updatedAt time.Time) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
			},
			expectedErr: service.ErrColumnNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.GetByIDFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, updatedAt time.Time) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
			},
			expectedErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column does not belong to board",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.GetByIDFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					otherBoardColumn := validColumn
					otherBoardColumn.BoardID = domain.NewBoardID()
					return otherBoardColumn, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, updatedAt time.Time) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
			},
			expectedErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.GetByIDFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return domain.Column{}, repository.ErrRowNotFound
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, updatedAt time.Time) (domain.Column, error) {
					t.Fatalf("should not be called")
					return domain.Column{}, nil
				}
			},
			expectedErr: service.ErrColumnNotFound,
		},
		{
			name:      "Update internal error",
			callerID:  validBoard.OwnerID,
			columnID:  validColumn.ID,
			patchName: &updatedName,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.GetByIDFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, updatedAt time.Time) (domain.Column, error) {
					return domain.Column{}, errors.New("update failed")
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoards(boardRepo)
			tt.setupColumn(columnRepo)

			s := service.NewColumn(columnRepo, boardRepo, testutil.FixedTimeNow)
			got, err := s.UpdateByID(context.Background(), tt.callerID, validBoard.ID, tt.columnID, tt.patchName)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && !reflect.DeepEqual(tt.wantColumn, got) {
				t.Errorf("UpdateByID() column = %#v, want %#v", got, tt.wantColumn)
			}
		})
	}
}

func TestColumn_Delete(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)

	tests := []struct {
		name        string
		callerID    domain.UserID
		columnID    domain.ColumnID
		setupBoards func(r *MockBoardRepository)
		setupColumn func(r *MockColumnRepository)
		expectedErr error
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, id)
					}
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					if boardID != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, boardID)
					}
					if columnID != validColumn.ID {
						t.Errorf("expected column id %v, got %v", validColumn.ID, columnID)
					}
					return nil
				}
			},
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					t.Fatalf("should not be called")
					return nil
				}
			},
			expectedErr: service.ErrColumnNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					t.Fatalf("should not be called")
					return nil
				}
			},
			expectedErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					return repository.ErrRowNotFound
				}
			},
			expectedErr: service.ErrColumnNotFound,
		},
		{
			name:     "Delete internal error",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoards: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumn: func(r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					return errors.New("delete failed")
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoards(boardRepo)
			tt.setupColumn(columnRepo)

			s := service.NewColumn(columnRepo, boardRepo, testutil.FixedTimeNow)
			err := s.Delete(context.Background(), tt.callerID, validBoard.ID, tt.columnID)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
