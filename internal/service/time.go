package service

import "time"

const RFC3339MillisLayout = "2006-01-02T15:04:05.000Z07:00"

type TimeStrFunc func() string

func FormatRFC3339Millis(t time.Time) string {
	return t.Format(RFC3339MillisLayout)
}

func TimeNowRFC3339Milli() string {
	return FormatRFC3339Millis(TimeNow())
}

type TimeFunc func() time.Time

func TimeNow() time.Time {
	return time.Now().UTC()
}
