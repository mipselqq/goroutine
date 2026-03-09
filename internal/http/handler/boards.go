package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
)

type BoardsService interface {
	Create(ctx context.Context, ownerID domain.UserID, name domain.BoardName, description domain.BoardDescription) (domain.Board, error)
}

type Boards struct {
	logger    *slog.Logger
	service   BoardsService
	responder *httpschema.ErrorResponder
}

func NewBoards(logger *slog.Logger, service BoardsService, responder *httpschema.ErrorResponder) *Boards {
	return &Boards{logger: logger, service: service, responder: responder}
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

	userID, ok := r.Context().Value(httpschema.ContextKeyUserID).(domain.UserID)
	if !ok {
		h.logger.Error("UserID not found in context")
		h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
		return
	}

	board, err := h.service.Create(r.Context(), userID, name, description)
	if err != nil {
		h.logger.Error("Failed to create board", slog.String("err", err.Error()))
		h.responder.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, boardResponse{
		ID:          board.ID.String(),
		OwnerID:     board.OwnerID.String(),
		Name:        board.Name.String(),
		Description: board.Description.String(),
		CreatedAt:   board.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   board.UpdatedAt.Format(time.RFC3339),
	})
}
