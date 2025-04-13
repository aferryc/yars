package utils

import (
    "time"
)

// ParseDate parses a date string in the given layout and returns a time.Time object.
func ParseDate(dateStr string, layout string) (time.Time, error) {
    return time.Parse(layout, dateStr)
}

// FormatDate formats a time.Time object into a string using the given layout.
func FormatDate(t time.Time, layout string) string {
    return t.Format(layout)
}

// GetCurrentDate returns the current date as a time.Time object.
func GetCurrentDate() time.Time {
    return time.Now()
}