package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
	"goroutine/internal/service"
)

type TasksService interface {
	Create(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, name domain.TaskName, description domain.TaskDescription) (domain.Task, error)
	ListByColumnID(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID) ([]domain.Task, error)
	Update(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, name *domain.TaskName, description *domain.TaskDescription) (domain.Task, error)
	Move(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, targetColumnID domain.ColumnID, targetPosition domain.TaskPosition) (domain.ColumnID, domain.TaskPosition, error)
	Delete(ctx context.Context, callerID domain.UserID, boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID) error
}

type Tasks struct {
	logger       *slog.Logger
	tasksService TasksService
	responder    *httpschema.ErrorResponder
}

func NewTasks(logger *slog.Logger, tasksService TasksService, responder *httpschema.ErrorResponder) *Tasks {
	return &Tasks{logger: logger, tasksService: tasksService, responder: responder}
}

type createTaskBody struct {
	Name        string `json:"name" example:"Write tests"`
	Description string `json:"description" example:"Cover the new endpoint with tests"`
}

type updateTaskBody struct {
	Name        *string `json:"name" example:"Rewrite tests"`
	Description *string `json:"description" example:"Cover edge cases"`
}

type moveTaskBody struct {
	TargetColumnID string `json:"targetColumnId" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a2"`
	TargetPosition int64  `json:"targetPosition" example:"1"`
}

type taskResponse struct {
	ID          string `json:"id" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a3"`
	ColumnID    string `json:"columnId" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a2"`
	Name        string `json:"name" example:"Write tests"`
	Description string `json:"description" example:"Cover the new endpoint with tests"`
	Position    int64  `json:"position" example:"1"`
	CreatedAt   string `json:"createdAt" example:"2026-03-07T20:56:50.000+03:00"`
	UpdatedAt   string `json:"updatedAt" example:"2026-03-07T20:56:50.000+03:00"`
}

type taskPositionResponse struct {
	ColumnID string `json:"columnId" example:"019cc971-e5be-7df9-ae8a-c6e3f29c86a2"`
	Position int64  `json:"position" example:"2"`
}

func newTaskResponse(task *domain.Task) taskResponse {
	return taskResponse{
		ID:          task.ID.String(),
		ColumnID:    task.ColumnID.String(),
		Name:        task.Name.String(),
		Description: task.Description.String(),
		Position:    task.Position.Int64(),
		CreatedAt:   service.FormatRFC3339Millis(task.CreatedAt),
		UpdatedAt:   service.FormatRFC3339Millis(task.UpdatedAt),
	}
}

// Create godoc
// @Summary Create a new task
// @Description Create a new task in a column for the current user. Task is appended to the end of the column.
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param columnId path string true "Column ID"
// @Param body body createTaskBody true "Task details"
// @Success 201 {object} taskResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 413 {object} httpschema.DetailedError "PAYLOAD_TOO_LARGE"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "COLUMN_NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId}/columns/{columnId}/tasks [post]
func (h *Tasks) Create(w http.ResponseWriter, r *http.Request) {
	boardID, columnID, ok := h.parseBoardAndColumnID(w, r)
	if !ok {
		return
	}

	var body createTaskBody
	err := decodeJSONLimited(r, &body)
	if err != nil {
		if errors.Is(err, errBodyTooLarge) {
			h.responder.PayloadTooLarge(w)
		} else {
			h.responder.ValidationError(w, []httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}})
		}
		return
	}

	details := []httpschema.Detail{}
	name := httpschema.ValidateField("name", body.Name, domain.NewTaskName, &details)
	description := httpschema.ValidateField("description", body.Description, domain.NewTaskDescription, &details)
	if len(details) > 0 {
		h.responder.ValidationError(w, details)
		return
	}

	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	task, err := h.tasksService.Create(r.Context(), userID, boardID, columnID, name, description)
	if err != nil {
		if errors.Is(err, service.ErrColumnNotFound) {
			h.responder.ColumnNotFound(w, []httpschema.Detail{{Field: "columnId", Issues: []string{"Column not found"}}})
			return
		}
		h.responder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusCreated, newTaskResponse(&task))
}

