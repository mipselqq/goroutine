package service_test

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/service"
	"goroutine/internal/testutil"
)

func TestBoard_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []struct {
		name        string
		setupMock   func(r *MockBoardRepository)
		expectedErr error
		wantBoard   domain.Board
	}{
		{
			name: "Success",
			setupMock: func(r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("expected ownerID %v, got %v", validBoard.OwnerID, ownerID)
					}
					if name != validBoard.Name {
						t.Errorf("expected name %v, got %v", validBoard.Name, name)
					}
					if description != validBoard.Description {
						t.Errorf("expected description %v, got %v", validBoard.Description, description)
					}
					return validBoard, nil
				}
			},
			expectedErr: nil,
			wantBoard:   validBoard,
		},
		{
			name: "Internal error",
			setupMock: func(r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, repository.ErrInternal
				}
			},
			expectedErr: service.ErrInternal,
		},
		{
			name: "Unexpected error",
			setupMock: func(r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("unexpected error")
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupMock(r)
			s := service.NewBoard(r)

			got, err := s.Create(context.Background(), validBoard.OwnerID, validBoard.Name, validBoard.Description)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && !reflect.DeepEqual(tt.wantBoard, got) {
				t.Errorf("Create() board = %#v, want %#v", got, tt.wantBoard)
			}
		})
	}
}

func TestBoard_GetMany(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []struct {
		name        string
		setupMock   func(r *MockBoardRepository)
		expectedErr error
		wantBoards  []domain.Board
	}{
		{
			name: "Success",
			setupMock: func(r *MockBoardRepository) {
				r.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					if ownerID != validBoard.OwnerID {
						t.Errorf("expected ownerID %v, got %v", validBoard.OwnerID, ownerID)
					}
					return []domain.Board{validBoard}, nil
				}
			},
			expectedErr: nil,
			wantBoards:  []domain.Board{validBoard},
		},
		{
			name: "Internal error",
			setupMock: func(r *MockBoardRepository) {
				r.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					return nil, repository.ErrInternal
				}
			},
			expectedErr: service.ErrInternal,
		},
		{
			name: "Unexpected error",
			setupMock: func(r *MockBoardRepository) {
				r.GetManyFunc = func(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error) {
					return nil, errors.New("unexpected error")
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupMock(r)
			s := service.NewBoard(r)

			got, err := s.GetMany(context.Background(), validBoard.OwnerID)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && !reflect.DeepEqual(tt.wantBoards, got) {
				t.Errorf("GetMany() = %#v, want %#v", got, tt.wantBoards)
			}
		})
	}
}

type boardServiceGetTestCase struct {
	name        string
	callerID    domain.UserID
	setupMock   func(r *MockBoardRepository)
	expectedErr error
	wantBoard   domain.Board
}

func TestBoard_Get(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	otherOwner := domain.NewUserID()

	tests := []boardServiceGetTestCase{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("expected board id %v, got %v", validBoard.ID, id)
					}
					return validBoard, nil
				}
			},
			expectedErr: nil,
			wantBoard:   validBoard,
		},
		{
			name:     "Not found when not owner",
			callerID: otherOwner,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Not found when row missing",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Internal error from repository",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrInternal
				}
			},
			expectedErr: service.ErrInternal,
		},
		{
			name:     "Unexpected error from repository",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, errors.New("db exploded")
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupMock(r)
			s := service.NewBoard(r)

			got, err := s.Get(context.Background(), tt.callerID, validBoard.ID)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && !reflect.DeepEqual(tt.wantBoard, got) {
				t.Errorf("Get() board = %#v, want %#v", got, tt.wantBoard)
			}
		})
	}
}

func TestBoard_UpdateById(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	updatedValidBoard := testutil.UpdateValidBoard(t, &validBoard, "Updated Board Name", "Updated Board Description")
	otherOwner := domain.NewUserID()

	tests := []struct {
		name        string
		callerID    domain.UserID
		setupMock   func(r *MockBoardRepository)
		expectedErr error
		wantBoard   domain.Board
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					if boardID != validBoard.ID {
						t.Errorf("unexpected boardID %v", boardID)
					}
					if name != updatedValidBoard.Name {
						t.Errorf("unexpected name %v", name)
					}
					if description != updatedValidBoard.Description {
						t.Errorf("unexpected description %v", description)
					}
					return updatedValidBoard, nil
				}

				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("unexpected boardID %v", id)
					}
					return validBoard, nil
				}
			},
			expectedErr: nil,
			wantBoard:   updatedValidBoard,
		},
		{
			name:     "Not found when wrong owner",
			callerID: otherOwner,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Not found when row missing",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Internal error",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, repository.ErrInternal
				}
			},
			expectedErr: service.ErrInternal,
		},
		{
			name:     "Unexpected error",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
				r.UpdateByIDFunc = func(ctx context.Context, boardID domain.BoardID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error) {
					return domain.Board{}, errors.New("db exploded")
				}
			},
			expectedErr: service.ErrInternal,
		},
		{
			name:     "Unexpected get by id error",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, errors.New("db exploded")
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupMock(r)
			s := service.NewBoard(r)

			got, err := s.UpdateById(context.Background(), tt.callerID, validBoard.ID, updatedValidBoard.Name, updatedValidBoard.Description)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
			if tt.expectedErr == nil && !reflect.DeepEqual(tt.wantBoard, got) {
				t.Errorf("UpdateById() board = %#v, want %#v", got, tt.wantBoard)
			}
		})
	}
}

func TestBoard_Delete(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()
	otherOwner := domain.NewUserID()

	tests := []struct {
		name        string
		callerID    domain.UserID
		setupMock   func(r *MockBoardRepository)
		expectedErr error
	}{
		{
			name:     "Success",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					if id != validBoard.ID {
						t.Errorf("unexpected boardID %v", id)
					}
					return validBoard, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID) error {
					if boardID != validBoard.ID {
						t.Errorf("unexpected boardID %v", boardID)
					}
					return nil
				}
			},
			expectedErr: nil,
		},
		{
			name:     "Not found when wrong owner",
			callerID: otherOwner,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Not found when row missing",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrRowNotFound
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Internal error from repository",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, repository.ErrInternal
				}
			},
			expectedErr: service.ErrInternal,
		},
		{
			name:     "Unexpected error from repository",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return domain.Board{}, errors.New("db exploded")
				}
			},
			expectedErr: service.ErrInternal,
		},
		{
			name:     "Delete returns not found",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID) error {
					return repository.ErrRowNotFound
				}
			},
			expectedErr: service.ErrBoardNotFound,
		},
		{
			name:     "Delete returns internal",
			callerID: validBoard.OwnerID,
			setupMock: func(r *MockBoardRepository) {
				r.GetByIDFunc = func(ctx context.Context, id domain.BoardID) (domain.Board, error) {
					return validBoard, nil
				}
				r.DeleteFunc = func(ctx context.Context, boardID domain.BoardID) error {
					return repository.ErrInternal
				}
			},
			expectedErr: service.ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &MockBoardRepository{}
			tt.setupMock(r)
			s := service.NewBoard(r)

			err := s.Delete(context.Background(), tt.callerID, validBoard.ID)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
