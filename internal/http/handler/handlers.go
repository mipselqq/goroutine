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
	Auth     *Auth
	Health   *Health
	Boards   *Boards
	Columns  *Columns
	Tasks    *Tasks
	User     *User
	Telegram *Telegram
}

var errBodyTooLarge = errors.New("request body too large")

func decodeJSONLimited(r *http.Request, v any) error {
	const maxBodySize = 20 * 1024 // 20KB is the absolute max for API
	err := json.NewDecoder(http.MaxBytesReader(nil, r.Body, maxBodySize)).Decode(v)
	if err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return errBodyTooLarge
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
	userID, ok := r.Context().Value(httpschema.ContextKeyUserID).(domain.UserID)
	if !ok {
		logger.ErrorContext(r.Context(), "BUG: valid UserID not found in context. Middleware should have handled this.")
		responder.InvalidToken(w, []httpschema.Detail{{Field: "Authorization", Issues: []string{"Invalid token"}}})
		return domain.UserID{}, false
	}

	return userID, true
}
