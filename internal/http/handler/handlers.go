package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"goroutine/internal/domain"
	"goroutine/internal/http/httpschema"
)

type Handlers struct {
	Auth    *Auth
	Health  *Health
	Boards  *Boards
	Columns *Columns
	Tasks   *Tasks
}

// ErrBodyTooLarge is returned by DecodeJSONLimited when the request body exceeds the limit.
var ErrBodyTooLarge = errors.New("request body too large")

func DecodeJSONLimited(r *http.Request, v any) error {
	const maxBodySize = 20 * 1024 // 20KB is the absolute max for API
	err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBodySize)).Decode(v)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return ErrBodyTooLarge
		}
		return err
	}
	return nil
}

func extractUserIDOrHandleMissing(w http.ResponseWriter,
	r *http.Request,
	logger *slog.Logger,
	responder *httpschema.ErrorResponder,
) (domain.UserID, bool) {
	valid := true

	userID, ok := r.Context().Value(httpschema.ContextKeyUserID).(domain.UserID)
	if !ok {
		valid = false
	}
	if userID.IsEmpty() {
		valid = false
	}

	if !valid {
		logger.ErrorContext(r.Context(), "BUG: valid UserID not found in context. Middleware should have handled this.")
		responder.InvalidToken(w, []httpschema.Detail{{Field: "Authorization", Issues: []string{"Invalid token"}}})
		return domain.UserID{}, false
	}

	return userID, true
}
