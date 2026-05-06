package utils

import (
	"time"
)

// WIB is the Indonesia Western Time (UTC+7) location
// Used consistently across the application for Jakarta timezone
var WIB *time.Location

// Common time format constants
const (
	// ISO8601Format is standard ISO 8601 format with milliseconds
	ISO8601Format = "2006-01-02T15:04:05.000Z"

	// DateOnlyFormat for date-only strings (YYYY-MM-DD)
	DateOnlyFormat = "2006-01-02"

	// TimeOnlyFormat for time comparisons (HH:MM)
	TimeOnlyFormat = "15:04"

	// DateTimeFormat for display (YYYY-MM-DD HH:MM:SS)
	DateTimeFormat = "2006-01-02 15:04:05"

	// WIBOffset is UTC+7 offset for Indonesia
	WIBOffset = 7
)

func init() {
	var err error
	WIB, err = time.LoadLocation("Asia/Jakarta")
	if err != nil {
		// Fallback to fixed offset if timezone data not available
		WIB = time.FixedZone("WIB", WIBOffset*60*60)
	}
}

// NowWIB returns current time in WIB (UTC+7) timezone
func NowWIB() time.Time {
	return time.Now().In(WIB)
}

// ToWIB converts any time to WIB timezone
func ToWIB(t time.Time) time.Time {
	return t.In(WIB)
}

