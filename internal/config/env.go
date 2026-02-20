package config

import (
	"os"
)

func getenvOrDefault(key, def string) string {
	env := os.Getenv(key)

	if env == "" {
		return def
	}
	return env
}
