package service

import "errors"

var (
	ErrInternal             = errors.New("internal error happened")
	ErrBoardNotFound        = errors.New("board not found")
	ErrColumnNotFound       = errors.New("column not found")
	ErrTaskNotFound         = errors.New("task not found")
	ErrIndexOutOfBounds     = errors.New("index out of bounds")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrInvalidCredentials   = errors.New("invalid email or password")
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidToken         = errors.New("invalid token")
	ErrTokenExpired         = errors.New("token expired")
	ErrInvalidSigningMethod = errors.New("invalid signing method")
)
