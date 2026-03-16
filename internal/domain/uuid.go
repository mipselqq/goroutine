package domain

import (
	"database/sql/driver"
	"fmt"
	"reflect"

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
		return UUID[T]{}, fmt.Errorf("parse id %s: %w", reflect.TypeFor[T](), err)
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

	if err := id.value.Scan(src); err != nil {
		return fmt.Errorf("id %s: %w: %v", reflect.TypeFor[T](), ErrDataCorrupted, err)
	}

	return nil
}

func (id UUID[T]) Value() (driver.Value, error) {
	return id.value.Value()
}
