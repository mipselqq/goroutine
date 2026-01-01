package config

import "os"

func getenvOrDefault(key string, def string) string {
	env := os.Getenv(key)

	if env == "" {
		return def
	}
	return env
}
