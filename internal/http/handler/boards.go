package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
)

type BoardsService interface {
	Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
	Get(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID) (domain.Board, error)
	GetMany(ctx context.Context, ownerID domain.UserID) ([]domain.Board, error)
	UpdateById(ctx context.Context, ownerID domain.UserID, boardID domain.BoardID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
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

type boardResponse struct {
	ID          string `json:"id" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a2"`
	OwnerID     string `json:"ownerId" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a2"`
	Name        string `json:"name" example:"My Todo Name"`
	Description string `json:"description" example:"My Todo Description"`
	CreatedAt   string `json:"createdAt" example:"2026-03-07T20:56:50+03:00"`
	UpdatedAt   string `json:"updatedAt" example:"2026-03-07T20:56:50+03:00"`
}

func NewBoardResponse(board *domain.Board) boardResponse {
	return boardResponse{
		ID:          board.ID.String(),
		OwnerID:     board.OwnerID.String(),
		Name:        board.Name.String(),
		Description: board.Description.String(),
		CreatedAt:   board.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   board.UpdatedAt.Format(time.RFC3339),
	}
}

type getManyBoardsResponse = []boardResponse

// TODO(refactor-1): allow manual auth header pass into swagger panel if possible

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
		h.responder.BadRequest(w, "VALIDATION_ERROR", []httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}})
		return
	}

	details := []httpschema.Detail{}
	name := httpschema.ValidateField("name", body.Name, domain.NewBoardName, &details)
	description := httpschema.ValidateField("description", body.Description, domain.NewBoardDescription, &details)
	if len(details) > 0 {
		h.responder.BadRequest(w, "VALIDATION_ERROR", details)
		return
	}

	userID, ok := ExtractUserIDOrHandleMissing(w, r, h)
	if !ok {
		return
	}

	board, err := h.service.Create(r.Context(), userID, name, description)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to create board", slog.String("err", err.Error()))
		h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
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
// @Failure 404 {object} httpschema.Error "NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId} [get]
func (h *Boards) Get(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("boardId")
	boardID, err := domain.ParseBoardID(rawID)
	if err != nil {
		h.responder.BadRequest(w, "VALIDATION_ERROR", []httpschema.Detail{{Field: "boardId", Issues: []string{"Invalid board id"}}})
		return
	}

	userID, ok := ExtractUserIDOrHandleMissing(w, r, h)
	if !ok {
		return
	}

	board, err := h.service.Get(r.Context(), userID, boardID)
	if err != nil {
		if errors.Is(err, service.ErrBoardNotFound) {
			h.responder.Error(w, http.StatusNotFound, "NOT_FOUND")
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to get board", slog.String("err", err.Error()))
		h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, NewBoardResponse(&board))
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
	userID, ok := ExtractUserIDOrHandleMissing(w, r, h)
	if !ok {
		return
	}

	boards, err := h.service.GetMany(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), "Failed to get many boards", slog.String("err", err.Error()))
		h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
		return
	}

	response := make(getManyBoardsResponse, len(boards))
	for i := range boards {
		response[i] = NewBoardResponse(&boards[i])
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, response)
}

// UpdateById godoc
// @Summary UpdateById a board by id
// @Description UpdateById board metadata for the current user (owner only)
// @Tags boards
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param body body createBoardBody true "Board details"
// @Success 200 {object} boardResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.Error "NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId} [put]
func (h *Boards) UpdateById(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("boardId")
	boardID, err := domain.ParseBoardID(rawID)
	if err != nil {
		h.responder.BadRequest(w, "VALIDATION_ERROR", []httpschema.Detail{{Field: "boardId", Issues: []string{"Invalid board id"}}})
		return
	}

	var body createBoardBody
	err = json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		h.responder.BadRequest(w, "VALIDATION_ERROR", []httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}})
		return
	}

	details := []httpschema.Detail{}
	name := httpschema.ValidateField("name", body.Name, domain.NewBoardName, &details)
	description := httpschema.ValidateField("description", body.Description, domain.NewBoardDescription, &details)
	if len(details) > 0 {
		h.responder.BadRequest(w, "VALIDATION_ERROR", details)
		return
	}

	userID, ok := ExtractUserIDOrHandleMissing(w, r, h)
	if !ok {
		return
	}

	board, err := h.service.UpdateById(r.Context(), userID, boardID, name, description)
	if err != nil {
		if errors.Is(err, service.ErrBoardNotFound) {
			h.responder.Error(w, http.StatusNotFound, "NOT_FOUND")
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to update board", slog.String("err", err.Error()))
		h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
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
// @Failure 404 {object} httpschema.Error "NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId} [delete]
func (h *Boards) Delete(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("boardId")
	boardID, err := domain.ParseBoardID(rawID)
	if err != nil {
		h.responder.BadRequest(w, "VALIDATION_ERROR", []httpschema.Detail{{Field: "boardId", Issues: []string{"Invalid board id"}}})
		return
	}

	userID, ok := ExtractUserIDOrHandleMissing(w, r, h)
	if !ok {
		return
	}

	err = h.service.Delete(r.Context(), userID, boardID)
	if err != nil {
		if errors.Is(err, service.ErrBoardNotFound) {
			h.responder.Error(w, http.StatusNotFound, "NOT_FOUND")
			return
		}
		h.logger.ErrorContext(r.Context(), "Failed to delete board", slog.String("err", err.Error()))
		// TODO(refactor-1): add more specific responder methods
		h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func ExtractUserIDOrHandleMissing(w http.ResponseWriter, r *http.Request, h *Boards) (domain.UserID, bool) {
	valid := true

	userID, ok := r.Context().Value(httpschema.ContextKeyUserID).(domain.UserID)
	if !ok {
		valid = false
	}
	if userID.IsEmpty() {
		valid = false
	}

	if !valid {
		h.logger.ErrorContext(r.Context(), "BUG: UserID not found in context. Middleware should have handled this.")
		h.responder.Unauthorized(w, "INVALID_TOKEN", []httpschema.Detail{{Field: "Authorization", Issues: []string{"Invalid token"}}})
		return domain.UserID{}, false
	}

	return userID, valid
}
