package handler

import (
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

func extractUserIDOrHandleMissing(
	w http.ResponseWriter,
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
