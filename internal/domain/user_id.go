package domain

import (
	"database/sql/driver"
	"fmt"

	"github.com/google/uuid"
)

type UserID struct {
	value uuid.UUID
}

func NewUserID() UserID {
	id, _ := uuid.NewV7()
	return UserID{value: id}
}

func ParseUserID(s string) (UserID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return UserID{}, fmt.Errorf("parse user id: %w", err)
	}
	return UserID{value: u}, nil
}

func (u UserID) String() string {
	return u.value.String()
}

func (u UserID) IsEmpty() bool {
	return u.value == uuid.Nil
}

func (u UserID) UUID() uuid.UUID {
	return u.value
}

// Scan implements the sql.Scanner interface.
func (u *UserID) Scan(src any) error {
	if src == nil {
		u.value = uuid.Nil
		return nil
	}

	switch v := src.(type) {
	case string:
		parsed, err := uuid.Parse(v)
		if err != nil {
			return err
		}
		u.value = parsed
	case []byte:
		const uuidByteLen = 16
		isUUIDBytes := len(v) == uuidByteLen

		if isUUIDBytes {
			parsed, err := uuid.FromBytes(v)
			if err != nil {
				return err
			}
			u.value = parsed
		} else { // Byte representation of string
			parsed, err := uuid.ParseBytes(v)
			if err != nil {
				return err
			}
			u.value = parsed
		}
	default:
		return fmt.Errorf("unexpected type for UserID: %T", src)
	}
	return nil
}

// Value implements the driver.Valuer interface.
func (u UserID) Value() (driver.Value, error) {
	if u.IsEmpty() {
		return nil, nil
	}
	return u.value.String(), nil
}
