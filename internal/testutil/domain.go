package testutil

import "goroutine/internal/domain"

func ParseUserID(s string) domain.UserID {
	u, err := domain.ParseUserID(s)
	if err != nil {
		panic(err)
	}
	return u
}
