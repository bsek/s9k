package utils

import (
	"time"
)

// Converts a Time instance to the local time zone
func toLocalTime(when time.Time) time.Time {
	localLoc, err := time.LoadLocation("Local")
	if err != nil {
		return when
	} else {
		return when.In(localLoc)
	}
}

// FormatLocalDateTime formats a datetime with the local time zone
func FormatLocalDateTime(when time.Time) string {
	return toLocalTime(when).Format("02-01-06 15:04:05")
}

// FormatLocalDate formats a date with the local time zone
func FormatLocalDate(when time.Time) string {
	return toLocalTime(when).Format("02-01-06")
}

// FormatLocalTime formats a time with the local time zone
func FormatLocalTime(when time.Time) string {
	return toLocalTime(when).Format("15:04:05")
}
