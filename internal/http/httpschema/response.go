package httpschema

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type TimeFunc func() string

type Responder struct {
	logger *slog.Logger
	timeFn TimeFunc
}

func NewResponder(logger *slog.Logger, timeFn TimeFunc) *Responder {
	if timeFn == nil {
		timeFn = func() string { return time.Now().Format(time.RFC3339) }
	}
	return &Responder{logger: logger, timeFn: timeFn}
}

func RespondJSON(w http.ResponseWriter, logger *slog.Logger, code int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		logger.Warn("Failed to send response:", slog.String("err", err.Error()))
	}
}

func (re *Responder) BadRequest(w http.ResponseWriter, errCode string, details []Detail) {
	RespondJSON(w, re.logger, http.StatusBadRequest, re.NewDetailedError(errCode, details))
}

func (re *Responder) Unauthorized(w http.ResponseWriter, errCode string, details []Detail) {
	RespondJSON(w, re.logger, http.StatusUnauthorized, re.NewDetailedError(errCode, details))
}

func (re *Responder) Error(w http.ResponseWriter, statusCode int, code string) {
	RespondJSON(w, re.logger, statusCode, re.NewError(code, MapCodeToDescription(code)))
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

func (re *Responder) NewError(code, message string) *Error {
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

func (re *Responder) NewDetailedError(code string, details []Detail) *DetailedError {
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
