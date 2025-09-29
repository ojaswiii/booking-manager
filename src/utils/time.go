package utils

import (
	"time"
)

// ParseTime parses a time string in ISO 8601 format
func ParseTime(timeStr string) (time.Time, error) {
	// Try different time formats
	formats := []string{
		time.RFC3339,           // 2006-01-02T15:04:05Z07:00
		time.RFC3339Nano,       // 2006-01-02T15:04:05.999999999Z07:00
		"2006-01-02T15:04:05Z", // 2006-01-02T15:04:05Z
		"2006-01-02 15:04:05",  // 2006-01-02 15:04:05
		"2006-01-02",           // 2006-01-02
	}

	for _, format := range formats {
		if t, err := time.Parse(format, timeStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, &time.ParseError{
		Layout:     "ISO 8601",
		Value:      timeStr,
		LayoutElem: "2006-01-02T15:04:05Z",
		ValueElem:  timeStr,
		Message:    "unable to parse time string",
	}
}

// FormatTime formats a time to ISO 8601 string
func FormatTime(t time.Time) string {
	return t.Format(time.RFC3339)
}

// Now returns current time in UTC
func Now() time.Time {
	return time.Now().UTC()
}

// AddMinutes adds minutes to a time
func AddMinutes(t time.Time, minutes int) time.Time {
	return t.Add(time.Duration(minutes) * time.Minute)
}

// IsExpired checks if a time has passed
func IsExpired(t time.Time) bool {
	return time.Now().After(t)
}
