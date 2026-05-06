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

func TestToWIB(t *testing.T) {
	utcTime := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	wibTime := ToWIB(utcTime)

	// UTC 10:00 should be WIB 17:00
	if wibTime.Hour() != 17 {
		t.Errorf("ToWIB hour = %d, want 17", wibTime.Hour())
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
