package models

import (
	"testing"
	"time"
)

// Phase 11.5 — RED tests for refresh-expiry window helpers.
// Powers the cron notifier that warns sellers 7 days before their
// refresh_token dies (before it's too late to auto-recover).

func TestRefreshExpiryDaysLeft_FutureExpiry(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	c := &CredentialConnection{RefreshExpiry: nowMs + 5*86_400_000}
	got := c.RefreshExpiryDaysLeft()
	if got < 4 || got > 5 {
		t.Fatalf("expected 4-5 days left, got %d", got)
	}
}

func TestRefreshExpiryDaysLeft_ZeroReturnsZero(t *testing.T) {
	c := &CredentialConnection{RefreshExpiry: 0}
	if c.RefreshExpiryDaysLeft() != 0 {
		t.Fatalf("expected 0 for unset RefreshExpiry")
	}
}

func TestRefreshExpiryDaysLeft_PastReturnsZero(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	c := &CredentialConnection{RefreshExpiry: nowMs - 10*86_400_000}
	if got := c.RefreshExpiryDaysLeft(); got != 0 {
		t.Fatalf("expected 0 for past expiry, got %d", got)
	}
}

func TestIsRefreshExpiringWithin_InsideWindow(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	c := &CredentialConnection{RefreshExpiry: nowMs + 3*86_400_000}
	if !c.IsRefreshExpiringWithin(7) {
		t.Fatalf("3 days left should be inside 7-day window")
	}
}

func TestIsRefreshExpiringWithin_OutsideWindow(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	c := &CredentialConnection{RefreshExpiry: nowMs + 10*86_400_000}
	if c.IsRefreshExpiringWithin(7) {
		t.Fatalf("10 days left should be outside 7-day window")
	}
}

func TestIsRefreshExpiringWithin_AlreadyExpired(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	c := &CredentialConnection{RefreshExpiry: nowMs - 86_400_000}
	// Already expired = "in window" so notifier still fires final warning.
	if !c.IsRefreshExpiringWithin(7) {
		t.Fatalf("expired refresh token should count as expiring within window")
	}
}

func TestIsRefreshExpiringWithin_ZeroWindowIsNever(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	c := &CredentialConnection{RefreshExpiry: nowMs + 86_400_000}
	// Window 0 or negative = disabled; never emit warning.
	if c.IsRefreshExpiringWithin(0) {
		t.Fatalf("window 0 must never flag as expiring")
	}
	if c.IsRefreshExpiringWithin(-3) {
		t.Fatalf("negative window must never flag as expiring")
	}
}

func TestIsRefreshExpiringWithin_UnsetExpiryIsExpired(t *testing.T) {
	// Consistent with IsRefreshTokenExpired: zero = expired = needs warning.
	c := &CredentialConnection{RefreshExpiry: 0}
	if !c.IsRefreshExpiringWithin(7) {
		t.Fatalf("unset RefreshExpiry (0) should count as expiring")
	}
}
