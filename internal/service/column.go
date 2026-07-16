package service

import (
	"context"
	"errors"
	"fmt"

	"goroutine/internal/domain"
	"goroutine/internal/repository"
	"goroutine/internal/template"
)

type ColumnRepository interface {
	Create(ctx context.Context, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error)
	ListByBoardID(ctx context.Context, boardID domain.BoardID) ([]domain.Column, error)
	Get(ctx context.Context, columnID domain.ColumnID) (domain.Column, error)
	Update(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, name *domain.ColumnName, description *domain.ColumnDescription) (domain.Column, error)
	Move(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID, targetPosition domain.ColumnPosition) (domain.ColumnPosition, error)
	Delete(ctx context.Context, boardID domain.BoardID, columnID domain.ColumnID) error
}

type ColumnBoardRepository interface {
	Get(ctx context.Context, boardID domain.BoardID) (domain.Board, error)
}

type columnNotif interface {
	NotifUser(ctx context.Context, userID domain.UserID, notification fmt.Stringer) error
}

type Column struct {
	columnRepo   ColumnRepository
	boardRepo    ColumnBoardRepository
	notifService columnNotif
}

func NewColumn(columnRepo ColumnRepository, boardRepo ColumnBoardRepository, notifService columnNotif) *Column {
	return &Column{
		columnRepo:   columnRepo,
		boardRepo:    boardRepo,
		notifService: notifService,
	}
}

func (s *Column) Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName, description domain.ColumnDescription) (domain.Column, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrBoardNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: create get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Column{}, ErrBoardNotFound
	}

	column, err := s.columnRepo.Create(ctx, boardID, name, description)
	if err != nil {
		return domain.Column{}, fmt.Errorf("column service: create: %v: %w", err, ErrInternal)
	}
	err = s.notifService.NotifUser(ctx, callerID, template.ColumnCreateNotif{Name: column.Name})
	if err != nil {
		return domain.Column{}, fmt.Errorf("column service: create notify: %v: %w", err, ErrInternal)
	}

	return column, nil
}

func (s *Column) ListByBoardID(ctx context.Context, callerID domain.UserID, boardID domain.BoardID) ([]domain.Column, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return nil, ErrBoardNotFound
		}
		return nil, fmt.Errorf("column service: list by board id get board: %v: %w", err, ErrInternal)
	}

	if board.OwnerID != callerID {
		return nil, ErrBoardNotFound
	}

	columns, err := s.columnRepo.ListByBoardID(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("column service: list by board id: %v: %w", err, ErrInternal)
	}

	return columns, nil
}

func (s *Column) Update(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	name *domain.ColumnName,
	description *domain.ColumnDescription,
) (domain.Column, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrColumnNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: update get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.Column{}, ErrColumnNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrColumnNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: update get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return domain.Column{}, ErrColumnNotFound
	}

	if name == nil && description == nil {
		return column, nil
	}

	updated, err := s.columnRepo.Update(ctx, boardID, columnID, name, description)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.Column{}, ErrColumnNotFound
		}
		return domain.Column{}, fmt.Errorf("column service: update: %v: %w", err, ErrInternal)
	}

	if name != nil {
		err = s.notifService.NotifUser(ctx, callerID, template.ColumnRenameNotif{
			Source: column.Name,
			Target: updated.Name,
		})
		if err != nil {
			return domain.Column{}, fmt.Errorf("column service: update name notify: %v: %w", err, ErrInternal)
		}
	}
	if description != nil {
		err = s.notifService.NotifUser(ctx, callerID, template.ColumnDescriptionUpdateNotif{Name: updated.Name})
		if err != nil {
			return domain.Column{}, fmt.Errorf("column service: update description notify: %v: %w", err, ErrInternal)
		}
	}

	return updated, nil
}

func (s *Column) Delete(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
) error {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrColumnNotFound
		}
		return fmt.Errorf("column service: delete get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return ErrColumnNotFound
	}

	err = s.columnRepo.Delete(ctx, boardID, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return ErrColumnNotFound
		}
		return fmt.Errorf("column service: delete: %v: %w", err, ErrInternal)
	}
	err = s.notifService.NotifUser(ctx, callerID, template.ColumnDeleteNotif{ID: columnID})
	if err != nil {
		return fmt.Errorf("column service: delete notify: %v: %w", err, ErrInternal)
	}

	return nil
}

func (s *Column) Move(
	ctx context.Context,
	callerID domain.UserID,
	boardID domain.BoardID,
	columnID domain.ColumnID,
	targetPosition domain.ColumnPosition,
) (domain.ColumnPosition, error) {
	board, err := s.boardRepo.Get(ctx, boardID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnPosition{}, ErrColumnNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column service: move get board: %v: %w", err, ErrInternal)
	}
	if board.OwnerID != callerID {
		return domain.ColumnPosition{}, ErrColumnNotFound
	}

	column, err := s.columnRepo.Get(ctx, columnID)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnPosition{}, ErrColumnNotFound
		}
		return domain.ColumnPosition{}, fmt.Errorf("column service: move get column: %v: %w", err, ErrInternal)
	}
	if column.BoardID != boardID {
		return domain.ColumnPosition{}, ErrColumnNotFound
	}

	position, err := s.columnRepo.Move(ctx, boardID, columnID, targetPosition)
	if err != nil {
		if errors.Is(err, repository.ErrRowNotFound) {
			return domain.ColumnPosition{}, ErrColumnNotFound
		}
		if errors.Is(err, repository.ErrIndexOutOfBounds) {
			return domain.ColumnPosition{}, ErrIndexOutOfBounds
		}
		return domain.ColumnPosition{}, fmt.Errorf("column service: move: %v: %w", err, ErrInternal)
	}
	err = s.notifService.NotifUser(ctx, callerID, template.ColumnMoveNotif{
		SourcePosition: column.Position,
		TargetPosition: position,
	})
	if err != nil {
		return domain.ColumnPosition{}, fmt.Errorf("column service: move notify: %v: %w", err, ErrInternal)
	}

	return position, nil
}
