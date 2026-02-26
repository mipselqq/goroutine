package middleware

import "net/http"

type Middleware interface {
	Wrap(next http.Handler) http.HandlerFunc
}

type Middlewares struct {
	Metrics Middleware
	CORS    Middleware
	Auth    Middleware
}

func NewMiddlewares(metrics *Metrics, cors *CORS, auth *Auth) *Middlewares {
	return &Middlewares{
		Metrics: metrics,
		CORS:    cors,
		Auth:    auth,
	}
}
