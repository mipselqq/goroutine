package middleware

import "net/http"

type middleware interface {
	Wrap(next http.Handler) http.Handler
}

type Middlewares struct {
	Metrics   middleware
	CORS      middleware
	Auth      middleware
	RequestID middleware
	Timeout   middleware
}
