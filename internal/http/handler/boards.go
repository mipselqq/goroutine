package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
)

type BoardsService interface {
	Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	Get(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error)
	GetAggregate(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (service.AggregateBoard, error)
	GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateByID(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name *domain.BoardName, description *domain.BoardDescription) (domain.Board, error)
	Delete(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) error
}

type Boards struct {
	logger    *slog.Logger
	service   BoardsService
	responder *httpschema.ErrorResponder
}

func NewBoards(logger *slog.Logger, svc BoardsService, responder *httpschema.ErrorResponder) *Boards {
	return &Boards{logger: logger, service: svc, responder: responder}
}

type createBoardBody struct {
	Name        string `json:"name" example:"My Board Name"`
	Description string `json:"description" example:"My Board Description"`
}

type updateBoardBody struct {
	Name        *string `json:"name" example:"My Board Name"`
	Description *string `json:"description" example:"My Board Description"`
}

type boardResponse struct {
	ID          string `json:"id" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a2"`
	OwnerID     string `json:"ownerId" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a2"`
	Name        string `json:"name" example:"My Todo Name"`
	Description string `json:"description" example:"My Todo Description"`
	CreatedAt   string `json:"createdAt" example:"2026-03-07T20:56:50.000+03:00"`
	UpdatedAt   string `json:"updatedAt" example:"2026-03-07T20:56:50.000+03:00"`
}

func NewBoardResponse(board *domain.Board) boardResponse {
	return boardResponse{
		ID:          board.ID.String(),
		OwnerID:     board.OwnerID.String(),
		Name:        board.Name.String(),
		Description: board.Description.String(),

		CreatedAt: service.FormatRFC3339Millis(board.CreatedAt),
		UpdatedAt: service.FormatRFC3339Millis(board.UpdatedAt),
	}
}

type getManyBoardsResponse = []boardResponse

// Create godoc
// @Summary Create a new board
// @Description Create a new board for the current user
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body createBoardBody true "Board details"
// @Success 201 {object} boardResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards [post]
func (h *Boards) Create(w http.ResponseWriter, r *http.Request) {
	var body createBoardBody

	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}})
		return
	}

	details := []httpschema.Detail{}
	name := httpschema.ValidateField("name", body.Name, domain.NewBoardName, &details)
	description := httpschema.ValidateField("description", body.Description, domain.NewBoardDescription, &details)
	if len(details) > 0 {
		h.responder.ValidationError(w, details)
		return
	}

	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	board, err := h.service.Create(r.Context(), userID, name, description)
	if err != nil {
		h.responder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusCreated, NewBoardResponse(&board))
}

// Get godoc
// @Summary Get a board by id
// @Description Get board metadata by id for the current user (owner only)
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Success 200 {object} boardResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId} [get]
func (h *Boards) Get(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("boardId")
	boardID, err := domain.ParseBoardID(rawID)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "boardId", Issues: []string{"Invalid board id"}}})
		return
	}

	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	board, err := h.service.Get(r.Context(), userID, boardID)
	if err != nil {
		if errors.Is(err, service.ErrBoardNotFound) {
			h.responder.NotFound(w, []httpschema.Detail{})
			return
		}
		h.responder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, NewBoardResponse(&board))
}

// GetAggregate godoc
// @Summary Get a board aggregate by id
// @Description Get a board with nested columns and tasks for the current user (owner only)
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Success 200 {object} map[string]any
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "BOARD_NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId}/aggregate [get]
func (h *Boards) GetAggregate(w http.ResponseWriter, r *http.Request) {
	panic("Aggregate handler not implemented")
}

// GetMany godoc
// @Summary Get many boards
// @Description Get many boards for the current user
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} getManyBoardsResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards [get]
func (h *Boards) GetMany(w http.ResponseWriter, r *http.Request) {
	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	boards, err := h.service.GetMany(r.Context(), userID)
	if err != nil {
		h.responder.InternalError(w, r, err)
		return
	}

	response := make(getManyBoardsResponse, len(boards))
	for i := range boards {
		response[i] = NewBoardResponse(&boards[i])
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, response)
}

// UpdateByID godoc
// @Summary UpdateByID a board by id
// @Description Partially update board metadata for the current user (owner only). Provided fields are updated; omitted or null fields are ignored.
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param body body updateBoardBody true "Board fields to update"
// @Success 200 {object} boardResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId} [patch]
func (h *Boards) UpdateByID(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("boardId")
	boardID, err := domain.ParseBoardID(rawID)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "boardId", Issues: []string{"Invalid board id"}}})
		return
	}

	var body updateBoardBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}})
		return
	}

	details := []httpschema.Detail{}
	var name *domain.BoardName
	if body.Name != nil {
		value := httpschema.ValidateField("name", *body.Name, domain.NewBoardName, &details)
		name = &value
	}

	var description *domain.BoardDescription
	if body.Description != nil {
		value := httpschema.ValidateField("description", *body.Description, domain.NewBoardDescription, &details)
		description = &value
	}

	if len(details) > 0 {
		h.responder.ValidationError(w, details)
		return
	}

	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	board, err := h.service.UpdateByID(r.Context(), userID, boardID, name, description)
	if err != nil {
		if errors.Is(err, service.ErrBoardNotFound) {
			h.responder.NotFound(w, []httpschema.Detail{})
			return
		}
		h.responder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, NewBoardResponse(&board))
}

// Delete godoc
// @Summary Delete a board by id
// @Description Permanently delete a board and its columns and tasks for the current user (owner only)
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId} [delete]
func (h *Boards) Delete(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("boardId")
	boardID, err := domain.ParseBoardID(rawID)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "boardId", Issues: []string{"Invalid board id"}}})
		return
	}

	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	err = h.service.Delete(r.Context(), userID, boardID)
	if err != nil {
		if errors.Is(err, service.ErrBoardNotFound) {
			h.responder.NotFound(w, []httpschema.Detail{})
			return
		}
		h.responder.InternalError(w, r, err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
