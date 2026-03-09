package service_test

import (
	"context"
	"errors"
	"testing"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/service"
)

func TestBoard_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setupMock   func(r *MockBoardRepository)
		expectedErr error
	}{
		{
			name: "Success",
			setupMock: func(r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) error {
					return nil
				}
			},
			expectedErr: nil,
		},

		{
			name: "Internal error",
			setupMock: func(r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) error {
					return repository.ErrInternal
				}
			},
			expectedErr: service.ErrInternal,
		},
		{
			name: "Unexpected error",
			setupMock: func(r *MockBoardRepository) {
				r.CreateFunc = func(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) error {
					return errors.New("unexpected error")
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

			err := s.Create(context.Background(), userID, boardName, boardDescription)

			if !errors.Is(err, tt.expectedErr) {
				t.Errorf("expected error %v, got %v", tt.expectedErr, err)
			}
		})
	}
}
