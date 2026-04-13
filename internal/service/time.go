package service

import "time"

type TimeStrFunc func() string

func TimeRFC3339Milli() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
}

type TimeFunc func() time.Time

func TimeNow() time.Time {
	return time.Now().UTC()
}
