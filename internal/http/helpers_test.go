package http_test

import (
	"context"
	"net/http"

	"goroutine/internal/domain"
)

type stubBoardsService struct{}

func (stubBoardsService) Create(context.Context, domain.UserID, domain.BoardName, domain.BoardDescription) (domain.Board, error) {
	return domain.Board{}, nil
}

func (stubBoardsService) Get(context.Context, domain.UserID, domain.BoardID) (domain.Board, error) {
	return domain.Board{}, nil
}

func (stubBoardsService) GetMany(context.Context, domain.UserID) ([]domain.Board, error) {
	return nil, nil
}

type spyMetricsMiddleware struct{}

func (s *spyMetricsMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Metrics-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}

type spyCorsMiddleware struct{}

func (s *spyCorsMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Cors-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}

type spyAuthMiddleware struct{}

func (s *spyAuthMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Auth-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}

type spyRequestIDMiddleware struct{}

func (s *spyRequestIDMiddleware) Wrap(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-RequestId-Tracked", "true")
		next.ServeHTTP(w, r)
	}
}
