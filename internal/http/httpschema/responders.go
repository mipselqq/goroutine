package httpschema

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func RespondJSON(w http.ResponseWriter, logger *slog.Logger, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		logger.Warn("Failed to send response:", slog.String("err", err.Error()))
	}
}

func RespondBadRequest(w http.ResponseWriter, logger *slog.Logger, errCode string, details []Detail) {
	RespondJSON(w, logger, http.StatusBadRequest, NewDetailedErrorResponse(errCode, details))
}

func RespondUnauthorized(w http.ResponseWriter, logger *slog.Logger, errCode string, details []Detail) {
	RespondJSON(w, logger, http.StatusUnauthorized, NewDetailedErrorResponse(errCode, details))
}

func RespondError(w http.ResponseWriter, logger *slog.Logger, statusCode int, code string) {
	RespondJSON(w, logger, statusCode, NewErrorResponse(code, MapCodeToDescription(code)))
}

type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

type BaseErrorResponse struct {
	Code      string `json:"code" example:"INVALID_CREDENTIALS"`
	Message   string `json:"message" example:"Invalid login or password"`
	Timestamp string `json:"timestamp" example:"2026-03-02T15:04:05.123Z"`
}

type ErrorResponse struct {
	BaseErrorResponse
}

func NewErrorResponse(code, message string) *ErrorResponse {
	return &ErrorResponse{
		BaseErrorResponse: BaseErrorResponse{
			Code:      code,
			Message:   MapCodeToDescription(code),
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
}

type DetailedErrorResponse struct {
	BaseErrorResponse
	Details []Detail `json:"details" example:"[{\"field\": \"email\", \"issue\": \"must be a valid email\"}, {\"field\": \"password\", \"issue\": \"too short, min 8 characters\"}]"`
}

func NewDetailedErrorResponse(code string, details []Detail) *DetailedErrorResponse {
	return &DetailedErrorResponse{
		BaseErrorResponse: BaseErrorResponse{
			Code:      code,
			Message:   MapCodeToDescription(code),
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Details: details,
	}
}

type Detail struct {
	Field  string   `json:"field" example:"email"`
	Issues []string `json:"issues" example:"[\"must be a valid email\", \"too short, min 8 characters\"]"`
}
