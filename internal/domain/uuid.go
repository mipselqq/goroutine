package domain

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

type UUID[T any] struct {
	value uuid.UUID
}

func NewID[T any]() UUID[T] {
	id, _ := uuid.NewV7()
	return UUID[T]{value: id}
}

func ParseID[T any](s string) (UUID[T], error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return UUID[T]{}, fmt.Errorf("parse id: %w", err)
	}
	return UUID[T]{value: u}, nil
}

func (id UUID[T]) String() string {
	return id.value.String()
}

func (id UUID[T]) IsEmpty() bool {
	return id.value == uuid.Nil
}

func (id UUID[T]) UUID() uuid.UUID {
	return id.value
}

func (id *UUID[T]) Scan(src any) error {
	if src == nil {
		id.value = uuid.Nil
		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return fmt.Errorf("id: %w: %v", ErrDataCorrupted, err)
		}
		id.value = parsed
	case []byte:
		const uuidByteLen = 16
		isUUIDBytes := len(v) == uuidByteLen

		if isUUIDBytes {
			parsed, err := uuid.FromBytes(v)
			if err != nil {
				return fmt.Errorf("id bytes: %w: %v", ErrDataCorrupted, err)
			}
			id.value = parsed
		} else { // Byte representation of string
			parsed, err := uuid.ParseBytes(v)
			if err != nil {
				return fmt.Errorf("id bytes: %w: %v", ErrDataCorrupted, err)
			}
			id.value = parsed
		}
	default:
		return fmt.Errorf("unexpected type for ID: %T", src)
	}
	return nil
}

func (id UUID[T]) Value() (driver.Value, error) {
	if id.IsEmpty() {
		return nil, nil
	}
	return id.value.String(), nil
}
