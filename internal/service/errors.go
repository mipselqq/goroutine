package service

import "errors"

var (
	ErrInternal           = errors.New("internal error happened")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserNotFound       = errors.New("user not found")
)
