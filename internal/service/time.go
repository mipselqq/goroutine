package service

import "time"

func TimeRFC3339() string {
	return time.Now().Format(time.RFC3339)
}
