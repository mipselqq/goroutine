package repository

import "errors"

var (
	ErrRowNotFound     = errors.New("row not found")
	ErrUniqueViolation = errors.New("attempt to insert unique value twice")
	ErrInternal   = errors.New("internal error")
)
