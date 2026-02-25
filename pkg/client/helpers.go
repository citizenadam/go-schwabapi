package client

import (
	"strconv"
	"time"
)

// FormatISO8601Date formats a time.Time to ISO8601 format (RFC3339).
// Example output: "2024-03-15T12:34:56Z"
func FormatISO8601Date(t time.Time) string {
	return t.Format(time.RFC3339)
}

// FormatEPOCH formats a time.Time to Unix timestamp in seconds.
// Example output: "1710502496"
func FormatEPOCH(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}

// FormatEPOCH_MS formats a time.Time to Unix timestamp in milliseconds.
// Example output: "1710502496000"
func FormatEPOCH_MS(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}

// FormatYYYYMMDD formats a time.Time to YYYY-MM-DD format.
// Example output: "2024-03-15"
func FormatYYYYMMDD(t time.Time) string {
	return t.Format("2006-01-02")
}
