package utils

import (
	"testing"
	"time"
)

func TestWIBInitialization(t *testing.T) {
	if WIB == nil {
		t.Error("WIB timezone should be initialized")
	}

	// Verify WIB is UTC+7
	now := time.Now()
	_, offset := now.In(WIB).Zone()
	expectedOffset := 7 * 60 * 60 // 7 hours in seconds
	if offset != expectedOffset {
		t.Errorf("WIB offset = %d, want %d", offset, expectedOffset)
	}
}

func TestNowWIB(t *testing.T) {
	now := NowWIB()

	// Should be in WIB timezone
	if now.Location() != WIB {
		t.Error("NowWIB() should return time in WIB timezone")
	}
}

func TestNowUTC(t *testing.T) {
	now := NowUTC()

	// Should be in UTC timezone
	if now.Location() != time.UTC {
		t.Error("NowUTC() should return time in UTC timezone")
	}
}

func TestToWIB(t *testing.T) {
	utcTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	wibTime := ToWIB(utcTime)

	// UTC 10:00 should be WIB 17:00
	if wibTime.Hour() != 17 {
		t.Errorf("ToWIB hour = %d, want 17", wibTime.Hour())
	}
}

func TestToUTC(t *testing.T) {
	wibTime := time.Date(2024, 1, 15, 17, 0, 0, 0, WIB)
	utcTime := ToUTC(wibTime)

	// WIB 17:00 should be UTC 10:00
	if utcTime.Hour() != 10 {
		t.Errorf("ToUTC hour = %d, want 10", utcTime.Hour())
	}
}

func TestFormatISO8601(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC)
	formatted := FormatISO8601(testTime)

	expected := "2024-01-15T10:30:45.123Z"
	if formatted != expected {
		t.Errorf("FormatISO8601() = %v, want %v", formatted, expected)
	}
}

func TestFormatISO8601WIB(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 17, 30, 45, 0, WIB)
	formatted := FormatISO8601WIB(testTime)

	expected := "2024-01-15T17:30:45+07:00"
	if formatted != expected {
		t.Errorf("FormatISO8601WIB() = %v, want %v", formatted, expected)
	}
}

func TestFormatDateOnly(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 17, 30, 45, 0, WIB)
	formatted := FormatDateOnly(testTime)

	expected := "2024-01-15"
	if formatted != expected {
		t.Errorf("FormatDateOnly() = %v, want %v", formatted, expected)
	}
}

func TestFormatTimeOnly(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 17, 30, 45, 0, WIB)
	formatted := FormatTimeOnly(testTime)

	expected := "17:30"
	if formatted != expected {
		t.Errorf("FormatTimeOnly() = %v, want %v", formatted, expected)
	}
}

func TestFormatDateTime(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 17, 30, 45, 0, WIB)
	formatted := FormatDateTime(testTime)

	expected := "2024-01-15 17:30:45"
	if formatted != expected {
		t.Errorf("FormatDateTime() = %v, want %v", formatted, expected)
	}
}

func TestParseTimeWindow(t *testing.T) {
	tests := []struct {
		input   string
		hour    int
		minute  int
		wantErr bool
	}{
		{"09:00", 9, 0, false},
		{"23:59", 23, 59, false},
		{"00:00", 0, 0, false},
		{"invalid", 0, 0, true},
		{"25:00", 0, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			hour, minute, err := ParseTimeWindow(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if hour != tt.hour || minute != tt.minute {
				t.Errorf("ParseTimeWindow() = (%d, %d), want (%d, %d)", hour, minute, tt.hour, tt.minute)
			}
		})
	}
}

func TestStartOfDayWIB(t *testing.T) {
	start := StartOfDayWIB()

	if start.Hour() != 0 || start.Minute() != 0 || start.Second() != 0 {
		t.Errorf("StartOfDayWIB() time = %v, want 00:00:00", start.Format("15:04:05"))
	}
	if start.Location() != WIB {
		t.Error("StartOfDayWIB() should be in WIB timezone")
	}
}

func TestEndOfDayWIB(t *testing.T) {
	end := EndOfDayWIB()

	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Errorf("EndOfDayWIB() time = %v, want 23:59:59", end.Format("15:04:05"))
	}
	if end.Location() != WIB {
		t.Error("EndOfDayWIB() should be in WIB timezone")
	}
}

func TestUnixMilliConversion(t *testing.T) {
	ms := int64(1705320000000) // 2024-01-15 10:00:00 UTC
	converted := UnixMilliToTime(ms)

	if converted.Unix()*1000 != ms {
		t.Errorf("UnixMilliToTime conversion mismatch")
	}

	// Round trip
	backToMs := TimeToUnixMilli(converted)
	if backToMs != ms {
		t.Errorf("TimeToUnixMilli() = %d, want %d", backToMs, ms)
	}
}

func TestDaysAgoWIB(t *testing.T) {
	today := StartOfDayWIB()

	oneDayAgo := DaysAgoWIB(1)
	diff := today.Sub(oneDayAgo)
	if diff != 24*time.Hour {
		t.Errorf("DaysAgoWIB(1) diff = %v, want 24h", diff)
	}

	sevenDaysAgo := DaysAgoWIB(7)
	diff = today.Sub(sevenDaysAgo)
	if diff != 7*24*time.Hour {
		t.Errorf("DaysAgoWIB(7) diff = %v, want 168h", diff)
	}
}

func TestFormatForAPI(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 123000000, time.UTC)
	formatted := FormatForAPI(testTime)

	// Should be same as FormatISO8601
	expected := FormatISO8601(testTime)
	if formatted != expected {
		t.Errorf("FormatForAPI() = %v, want %v", formatted, expected)
	}
}

func TestFormatForDisplay(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 17, 30, 45, 0, WIB)
	formatted := FormatForDisplay(testTime)

	// Should be same as FormatDateTime
	expected := FormatDateTime(testTime)
	if formatted != expected {
		t.Errorf("FormatForDisplay() = %v, want %v", formatted, expected)
	}
}

func TestConstants(t *testing.T) {
	if ISO8601Format != "2006-01-02T15:04:05.000Z" {
		t.Errorf("ISO8601Format = %v, want '2006-01-02T15:04:05.000Z'", ISO8601Format)
	}
	if DateOnlyFormat != "2006-01-02" {
		t.Errorf("DateOnlyFormat = %v, want '2006-01-02'", DateOnlyFormat)
	}
	if TimeOnlyFormat != "15:04" {
		t.Errorf("TimeOnlyFormat = %v, want '15:04'", TimeOnlyFormat)
	}
	if WIBOffset != 7 {
		t.Errorf("WIBOffset = %v, want 7", WIBOffset)
	}
}
