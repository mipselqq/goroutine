package testutil

import "goroutine/internal/domain"

// TODO: remove this code function or create local domain/helpers_test.go
// and use it there to avoid uncontrolled global testutil growth
func ParseUserID(s string) domain.UserID {
	u, err := domain.ParseUserID(s)
	if err != nil {
		panic(err)
	}
	return u
}
