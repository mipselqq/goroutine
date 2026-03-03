package service

import "time"

func TimeRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
