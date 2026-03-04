package service

import "time"

// TODO: make actually nano
func TimeRFC3339Nano() string {
	return time.Now().UTC().Format(time.RFC3339)
}
