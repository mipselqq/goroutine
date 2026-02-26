package domain_test

import (
	"testing"

	"goroutine/internal/domain"

	"github.com/google/uuid"
)

func TestNewUserID(t *testing.T) {
	t.Parallel()

	id := domain.NewUserID()

	if id.IsEmpty() {
		t.Error("NewUserID() should not be empty")
	}

	if id.UUID().Version() != 7 {
		t.Errorf("Expected UUID v7, got v%d", id.UUID().Version())
	}
}

func TestParseUserID(t *testing.T) {
	t.Parallel()

	u := uuid.New()
	s := u.String()

	id, err := domain.ParseUserID(s)
	if err != nil {
		t.Errorf("ParseUserID() error = %v", err)
	}

	if id.String() != s {
		t.Errorf("Expected %s, got %s", s, id.String())
	}

	_, err = domain.ParseUserID("invalid")
	if err == nil {
		t.Error("ParseUserID() with invalid string should return error")
	}
}

func TestUserID_Scan(t *testing.T) {
	t.Parallel()

	u := uuid.New()
	var id domain.UserID

	t.Run("string", func(t *testing.T) {
		err := id.Scan(u.String())
		if err != nil {
			t.Errorf("Scan string error = %v", err)
		}
		if id.String() != u.String() {
			t.Errorf("Expected %s, got %s", u.String(), id.String())
		}
	})

	t.Run("bytes", func(t *testing.T) {
		err := id.Scan(u[:])
		if err != nil {
			t.Errorf("Scan bytes error = %v", err)
		}
		if id.String() != u.String() {
			t.Errorf("Expected %s, got %s", u.String(), id.String())
		}
	})

	t.Run("nil", func(t *testing.T) {
		err := id.Scan(nil)
		if err != nil {
			t.Errorf("Scan nil error = %v", err)
		}
		if !id.IsEmpty() {
			t.Error("Expected empty UserID after scanning nil")
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		err := id.Scan(123)

		if err == nil {
			t.Error("Scan with invalid type should return error")
		}
	})
}

func TestUserID_Value(t *testing.T) {
	t.Parallel()

	u, _ := uuid.NewV7()
	id, _ := domain.ParseUserID(u.String())

	val, err := id.Value()
	if err != nil {
		t.Errorf("Value() error = %v", err)
	}
	if val != u.String() {
		t.Errorf("Expected %v, got %v", u.String(), val)
	}

	var empty domain.UserID
	val, err = empty.Value()
	if err != nil {
		t.Errorf("Empty Value() error = %v", err)
	}
	if val != nil {
		t.Errorf("Expected nil value for empty UserID, got %v", val)
	}
}
