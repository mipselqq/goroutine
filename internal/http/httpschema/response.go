package httpschema

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type TimeFunc func() string

type Response struct {
	timeFn TimeFunc
}

func NewResponse(timeFn TimeFunc) *Response {
	if timeFn == nil {
		timeFn = func() string { return time.Now().Format(time.RFC3339) }
	}
	return &Response{timeFn: timeFn}
}

func RespondJSON(w http.ResponseWriter, logger *slog.Logger, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		logger.Warn("Failed to send response:", slog.String("err", err.Error()))
	}
}

func (re *Response) BadRequest(w http.ResponseWriter, logger *slog.Logger, errCode string, details []Detail) {
	RespondJSON(w, logger, http.StatusBadRequest, re.NewDetailedError(errCode, details))
}

func (re *Response) Unauthorized(w http.ResponseWriter, logger *slog.Logger, errCode string, details []Detail) {
	RespondJSON(w, logger, http.StatusUnauthorized, re.NewDetailedError(errCode, details))
}

func (re *Response) Error(w http.ResponseWriter, logger *slog.Logger, statusCode int, code string) {
	RespondJSON(w, logger, statusCode, re.NewError(code, MapCodeToDescription(code)))
}

type Status struct {
	Status string `json:"status" example:"ok"`
}

type baseError struct {
	Code      string `json:"code" example:"INVALID_CREDENTIALS"`
	Message   string `json:"message" example:"Invalid login or password"`
	Timestamp string `json:"timestamp" example:"2026-03-02T15:04:05.123Z"`
}

type Error struct {
	baseError
}

func (re *Response) NewError(code, message string) *Error {
	return &Error{
		baseError: baseError{
			Code:      code,
			Message:   message,
			Timestamp: re.timeFn(),
		},
	}
}

type DetailedError struct {
	baseError
	Details []Detail `json:"details"`
}

func (re *Response) NewDetailedError(code string, details []Detail) *DetailedError {
	return &DetailedError{
		baseError: baseError{
			Code:      code,
			Message:   MapCodeToDescription(code),
			Timestamp: re.timeFn(),
		},
		Details: details,
	}
}

type Detail struct {
	Field  string   `json:"field" example:"email"`
	Issues []string `json:"issues" example:"must be a valid email,too short"`
}
