package secrecy

import (
	"fmt"
	"log/slog"
)

type SecretString string

func (s SecretString) RevealSecret() string {
	return string(s)
}

func (s SecretString) String() string {
	if len(s) == 0 {
		return ""
	}
	return fmt.Sprintf("(%d chars)", len(s))
}

func (s SecretString) LogValue() slog.Value {
	return slog.StringValue(s.String())
}
