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
