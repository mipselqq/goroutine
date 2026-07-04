package repository

import "errors"

var (
	ErrRowNotFound      = errors.New("row not found")
	ErrUniqueViolation  = errors.New("attempt to insert unique value twice")
	ErrIndexOutOfBounds = errors.New("index out of bounds")

	ErrKeyExists   = errors.New("key already exists")
	ErrKeyNotFound = errors.New("key not found")
	ErrInternal    = errors.New("internal error")
)
