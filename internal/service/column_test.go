package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestColumn_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)

	tests := []struct {
		name            string
		callerID        domain.UserID
		setupBoardRepo  func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo func(t *testing.T, r *MockColumnRepository)
		wantErr         error
		wantColumn      domain.Column
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
				r.CreateFunc = func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
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
			wantColumn: validColumn,
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
				r.CreateFunc = func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			wantErr: service.ErrBoardNotFound,
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
				r.CreateFunc = func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			wantErr: service.ErrBoardNotFound,
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
				r.CreateFunc = func(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
					return domain.Column{}, errors.New("insert failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)

			s := service.NewColumn(columnRepo, boardRepo)
			got, err := s.Create(context.Background(), tt.callerID, validBoard.ID, validColumn.Name, validColumn.Description)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantColumn, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("Create() column mismatch (-want +got):\n%s", diff)
				}
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
		name            string
		callerID        domain.UserID
		setupBoardRepo  func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo func(t *testing.T, r *MockColumnRepository)
		wantErr         error
		wantColumns     []domain.Column
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
				r.ListFunc = func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					return []domain.Column{first, second}, nil
				}
			},
			wantColumns: []domain.Column{first, second},
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
				r.ListFunc = func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
					t.Fatalf("got call, want no call")
					return nil, nil
				}
			},
			wantErr: service.ErrBoardNotFound,
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
				r.ListFunc = func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
					t.Fatalf("got call, want no call")
					return nil, nil
				}
			},
			wantErr: service.ErrBoardNotFound,
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
				r.ListFunc = func(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error) {
					return nil, errors.New("db failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)

			s := service.NewColumn(columnRepo, boardRepo)
			got, err := s.List(context.Background(), tt.callerID, validBoard.ID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantColumns, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("List() columns mismatch (-want +got):\n%s", diff)
				}
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
		t.Fatalf("NewColumnName() error = %v", err)
	}
	updatedColumn := validColumn
	updatedColumn.Name = updatedName
	updatedColumn.UpdatedAt = testutil.FixedTimeNow()

	updatedDesc, errDesc := domain.NewColumnDescription("Patched column description")
	if errDesc != nil {
		t.Fatalf("NewColumnDescription() error = %v", errDesc)
	}
	updatedColumnDescOnly := validColumn
	updatedColumnDescOnly.Description = updatedDesc
	updatedColumnDescOnly.UpdatedAt = testutil.FixedTimeNow()

	tests := []struct {
		name             string
		callerID         domain.UserID
		columnID         domain.ColumnID
		patchName        *domain.ColumnName
		patchDescription *domain.ColumnDescription
		setupBoardRepo   func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo  func(t *testing.T, r *MockColumnRepository)
		wantErr          error
		wantColumn       domain.Column
	}{
		{
			name:      "Success",
			callerID:  validBoard.OwnerID,
			columnID:  validColumn.ID,
			patchName: &updatedName,
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
				r.UpdateFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
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
			wantColumn: updatedColumn,
		},
		{
			name:             "Success description only",
			callerID:         validBoard.OwnerID,
			columnID:         validColumn.ID,
			patchDescription: &updatedDesc,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
				r.UpdateFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					if name != nil {
						t.Errorf("got name %+v, want nil", name)
					}
					if description == nil || *description != updatedDesc {
						t.Errorf("got description %v, want %v", description, updatedDesc)
					}
					return updatedColumnDescOnly, nil
				}
			},
			wantColumn: updatedColumnDescOnly,
		},
		{
			name:     "Success no-op patch",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
				r.UpdateFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			wantColumn: validColumn,
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
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
				r.UpdateFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			columnID: validColumn.ID,
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
				r.UpdateFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column does not belong to board",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
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
				r.UpdateFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return domain.Column{}, repository.ErrRowNotFound
				}
				r.UpdateFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					t.Fatalf("got call, want no call")
					return domain.Column{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:      "Update internal error",
			callerID:  validBoard.OwnerID,
			columnID:  validColumn.ID,
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
				r.UpdateFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error) {
					return domain.Column{}, errors.New("update failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)

			s := service.NewColumn(columnRepo, boardRepo)
			got, err := s.Update(context.Background(), tt.callerID, validBoard.ID, tt.columnID, tt.patchName, tt.patchDescription)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantColumn, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("UpdateByID() column mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestColumn_Move(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)
	targetPosition, err := domain.NewColumnPosition(2)
	if err != nil {
		t.Fatalf("NewColumnPosition() error = %v", err)
	}

	tests := []struct {
		name            string
		callerID        domain.UserID
		columnID        domain.ColumnID
		setupBoardRepo  func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo func(t *testing.T, r *MockColumnRepository)
		wantErr         error
		wantPosition    domain.ColumnPosition
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
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
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, gotTargetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
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
			wantPosition: targetPosition,
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
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
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					t.Fatalf("got call, want no call")
					return domain.ColumnPosition{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			columnID: validColumn.ID,
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
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					t.Fatalf("got call, want no call")
					return domain.ColumnPosition{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column belongs to another board",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
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
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					t.Fatalf("got call, want no call")
					return domain.ColumnPosition{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return domain.Column{}, repository.ErrRowNotFound
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					t.Fatalf("got call, want no call")
					return domain.ColumnPosition{}, nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Index out of bounds",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, repository.ErrIndexOutOfBounds
				}
			},
			wantErr: service.ErrIndexOutOfBounds,
		},
		{
			name:     "Move row not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Move internal error",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.GetFunc = func(ctx context.Context, columnID domain.ColumnID) (domain.Column, error) {
					return validColumn, nil
				}
				r.MoveFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error) {
					return domain.ColumnPosition{}, errors.New("move failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)

			s := service.NewColumn(columnRepo, boardRepo)
			got, err := s.Move(context.Background(), tt.callerID, validBoard.ID, tt.columnID, targetPosition)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got != tt.wantPosition {
				t.Errorf("Move() position = %v, want %v", got, tt.wantPosition)
			}
		})
	}
}

func TestColumn_Delete(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	validColumn := testutil.ValidColumn(validBoard.ID)

	tests := []struct {
		name            string
		callerID        domain.UserID
		columnID        domain.ColumnID
		setupBoardRepo  func(t *testing.T, r *MockBoardRepository)
		setupColumnRepo func(t *testing.T, r *MockColumnRepository)
		wantErr         error
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("got board id %v, want %v", id, validBoard.ID)
					}
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					if boardID != validBoard.ID {
						t.Errorf("got board id %v, want %v", boardID, validBoard.ID)
					}
					if columnID != validColumn.ID {
						t.Errorf("got column id %v, want %v", columnID, validColumn.ID)
					}
					return nil
				}
			},
		},
		{
			name:     "Board not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					t.Fatalf("got call, want no call")
					return nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Caller has no access",
			callerID: domain.NewUserID(),
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					t.Fatalf("got call, want no call")
					return nil
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Column not found",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					return repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrColumnNotFound,
		},
		{
			name:     "Delete internal error",
			callerID: validBoard.OwnerID,
			columnID: validColumn.ID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			setupColumnRepo: func(t *testing.T, r *MockColumnRepository) {
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error {
					return errors.New("delete failed")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			boardRepo := &MockBoardRepository{}
			columnRepo := &MockColumnRepository{}
			tt.setupBoardRepo(t, boardRepo)
			tt.setupColumnRepo(t, columnRepo)

			s := service.NewColumn(columnRepo, boardRepo)
			err := s.Delete(context.Background(), tt.callerID, validBoard.ID, tt.columnID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}
