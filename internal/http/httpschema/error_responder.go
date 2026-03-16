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

func MustNewErrorResponder(logger *slog.Logger, timeFn TimeFunc) *ErrorResponder {
	if timeFn == nil {
		panic("BUG: timeFn is nil")
	}

	return &ErrorResponder{logger: logger, timeFn: timeFn}
}

func (r *ErrorResponder) BadRequest(w http.ResponseWriter, errCode string, details []Detail) {
	RespondJSON(w, r.logger, http.StatusBadRequest, r.NewDetailedError(errCode, details))
}

func (r *ErrorResponder) Unauthorized(w http.ResponseWriter, errCode string, details []Detail) {
	RespondJSON(w, r.logger, http.StatusUnauthorized, r.NewDetailedError(errCode, details))
}

func (r *ErrorResponder) Error(w http.ResponseWriter, statusCode int, code string) {
	RespondJSON(w, r.logger, statusCode, r.NewError(code, MapCodeToDescription(code)))
}

type Error struct {
	Code      string `json:"code" example:"INVALID_CREDENTIALS"`
	Message   string `json:"message" example:"Invalid login or password"`
	Timestamp string `json:"timestamp" example:"2026-03-02T15:04:05.123Z"`
}

func (r *ErrorResponder) NewError(code, message string) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Timestamp: r.timeFn(),
	}
}

type DetailedError struct {
	Error
	Details []Detail `json:"details"`
}

func (r *ErrorResponder) NewDetailedError(code string, details []Detail) *DetailedError {
	return &DetailedError{
		Error:   *r.NewError(code, MapCodeToDescription(code)),
		Details: details,
	}
}

type Detail struct {
	Field  string   `json:"field" example:"email"`
	Issues []string `json:"issues" example:"must be a valid email,too short"`
}
