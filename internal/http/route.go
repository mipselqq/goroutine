package http

import (
	"net/http"

	"goroutine/internal/http/handler"
	"goroutine/internal/http/middleware"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	swaggerBasePath = "/v1/swagger/"
	loginPath       = "/v1/login"
)

func NewRouter(handlers *handler.Handlers, middlewares *middleware.Middlewares) http.Handler {
	mux := http.NewServeMux()
	public := func(h http.HandlerFunc) http.Handler {
		return middlewares.Metrics.Wrap(h)
	}
	protected := func(h http.HandlerFunc) http.Handler {
		return middlewares.Metrics.Wrap(middlewares.Auth.Wrap(h))
	}

	mux.Handle("POST /v1/register", public(handlers.Auth.Register))
	mux.Handle("POST "+loginPath, public(handlers.Auth.Login))
	mux.Handle("GET /v1/health", public(handlers.Health.Health))
	mux.Handle("GET /v1/whoami", protected(handlers.Auth.WhoAmI))
	mux.Handle("POST /v1/users/me/telegram/link", protected(handlers.User.CreateTelegramLinkToken))
	mux.Handle("POST /v1/boards", protected(handlers.Boards.Create))
	mux.Handle("GET /v1/boards/{boardId}", protected(handlers.Boards.Get))
	mux.Handle("GET /v1/boards/{boardId}/aggregate", protected(handlers.Boards.GetAggregate))
	mux.Handle("PATCH /v1/boards/{boardId}", protected(handlers.Boards.Update))
	mux.Handle("DELETE /v1/boards/{boardId}", protected(handlers.Boards.Delete))
	mux.Handle("GET /v1/boards", protected(handlers.Boards.ListByOwnerID))
	mux.Handle("POST /v1/boards/{boardId}/columns", protected(handlers.Columns.Create))
	mux.Handle("GET /v1/boards/{boardId}/columns", protected(handlers.Columns.ListByBoardID))
	mux.Handle("PATCH /v1/boards/{boardId}/columns/{columnId}", protected(handlers.Columns.Update))
	mux.Handle("PUT /v1/boards/{boardId}/columns/{columnId}/position", protected(handlers.Columns.Move))
	mux.Handle("DELETE /v1/boards/{boardId}/columns/{columnId}", protected(handlers.Columns.Delete))
	mux.Handle("POST /v1/boards/{boardId}/columns/{columnId}/tasks", protected(handlers.Tasks.Create))
	mux.Handle("GET /v1/boards/{boardId}/columns/{columnId}/tasks", protected(handlers.Tasks.ListByColumnID))
	mux.Handle("PATCH /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}", protected(handlers.Tasks.Update))
	mux.Handle("PUT /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}/position", protected(handlers.Tasks.Move))
	mux.Handle("DELETE /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}", protected(handlers.Tasks.Delete))
	mux.Handle("POST /webhook/telegram", public(handlers.Telegram.Webhook))
	mux.Handle("GET "+swaggerBasePath, NewSwaggerHandler(swaggerBasePath, loginPath))

	return middlewares.Timeout.Wrap(middlewares.RequestID.Wrap(middlewares.CORS.Wrap(mux)))
}

func NewAdminRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}
