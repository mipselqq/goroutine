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

func TestBoard_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []struct {
		name           string
		setupBoardRepo func(t *testing.T, r *MockBoardRepository)
		wantErr        error
		wantBoard      domain.Board
	}{
		{
			name: "Success",
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("got ownerID %v, want %v", ownerID, validBoard.OwnerID)
					}
					if name != validBoard.Name {
						t.Errorf("got name %v, want %v", name, validBoard.Name)
					}
					if description != validBoard.Description {
						t.Errorf("got description %v, want %v", description, validBoard.Description)
					}
					return validBoard, nil
				}
			},
			wantErr:   nil,
			wantBoard: validBoard,
		},
		{
			name: "Internal error",
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, repository.ErrInternal
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name: "Unexpected error",
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("unexpected error")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupBoardRepo(t, r)
			s := service.NewBoard(r)

			got, err := s.Create(context.Background(), validBoard.OwnerID, validBoard.Name, validBoard.Description)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantBoard, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("Create() board mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestBoard_GetMany(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []struct {
		name           string
		setupBoardRepo func(t *testing.T, r *MockBoardRepository)
		wantErr        error
		wantBoards     []domain.Board
	}{
		{
			name: "Success",
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("got ownerID %v, want %v", ownerID, validBoard.OwnerID)
					}
					return []domain.Board{validBoard}, nil
				}
			},
			wantErr:    nil,
			wantBoards: []domain.Board{validBoard},
		},
		{
			name: "Internal error",
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					return nil, repository.ErrInternal
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name: "Unexpected error",
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					return nil, errors.New("unexpected error")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupBoardRepo(t, r)
			s := service.NewBoard(r)

			got, err := s.GetMany(context.Background(), validBoard.OwnerID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantBoards, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("GetMany() mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

type boardServiceGetTestCase struct {
	name           string
	callerID       domain.UserID
	setupBoardRepo func(t *testing.T, r *MockBoardRepository)
	wantErr        error
	wantBoard      domain.Board
}

func TestBoard_Get(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	otherOwner := domain.NewUserID()

	tests := []boardServiceGetTestCase{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("got board id %v, want %v", id, validBoard.ID)
					}
					return validBoard, nil
				}
			},
			wantErr:   nil,
			wantBoard: validBoard,
		},
		{
			name:     "Not found when not owner",
			callerID: otherOwner,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			wantErr: service.ErrBoardNotFound,
		},
		{
			name:     "Not found when row missing",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrBoardNotFound,
		},
		{
			name:     "Internal error from repository",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrInternal
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name:     "Unexpected error from repository",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, errors.New("db exploded")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupBoardRepo(t, r)
			s := service.NewBoard(r)

			got, err := s.Get(context.Background(), tt.callerID, validBoard.ID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantBoard, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("Get() board mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestBoard_UpdateByID(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	updatedValidBoard := testutil.UpdateValidBoard(t, &validBoard, "Updated Board Name", "Updated Board Description", testutil.FixedTime5mFromNow())
	updatedNameOnlyBoard := testutil.UpdateValidBoard(t, &validBoard, "Updated Board Name Only", validBoard.Description.String(), testutil.FixedTime5mFromNow())
	updatedDescriptionOnlyBoard := testutil.UpdateValidBoard(t, &validBoard, validBoard.Name.String(), "Updated Board Description Only", testutil.FixedTime5mFromNow())
	otherOwner := domain.NewUserID()

	updatedName := updatedValidBoard.Name
	updatedDescription := updatedValidBoard.Description
	updatedNameOnly := updatedNameOnlyBoard.Name
	updatedDescriptionOnly := updatedDescriptionOnlyBoard.Description

	tests := []struct {
		name             string
		callerID         domain.UserID
		inputName        *domain.BoardName
		inputDescription *domain.BoardDescription
		setupBoardRepo   func(t *testing.T, r *MockBoardRepository)
		wantErr          error
		wantBoard        domain.Board
	}{
		{
			name:             "Success full update",
			callerID:         validBoard.OwnerID,
			inputName:        &updatedName,
			inputDescription: &updatedDescription,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
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

				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("got boardID %v, want %v", id, validBoard.ID)
					}
					return validBoard, nil
				}
			},
			wantErr:   nil,
			wantBoard: updatedValidBoard,
		},
		{
			name:      "Success update only name",
			callerID:  validBoard.OwnerID,
			inputName: &updatedNameOnly,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name == nil || *name != updatedNameOnlyBoard.Name {
						t.Errorf("got name %+v, want %+v", name, updatedNameOnlyBoard.Name)
					}
					if description != nil {
						t.Errorf("got description %+v, want nil", description)
					}
					return updatedNameOnlyBoard, nil
				}
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			wantErr:   nil,
			wantBoard: updatedNameOnlyBoard,
		},
		{
			name:             "Success update only description",
			callerID:         validBoard.OwnerID,
			inputDescription: &updatedDescriptionOnly,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					if name != nil {
						t.Errorf("got name %+v, want nil", name)
					}
					if description == nil || *description != updatedDescriptionOnlyBoard.Description {
						t.Errorf("got description %+v, want %+v", description, updatedDescriptionOnlyBoard.Description)
					}
					return updatedDescriptionOnlyBoard, nil
				}
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			wantErr:   nil,
			wantBoard: updatedDescriptionOnlyBoard,
		},
		{
			name:     "Success update no fields",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			wantErr:   nil,
			wantBoard: validBoard,
		},
		{
			name:     "Not found when wrong owner",
			callerID: otherOwner,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			wantErr: service.ErrBoardNotFound,
		},
		{
			name:     "Not found when row missing",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrBoardNotFound,
		},
		{
			name:      "Internal error",
			callerID:  validBoard.OwnerID,
			inputName: &updatedName,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, repository.ErrInternal
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name:      "Unexpected error",
			callerID:  validBoard.OwnerID,
			inputName: &updatedName,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("db exploded")
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name:     "Unexpected get by id error",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, errors.New("db exploded")
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupBoardRepo(t, r)
			s := service.NewBoard(r)

			got, err := s.UpdateByID(context.Background(), tt.callerID, validBoard.ID, tt.inputName, tt.inputDescription)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if diff := cmp.Diff(tt.wantBoard, got, testutil.CmpAllowUnexported()); diff != "" {
					t.Errorf("UpdateByID() board mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func TestBoard_Delete(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	otherOwner := domain.NewUserID()

	tests := []struct {
		name           string
		callerID       domain.UserID
		setupBoardRepo func(t *testing.T, r *MockBoardRepository)
		wantErr        error
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("got boardID %v, want %v", id, validBoard.ID)
					}
					return validBoard, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID) error {
					if boardID != validBoard.ID {
						t.Errorf("got boardID %v, want %v", boardID, validBoard.ID)
					}
					return nil
				}
			},
			wantErr: nil,
		},
		{
			name:     "Not found when wrong owner",
			callerID: otherOwner,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			wantErr: service.ErrBoardNotFound,
		},
		{
			name:     "Not found when row missing",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrBoardNotFound,
		},
		{
			name:     "Internal error from repository",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrInternal
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name:     "Unexpected error from repository",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, errors.New("db exploded")
				}
			},
			wantErr: service.ErrInternal,
		},
		{
			name:     "Delete returns not found",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID) error {
					return repository.ErrRowNotFound
				}
			},
			wantErr: service.ErrBoardNotFound,
		},
		{
			name:     "Delete returns internal",
			callerID: validBoard.OwnerID,
			setupBoardRepo: func(t *testing.T, r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID) error {
					return repository.ErrInternal
				}
			},
			wantErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupBoardRepo(t, r)
			s := service.NewBoard(r)

			err := s.Delete(context.Background(), tt.callerID, validBoard.ID)

			if !errors.Is(err, tt.wantErr) {
				t.Errorf("got error %v, want %v", err, tt.wantErr)
			}
		})
	}
}
