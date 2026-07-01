package repository

import "errors"

var (
	ErrRowNotFound                    = errors.New("row not found")
	ErrUniqueViolation                = errors.New("attempt to insert unique value twice")
	ErrIndexOutOfBounds               = errors.New("index out of bounds")
	ErrTelegramLinkTokenAlreadyExists = errors.New("telegram link token already exists")
	ErrTelegramLinkTokenNotFound      = errors.New("telegram link token not found")
	ErrInternal                       = errors.New("internal error")
)