// List godoc
// @Summary List all tasks in a column
// @Description Get all tasks belonging to the specified column. Results are returned in increasing position order.
// @Tags tasks
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param columnId path string true "Column ID"
// @Success 200 {array} taskResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "COLUMN_NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId}/columns/{columnId}/tasks [get]
func (h *Tasks) List(w http.ResponseWriter, r *http.Request) {
	boardID, columnID, ok := h.parseBoardAndColumnID(w, r)
	if !ok {
		return
	}

	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	tasks, err := h.tasksService.ListByColumnID(r.Context(), userID, boardID, columnID)
	if err != nil {
		if errors.Is(err, service.ErrColumnNotFound) {
			h.responder.ColumnNotFound(w, []httpschema.Detail{{Field: "columnId", Issues: []string{"Column not found"}}})
			return
		}
		h.responder.InternalError(w, r, err)
		return
	}

	response := make([]taskResponse, 0, len(tasks))
	for i := range tasks {
		response = append(response, newTaskResponse(&tasks[i]))
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, response)
}

// Update godoc
// @Summary Update a task by id
// @Description Partially update task metadata for the current user. Provided fields are updated; omitted or null fields are ignored.
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param columnId path string true "Column ID"
// @Param taskId path string true "Task ID"
// @Param body body updateTaskBody true "Task fields to update"
// @Success 200 {object} taskResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 413 {object} httpschema.DetailedError "PAYLOAD_TOO_LARGE"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "TASK_NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId} [patch]
func (h *Tasks) Update(w http.ResponseWriter, r *http.Request) {
	boardID, columnID, taskID, ok := h.parseBoardColumnAndTaskID(w, r)
	if !ok {
		return
	}

	var body updateTaskBody
	err := decodeJSONLimited(r, &body)
	if err != nil {
		if errors.Is(err, errBodyTooLarge) {
			h.responder.PayloadTooLarge(w)
		} else {
			h.responder.ValidationError(w, []httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}})
		}
		return
	}

	details := []httpschema.Detail{}
	var name *domain.TaskName
	if body.Name != nil {
		value := httpschema.ValidateField("name", *body.Name, domain.NewTaskName, &details)
		name = &value
	}
	var description *domain.TaskDescription
	if body.Description != nil {
		value := httpschema.ValidateField("description", *body.Description, domain.NewTaskDescription, &details)
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

	task, err := h.tasksService.Update(r.Context(), userID, boardID, columnID, taskID, name, description)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			h.responder.TaskNotFound(w, []httpschema.Detail{{Field: "taskId", Issues: []string{"Task not found"}}})
			return
		}
		h.responder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, newTaskResponse(&task))
}

// Move godoc
// @Summary Move a task to a new position, possibly to another column
// @Description Move a task within its column or across columns in the same board and shift neighboring tasks accordingly.
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param columnId path string true "Column ID"
// @Param taskId path string true "Task ID"
// @Param body body moveTaskBody true "Target column and position"
// @Success 200 {object} taskPositionResponse
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 413 {object} httpschema.DetailedError "PAYLOAD_TOO_LARGE"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "TASK_NOT_FOUND or COLUMN_NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}/position [put]
func (h *Tasks) Move(w http.ResponseWriter, r *http.Request) {
	boardID, columnID, taskID, ok := h.parseBoardColumnAndTaskID(w, r)
	if !ok {
		return
	}

	var body moveTaskBody
	err := decodeJSONLimited(r, &body)
	if err != nil {
		if errors.Is(err, errBodyTooLarge) {
			h.responder.PayloadTooLarge(w)
		} else {
			h.responder.ValidationError(w, []httpschema.Detail{{Field: "body", Issues: []string{"Invalid JSON body"}}})
		}
		return
	}

	details := []httpschema.Detail{}
	targetColumnID, err := domain.ParseColumnID(body.TargetColumnID)
	if err != nil {
		details = append(details, httpschema.Detail{Field: "targetColumnId", Issues: []string{"Invalid target column id"}})
	}
	targetPosition := httpschema.ValidateField("targetPosition", body.TargetPosition, domain.NewTaskPosition, &details)
	if len(details) > 0 {
		h.responder.ValidationError(w, details)
		return
	}

	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	newColumnID, newPosition, err := h.tasksService.Move(r.Context(), userID, boardID, columnID, taskID, targetColumnID, targetPosition)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			h.responder.TaskNotFound(w, []httpschema.Detail{{Field: "taskId", Issues: []string{"Task not found"}}})
			return
		}
		if errors.Is(err, service.ErrColumnNotFound) {
			h.responder.ColumnNotFound(w, []httpschema.Detail{{Field: "targetColumnId", Issues: []string{"Column not found"}}})
			return
		}
		if errors.Is(err, service.ErrIndexOutOfBounds) {
			h.responder.ValidationError(w, []httpschema.Detail{{Field: "targetPosition", Issues: []string{"Index out of bounds"}}})
			return
		}
		h.responder.InternalError(w, r, err)
		return
	}

	httpschema.RespondJSON(w, h.logger, http.StatusOK, taskPositionResponse{
		ColumnID: newColumnID.String(),
		Position: newPosition.Int64(),
	})
}

