package service

import "time"

const rfc3339MillisLayout = "2006-01-02T15:04:05.000Z07:00"

func FormatRFC3339Millis(t time.Time) string {
	return t.UTC().Format(rfc3339MillisLayout)
}

func TimeNowRFC3339Millis() string {
	return FormatRFC3339Millis(timeNow())
}

func timeNow() time.Time {
	return time.Now().UTC()
}
