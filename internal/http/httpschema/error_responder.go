package httpschema

import (
	"log/slog"
	"net/http"
)

type timeFunc func() string

type ErrorResponder struct {
	logger *slog.Logger
	timeFn timeFunc
}

func MustNewErrorResponder(logger *slog.Logger, timeFn timeFunc) *ErrorResponder {
	if timeFn == nil {
		panic("BUG: timeFn is nil")
	}

	return &ErrorResponder{logger: logger, timeFn: timeFn}
}

func (r *ErrorResponder) InternalError(w http.ResponseWriter, req *http.Request, err error) {
	r.logger.ErrorContext(req.Context(), "Internal server error", slog.String("err", err.Error()))
	r.Error(w, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR")
}

func (r *ErrorResponder) NotFound(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusNotFound, "NOT_FOUND", details)
}

func (r *ErrorResponder) BoardNotFound(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusNotFound, "BOARD_NOT_FOUND", details)
}

func (r *ErrorResponder) ColumnNotFound(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusNotFound, "COLUMN_NOT_FOUND", details)
}

func (r *ErrorResponder) UserNotFound(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusUnauthorized, "USER_NOT_FOUND", details)
}

func (r *ErrorResponder) ValidationError(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusBadRequest, "VALIDATION_ERROR", details)
}

func (r *ErrorResponder) InvalidCredentials(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", details)
}

// This looks like a design mistake.
// For example, if login and password are both valid and user already exists, the request is not bad.
// FIXME: break API contract before the 1.0.0 release and use a better status code.
func (r *ErrorResponder) ValidButInappropriateCredentials(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusBadRequest, "INVALID_CREDENTIALS", details)
}

func (r *ErrorResponder) InvalidToken(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusUnauthorized, "INVALID_TOKEN", details)
}

func (r *ErrorResponder) InvalidAuthHeader(w http.ResponseWriter, details []Detail) {
	r.DetailedError(w, http.StatusUnauthorized, "INVALID_AUTH_HEADER", details)
}

type Error struct {
	Code      string `json:"code" example:"INVALID_CREDENTIALS"`
	Message   string `json:"message" example:"Invalid login or password"`
	Timestamp string `json:"timestamp" example:"2026-03-02T15:04:05.123Z"`
}

type DetailedError struct {
	Error
	Details []Detail `json:"details"`
}

type Detail struct {
	Field  string   `json:"field" example:"email"`
	Issues []string `json:"issues" example:"must be a valid email,too short"`
}

func (r *ErrorResponder) NewError(code, message string) *Error {
	return &Error{
		Code:      code,
		Message:   message,
		Timestamp: r.timeFn(),
	}
}

func (r *ErrorResponder) NewDetailedError(code string, details []Detail) *DetailedError {
	return &DetailedError{
		Error:   *r.NewError(code, MapCodeToDescription(code)),
		Details: details,
	}
}

func (r *ErrorResponder) Error(w http.ResponseWriter, statusCode int, code string) {
	RespondJSON(w, r.logger, statusCode, r.NewError(code, MapCodeToDescription(code)))
}

func (r *ErrorResponder) DetailedError(w http.ResponseWriter, statusCode int, code string, details []Detail) {
	RespondJSON(w, r.logger, statusCode, r.NewDetailedError(code, details))
}
