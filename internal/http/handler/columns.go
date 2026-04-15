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

type ColumnsService interface {
	Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, name domain.ColumnName) (domain.Column, error)
}

type Columns struct {
	logger    *slog.Logger
	service   ColumnsService
	responder *httpschema.ErrorResponder
}

func NewColumns(logger *slog.Logger, svc ColumnsService, responder *httpschema.ErrorResponder) *Columns {
	return &Columns{logger: logger, service: svc, responder: responder}
}

type createColumnBody struct {
	Name string `json:"name" example:"To Do"`
}

type columnResponse struct {
	ID        string `json:"id" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a2"`
	BoardID   string `json:"boardId" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a1"`
	Name      string `json:"name" example:"In Progress"`
	Position  int64  `json:"position" example:"1"`
	CreatedAt string `json:"createdAt" example:"2026-03-07T20:56:50.000+03:00"`
	UpdatedAt string `json:"updatedAt" example:"2026-03-07T20:56:50.000+03:00"`
}

func NewColumnResponse(column *domain.Column) columnResponse {
	return columnResponse{
		ID:        column.ID.String(),
		BoardID:   column.BoardID.String(),
		Name:      column.Name.String(),
		Position:  column.Position.Int64(),
		CreatedAt: service.FormatRFC3339Millis(column.CreatedAt),
		UpdatedAt: service.FormatRFC3339Millis(column.UpdatedAt),
	}
}

// Create godoc
// @Summary Create a new column
// @Description Create a new column in board for the current user. Column is appended to the end.
// @Tags columns
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param body body createColumnBody true "Column details"
// @Success 201 {object} columnResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "BOARD_NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId}/columns [post]
func (h *Columns) Create(w http.ResponseWriter, r *http.Request) {
	rawBoardID := r.PathValue("boardId")
	boardID, err := domain.ParseBoardID(rawBoardID)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "boardId", Issues: []string{"Invalid board id"}}})
		return
	}

	var body createColumnBody
	if err = json.NewDecoder(r.Body).Decode(&body); err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}})
		return
	}

	details := []httpschema.Detail{}
	name := httpschema.ValidateField("name", body.Name, domain.NewColumnName, &details)
	if len(details) > 0 {
		h.responder.ValidationError(w, details)
		return
	}

	userID, ok := h.extractUserIDOrHandleMissing(w, r)
	if !ok {
		return
	}

	column, err := h.service.Create(r.Context(), userID, boardID, name)
	if err != nil {
		if errors.Is(err, service.ErrBoardNotFound) {
			h.responder.BoardNotFound(w, []httpschema.Detail{{Field: "boardId", Issues: []string{"Board not found"}}})
			return
		}
		h.responder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusCreated, NewColumnResponse(&column))
}

func (h *Columns) extractUserIDOrHandleMissing(w http.ResponseWriter, r *http.Request) (domain.UserID, bool) {
	valid := true

	userID, ok := r.Context().Value(httpschema.ContextKeyUserID).(domain.UserID)
	if !ok {
		valid = false
	}
	if userID.IsEmpty() {
		valid = false
	}

	if !valid {
		h.logger.ErrorContext(r.Context(), "BUG: valid UserID not found in context. Middleware should have handled this.")
		h.responder.InvalidToken(w, []httpschema.Detail{{Field: "Authorization", Issues: []string{"Invalid token"}}})
		return domain.UserID{}, false
	}

	return userID, true
}
