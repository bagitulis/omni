package services

import (
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
)

// Phase 11.5 — RED/GREEN tests for the scanner + formatter.
// The DB-driven cron scheduler is tested at the integration layer.

func mkConn(platform, store string, refreshInMs int64, disabled bool) models.CredentialConnection {
	c := models.CredentialConnection{
		Platform:        platform,
		StoreIdentifier: store,
		RefreshExpiry:   time.Now().UnixMilli() + refreshInMs,
	}
	if disabled {
		now := time.Now()
		c.DisabledAt = &now
	}
	return c
}

func TestScanConnectionsForExpiryWarnings_HappyPath(t *testing.T) {
	dayMs := int64(86_400_000)
	conns := []models.CredentialConnection{
		mkConn("shopee", "530635055", 3*dayMs, false), // in window → warn
		mkConn("tiktok", "7495139858797791519", 300*dayMs, false), // way out of window → skip
		mkConn("lazada", "shop-99", -5*dayMs, false),  // expired → warn
	}
	got := ScanConnectionsForExpiryWarnings("yumna_bertigamart", conns, 7)
	if len(got) != 2 {
		t.Fatalf("expected 2 warnings, got %d: %+v", len(got), got)
	}
	// Order matches input order.
	if got[0].Platform != "shopee" || got[0].AlreadyExpired {
		t.Fatalf("shopee should be warned but not-yet-expired: %+v", got[0])
	}
	if got[1].Platform != "lazada" || !got[1].AlreadyExpired {
		t.Fatalf("lazada should be flagged as already expired: %+v", got[1])
	}
}

func TestScanConnectionsForExpiryWarnings_SkipsDisabled(t *testing.T) {
	conns := []models.CredentialConnection{
		mkConn("shopee", "s1", 1*86_400_000, true), // expiring but disabled → skip
	}
	got := ScanConnectionsForExpiryWarnings("t1", conns, 7)
	if len(got) != 0 {
		t.Fatalf("expected 0 warnings for disabled conn, got %d", len(got))
	}
}

func TestScanConnectionsForExpiryWarnings_ZeroWindowUsesDefault(t *testing.T) {
	conns := []models.CredentialConnection{
		mkConn("shopee", "s1", 5*86_400_000, false),
	}
	// 0 window → falls back to DefaultRefreshExpiryWindowDays (7)
	got := ScanConnectionsForExpiryWarnings("t1", conns, 0)
	if len(got) != 1 {
		t.Fatalf("expected 1 warning with default window, got %d", len(got))
	}
}

func TestScanConnectionsForExpiryWarnings_EmptyInput(t *testing.T) {
	got := ScanConnectionsForExpiryWarnings("t1", nil, 7)
	if got == nil {
		t.Fatalf("expected empty slice, got nil")
	}
	if len(got) != 0 {
		t.Fatalf("expected 0 warnings for nil input, got %d", len(got))
	}
}

func TestFormatExpiryWarningMessage_Cases(t *testing.T) {
	cases := []struct {
		w           RefreshExpiryWarning
		wantSubstr  string
		description string
	}{
		{RefreshExpiryWarning{Platform: "shopee", AlreadyExpired: true}, "expired", "expired plain"},
		{RefreshExpiryWarning{Platform: "tiktok", DaysLeft: 1}, "TOMORROW", "one day left plural TOMORROW form"},
		{RefreshExpiryWarning{Platform: "lazada", DaysLeft: 5}, "expires in 5 days", "multi-day plural form"},
	}
	for _, tc := range cases {
		tc := tc
		t.Run(tc.description, func(t *testing.T) {
			msg := FormatExpiryWarningMessage(tc.w)
			if msg == "" {
				t.Fatalf("empty message for %+v", tc.w)
			}
			if !contains(msg, tc.wantSubstr) {
				t.Fatalf("message %q missing substr %q", msg, tc.wantSubstr)
			}
		})
	}
}

func TestFormatExpiryWarningTitle_Cases(t *testing.T) {
	if got := FormatExpiryWarningTitle(RefreshExpiryWarning{Platform: "shopee", AlreadyExpired: true}); !contains(got, "expired") {
		t.Fatalf("expected 'expired' in title, got %q", got)
	}
	if got := FormatExpiryWarningTitle(RefreshExpiryWarning{Platform: "tiktok", DaysLeft: 3}); !contains(got, "3d") {
		t.Fatalf("expected '3d' in title, got %q", got)
	}
}

func contains(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
