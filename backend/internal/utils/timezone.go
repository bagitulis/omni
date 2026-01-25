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

// NowUTC returns current time in UTC
func NowUTC() time.Time {
	return time.Now().UTC()
}

// ToWIB converts any time to WIB timezone
func ToWIB(t time.Time) time.Time {
	return t.In(WIB)
}

// ToUTC converts any time to UTC
func ToUTC(t time.Time) time.Time {
	return t.UTC()
}

// FormatISO8601 formats time as ISO8601 string in UTC
// This matches JavaScript's toISOString() output
func FormatISO8601(t time.Time) string {
	return t.UTC().Format(ISO8601Format)
}

// FormatISO8601WIB formats time as ISO8601 string with WIB offset
func FormatISO8601WIB(t time.Time) string {
	return t.In(WIB).Format("2006-01-02T15:04:05+07:00")
}

// FormatDateOnly formats time as YYYY-MM-DD in WIB timezone
func FormatDateOnly(t time.Time) string {
	return t.In(WIB).Format(DateOnlyFormat)
}

// FormatTimeOnly formats time as HH:MM in WIB timezone
func FormatTimeOnly(t time.Time) string {
	return t.In(WIB).Format(TimeOnlyFormat)
}

// FormatDateTime formats time as YYYY-MM-DD HH:MM:SS in WIB timezone
func FormatDateTime(t time.Time) string {
	return t.In(WIB).Format(DateTimeFormat)
}

// GetWIBHour returns current hour in WIB timezone (0-23)
func GetWIBHour() int {
	return NowWIB().Hour()
}

// GetWIBTimeString returns current time as HH:MM in WIB timezone
func GetWIBTimeString() string {
	return NowWIB().Format(TimeOnlyFormat)
}

// IsWithinTimeWindow checks if current time is within start-end window
// Start and end should be in "HH:MM" format (24-hour)
// Uses WIB timezone for comparison
func IsWithinTimeWindow(start, end string) bool {
	current := GetWIBTimeString()
	return current >= start && current <= end
}

// ParseTimeWindow parses HH:MM format and returns hour and minute
func ParseTimeWindow(timeStr string) (hour, minute int, err error) {
	t, err := time.Parse(TimeOnlyFormat, timeStr)
	if err != nil {
		return 0, 0, err
	}
	return t.Hour(), t.Minute(), nil
}

// StartOfDayWIB returns the start of the current day in WIB
func StartOfDayWIB() time.Time {
	now := NowWIB()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, WIB)
}

// EndOfDayWIB returns the end of the current day in WIB
func EndOfDayWIB() time.Time {
	now := NowWIB()
	return time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, WIB)
}

// UnixMilliToTime converts Unix milliseconds to time.Time
func UnixMilliToTime(ms int64) time.Time {
	return time.UnixMilli(ms)
}

// TimeToUnixMilli converts time.Time to Unix milliseconds
func TimeToUnixMilli(t time.Time) int64 {
	return t.UnixMilli()
}

// DaysAgoWIB returns time N days ago at start of day in WIB
func DaysAgoWIB(days int) time.Time {
	return StartOfDayWIB().AddDate(0, 0, -days)
}

// FormatForAPI formats time for JSON API response (UTC ISO8601)
// This is the standard format for all API responses
func FormatForAPI(t time.Time) string {
	return FormatISO8601(t)
}

// FormatForDisplay formats time for user display (WIB DateTime)
func FormatForDisplay(t time.Time) string {
	return FormatDateTime(t)
}
