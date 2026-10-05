package driver

import (
	"errors"
	"fmt"
	"time"
)

var ErrNetwork = errors.New("telegram network error")

type ErrTelegramResponse struct {
	Status     int
	Type       string
	RetryAfter time.Time
}

func (e *ErrTelegramResponse) Error() string {
	return fmt.Sprintf("telegram response error: status %d, type %q", e.Status, e.Type)
}

func newTelegramResponseError(status int, errorType string, retryAfterSeconds int) *ErrTelegramResponse {
	retryAfter := time.Now().UTC()
	if retryAfterSeconds > 0 {
		retryAfter = retryAfter.Add(time.Duration(retryAfterSeconds) * time.Second)
	}

	return &ErrTelegramResponse{
		Status:     status,
		Type:       errorType,
		RetryAfter: retryAfter,
	}
}
