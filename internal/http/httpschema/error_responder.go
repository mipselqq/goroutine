package httpschema

import (
	"log/slog"
	"net/http"
)

type TimeFunc func() string

type ErrorResponder struct {
	logger *slog.Logger
	timeFn TimeFunc
}

func NewErrorResponder(logger *slog.Logger, timeFn TimeFunc) *ErrorResponder {
	// TODO: ensure functions as arguments are not nil everywhere like here
	if timeFn == nil {
		panic("BUG: timeFn is nil")
	}

	return &ErrorResponder{logger: logger, timeFn: timeFn}
}

func (re *ErrorResponder) BadRequest(w http.ResponseWriter, errCode string, details []Detail) {
	RespondJSON(w, re.logger, http.StatusBadRequest, re.NewDetailedError(errCode, details))
}

func (re *ErrorResponder) Unauthorized(w http.ResponseWriter, errCode string, details []Detail) {
	RespondJSON(w, re.logger, http.StatusUnauthorized, re.NewDetailedError(errCode, details))
}

func (re *ErrorResponder) Error(w http.ResponseWriter, statusCode int, code string) {
	RespondJSON(w, re.logger, statusCode, re.NewError(code, MapCodeToDescription(code)))
}

type Status struct {
	Status string `json:"status" example:"ok"`
}

type Error struct {
	Code      string `json:"code" example:"INVALID_CREDENTIALS"`
	Message   string `json:"message" example:"Invalid login or password"`
	Timestamp string `json:"timestamp" example:"2026-03-02T15:04:05.123Z"`
}

func (re *ErrorResponder) NewError(code, message string) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Timestamp: re.timeFn(),
	}
}

type DetailedError struct {
	Error
	Details []Detail `json:"details"`
}

func (re *ErrorResponder) NewDetailedError(code string, details []Detail) *DetailedError {
	return &DetailedError{
		Error:   *re.NewError(code, MapCodeToDescription(code)),
		Details: details,
	}
}

type Detail struct {
	Field  string   `json:"field" example:"email"`
	Issues []string `json:"issues" example:"must be a valid email,too short"`
}
