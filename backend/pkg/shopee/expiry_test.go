package shopee

import (
	"testing"
	"time"
)

// Phase 8 — Shopee partner-key expiry helpers.
//
// The UI needs to render a "Rotate soon" warning banner before a partner_key
// hard-expires. Keeping the policy centralised (single function, single warning
// window definition) so backend, frontend, and any future ops script share the
// same math.

func TestIsPartnerKeyExpiring_ExpiringWithinWindow(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name    string
		expires time.Time
	}{
		{"1_day_left", now.Add(24 * time.Hour)},
		{"6_days_left", now.Add(6 * 24 * time.Hour)},
		{"7_days_exact", now.Add(7 * 24 * time.Hour)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !IsPartnerKeyExpiring(tc.expires, 7) {
				t.Errorf("IsPartnerKeyExpiring(%v, 7) = false, want true", tc.expires)
			}
		})
	}
}

func TestIsPartnerKeyExpiring_AlreadyExpired(t *testing.T) {
	// A past date must also report "expiring" so the UI banner surfaces the
	// urgent case rather than silently going green.
	past := time.Now().Add(-24 * time.Hour)
	if !IsPartnerKeyExpiring(past, 7) {
		t.Errorf("IsPartnerKeyExpiring(past, 7) = false, want true (past is worse than expiring)")
	}
}

func TestIsPartnerKeyExpiring_FreshKey(t *testing.T) {
	// Clearly not expiring — 30 days out with a 7-day window.
	fresh := time.Now().Add(30 * 24 * time.Hour)
	if IsPartnerKeyExpiring(fresh, 7) {
		t.Errorf("IsPartnerKeyExpiring(30d_future, 7) = true, want false")
	}
}

func TestIsPartnerKeyExpiring_ZeroTime_NotExpiring(t *testing.T) {
	// A zero-value time (no expiry recorded) MUST be treated as "not
	// expiring" — otherwise every unmigrated legacy row would spawn a false
	// warning banner. This is the negative-path guard.
	if IsPartnerKeyExpiring(time.Time{}, 7) {
		t.Errorf("IsPartnerKeyExpiring(zero_time, 7) = true, want false")
	}
}

func TestIsPartnerKeyExpiring_ZeroWindow_OnlyExpiredIsFlagged(t *testing.T) {
	// warningDays == 0 → only past dates count as expiring. Helps a caller
	// that just wants a plain "is-expired" check without a window.
	now := time.Now()
	if IsPartnerKeyExpiring(now.Add(24*time.Hour), 0) {
		t.Errorf("warningDays=0 with future date: want false, got true")
	}
	if !IsPartnerKeyExpiring(now.Add(-time.Second), 0) {
		t.Errorf("warningDays=0 with past date: want true, got false")
	}
}

func TestIsPartnerKeyExpiring_NegativeWindow_Defensive(t *testing.T) {
	// Caller passed a garbage negative value — behave like zero (only
	// past-dates flagged) rather than panic or silently invert the check.
	if IsPartnerKeyExpiring(time.Now().Add(24*time.Hour), -5) {
		t.Errorf("negative window with future date: want false, got true")
	}
}
