package domain

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

type ID[T any] struct {
	value uuid.UUID
}

func NewID[T any]() ID[T] {
	id, _ := uuid.NewV7()
	return ID[T]{value: id}
}

func ParseID[T any](s string) (ID[T], error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return ID[T]{}, fmt.Errorf("parse id: %w", err)
	}
	return ID[T]{value: u}, nil
}

func (id ID[T]) String() string {
	return id.value.String()
}

func (id ID[T]) IsEmpty() bool {
	return id.value == uuid.Nil
}

func (id ID[T]) UUID() uuid.UUID {
	return id.value
}

func (id *ID[T]) Scan(src any) error {
	if src == nil {
		id.value = uuid.Nil
		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		id.value = parsed
	case []byte:
		const uuidByteLen = 16
		isUUIDBytes := len(v) == uuidByteLen

		if isUUIDBytes {
			parsed, err := uuid.FromBytes(v)
			if err != nil {
				return err
			}
			id.value = parsed
		} else { // Byte representation of string
			parsed, err := uuid.ParseBytes(v)
			if err != nil {
				return err
			}
			id.value = parsed
		}
	default:
		return fmt.Errorf("unexpected type for ID: %T", src)
	}
	return nil
}

func (id ID[T]) Value() (driver.Value, error) {
	if id.IsEmpty() {
		return nil, nil
	}
	return id.value.String(), nil
}
