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

func NewRouter(h *handler.Handlers, m *middleware.Middlewares) http.Handler {
	mux := http.NewServeMux()
	public := func(h http.HandlerFunc) http.Handler {
		return m.Metrics.Wrap(h)
	}
	protected := func(h http.HandlerFunc) http.Handler {
		return m.Metrics.Wrap(m.Auth.Wrap(h))
	}

	mux.Handle("POST /v1/register", public(h.Auth.Register))
	mux.Handle("POST "+loginPath, public(h.Auth.Login))
	mux.Handle("GET /v1/health", public(h.Health.Health))
	mux.Handle("GET /v1/whoami", protected(h.Auth.WhoAmI))
	mux.Handle("POST /v1/boards", protected(h.Boards.Create))
	mux.Handle("GET /v1/boards/{boardId}", protected(h.Boards.Get))
	mux.Handle("GET /v1/boards/{boardId}/aggregate", protected(h.Boards.GetAggregate))
	mux.Handle("PATCH /v1/boards/{boardId}", protected(h.Boards.UpdateByID))
	mux.Handle("DELETE /v1/boards/{boardId}", protected(h.Boards.Delete))
	mux.Handle("GET /v1/boards", protected(h.Boards.GetMany))
	mux.Handle("POST /v1/boards/{boardId}/columns", protected(h.Columns.Create))
	mux.Handle("GET /v1/boards/{boardId}/columns", protected(h.Columns.List))
	mux.Handle("PATCH /v1/boards/{boardId}/columns/{columnId}", protected(h.Columns.UpdateByID))
	mux.Handle("PUT /v1/boards/{boardId}/columns/{columnId}/position", protected(h.Columns.Move))
	mux.Handle("DELETE /v1/boards/{boardId}/columns/{columnId}", protected(h.Columns.Delete))
	mux.Handle("POST /v1/boards/{boardId}/columns/{columnId}/tasks", protected(h.Tasks.Create))
	mux.Handle("GET /v1/boards/{boardId}/columns/{columnId}/tasks", protected(h.Tasks.List))
	mux.Handle("PATCH /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}", protected(h.Tasks.UpdateByID))
	mux.Handle("PUT /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}/position", protected(h.Tasks.Move))
	mux.Handle("DELETE /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}", protected(h.Tasks.Delete))
	mux.Handle("GET "+swaggerBasePath, NewSwaggerHandler(swaggerBasePath, loginPath))

	return m.RequestID.Wrap(m.CORS.Wrap(mux))
}

func NewAdminRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}