// Delete godoc
// @Summary Delete a task by id
// @Description Permanently delete a task from a column for the current user and shift positions to close the gap.
// @Tags tasks
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param boardId path string true "Board ID"
// @Param columnId path string true "Column ID"
// @Param taskId path string true "Task ID"
// @Success 204 "No Content"
// @Failure 400 {object} httpschema.DetailedError "VALIDATION_ERROR"
// @Failure 401 {object} httpschema.DetailedError "Unauthorized: INVALID_TOKEN or INVALID_AUTH_HEADER"
// @Failure 404 {object} httpschema.DetailedError "TASK_NOT_FOUND"
// @Failure 500 {object} httpschema.Error "Internal server error"
// @Router /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId} [delete]
func (h *Tasks) Delete(w http.ResponseWriter, r *http.Request) {
	boardID, columnID, taskID, ok := h.parseBoardColumnAndTaskID(w, r)
	if !ok {
		return
	}

	userID, ok := extractUserIDOrHandleMissing(w, r, h.logger, h.responder)
	if !ok {
		return
	}

	err := h.tasksService.Delete(r.Context(), userID, boardID, columnID, taskID)
	if err != nil {
		if errors.Is(err, service.ErrTaskNotFound) {
			h.responder.TaskNotFound(w, []httpschema.Detail{{Field: "taskId", Issues: []string{"Task not found"}}})
			return
		}
		h.responder.InternalError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Tasks) parseBoardAndColumnID(w http.ResponseWriter, r *http.Request) (boardID domain.BoardID, columnID domain.ColumnID, ok bool) {
	rawBoardID := r.PathValue("boardId")
	boardID, err := domain.ParseBoardID(rawBoardID)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "boardId", Issues: []string{"Invalid board id"}}})
		return domain.BoardID{}, domain.ColumnID{}, false
	}

	rawColumnID := r.PathValue("columnId")
	columnID, err = domain.ParseColumnID(rawColumnID)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "columnId", Issues: []string{"Invalid column id"}}})
		return domain.BoardID{}, domain.ColumnID{}, false
	}

	return boardID, columnID, true
}

func (h *Tasks) parseBoardColumnAndTaskID(w http.ResponseWriter, r *http.Request) (boardID domain.BoardID, columnID domain.ColumnID, taskID domain.TaskID, ok bool) {
	boardID, columnID, ok = h.parseBoardAndColumnID(w, r)
	if !ok {
		return domain.BoardID{}, domain.ColumnID{}, domain.TaskID{}, false
	}

	rawTaskID := r.PathValue("taskId")
	taskID, err := domain.ParseTaskID(rawTaskID)
	if err != nil {
		h.responder.ValidationError(w, []httpschema.Detail{{Field: "taskId", Issues: []string{"Invalid task id"}}})
		return domain.BoardID{}, domain.ColumnID{}, domain.TaskID{}, false
	}

	return boardID, columnID, taskID, true
}
