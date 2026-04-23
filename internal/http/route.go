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

	mux.Handle("POST /v1/register", m.Metrics.Wrap(http.HandlerFunc(h.Auth.Register)))
	mux.Handle("POST "+loginPath, m.Metrics.Wrap(http.HandlerFunc(h.Auth.Login)))
	mux.Handle("GET /v1/health", m.Metrics.Wrap(http.HandlerFunc(h.Health.Health)))
	mux.Handle("GET /v1/whoami", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Auth.WhoAmI))))
	mux.Handle("POST /v1/boards", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Boards.Create))))
	mux.Handle("GET /v1/boards/{boardId}", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Boards.Get))))
	mux.Handle("GET /v1/boards/{boardId}/aggregate", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Boards.GetAggregate))))
	mux.Handle("PATCH /v1/boards/{boardId}", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Boards.UpdateByID))))
	mux.Handle("DELETE /v1/boards/{boardId}", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Boards.Delete))))
	mux.Handle("GET /v1/boards", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Boards.GetMany))))
	mux.Handle("POST /v1/boards/{boardId}/columns", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Columns.Create))))
	mux.Handle("GET /v1/boards/{boardId}/columns", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Columns.List))))
	mux.Handle("PATCH /v1/boards/{boardId}/columns/{columnId}", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Columns.UpdateByID))))
	mux.Handle("PUT /v1/boards/{boardId}/columns/{columnId}/position", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Columns.Move))))
	mux.Handle("DELETE /v1/boards/{boardId}/columns/{columnId}", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Columns.Delete))))
	mux.Handle("POST /v1/boards/{boardId}/columns/{columnId}/tasks", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Tasks.Create))))
	mux.Handle("GET /v1/boards/{boardId}/columns/{columnId}/tasks", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Tasks.List))))
	mux.Handle("PATCH /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Tasks.UpdateByID))))
	mux.Handle("PUT /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}/position", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Tasks.Move))))
	mux.Handle("DELETE /v1/boards/{boardId}/columns/{columnId}/tasks/{taskId}", m.Metrics.Wrap(m.Auth.Wrap(http.HandlerFunc(h.Tasks.Delete))))
	mux.Handle("GET "+swaggerBasePath, NewSwaggerHandler(swaggerBasePath, loginPath))

	return m.RequestID.Wrap(m.CORS.Wrap(mux))
}

func NewAdminRouter() *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", promhttp.Handler())

	return mux
}
