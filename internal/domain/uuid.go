package domain

import (
	"database/sql/driver"
	"fmt"
	"reflect"

	"github.com/google/uuid"
)

type UUID[Tag any] struct {
	value uuid.UUID
}

func NewID[Tag any]() UUID[Tag] {
	id, _ := uuid.NewV7()
	return UUID[Tag]{value: id}
}

func ParseID[Tag any, Struct UUID[Tag]](s string) (Struct, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return Struct{}, fmt.Errorf("parse id %s: %w", reflect.TypeFor[Tag](), err)
	}
	if u == uuid.Nil {
		return Struct{}, fmt.Errorf("parse id %s: nil UUID", reflect.TypeFor[Tag]())
	}
	return Struct{value: u}, nil
}

func NewIDFromUUID[Tag any, Struct UUID[Tag]](u uuid.UUID) (Struct, error) {
	if u == uuid.Nil {
		return Struct{}, fmt.Errorf("nil UUID")
	}
	return Struct{value: u}, nil
}

func (id UUID[Tag]) String() string {
	return id.value.String()
}

func (id UUID[Tag]) IsNil() bool {
	return id.value == uuid.Nil
}

func (id UUID[Tag]) UUID() uuid.UUID {
	return id.value
}

func (id UUID[Tag]) Value() (driver.Value, error) {
	return id.value.Value()
}
