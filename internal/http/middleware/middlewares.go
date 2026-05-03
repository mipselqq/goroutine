package middleware

import "net/http"

type Middleware interface {
	Wrap(next http.Handler) http.Handler
}

type Middlewares struct {
	Metrics   Middleware
	CORS      Middleware
	Auth      Middleware
	RequestID Middleware
}
