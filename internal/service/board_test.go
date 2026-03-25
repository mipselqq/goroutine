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

type boardServiceCreateTestCase struct {
	name        string
	setupMock   func(r *MockBoardRepository)
	expectedErr error
	wantBoard   domain.Board
}

func TestBoard_Create(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []boardServiceCreateTestCase{
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

type boardServiceGetManyTestCase struct {
	name        string
	setupMock   func(r *MockBoardRepository)
	expectedErr error
	wantBoards  []domain.Board
}

func TestBoard_GetMany(t *testing.T) {
	t.Parallel()

	validBoard := testutil.ValidBoard()

	tests := []boardServiceGetManyTestCase{
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
