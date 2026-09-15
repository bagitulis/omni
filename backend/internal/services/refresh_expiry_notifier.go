// Phase 11.5 — refresh_token expiry cron notifier.
//
// Prevents a repeat of the 2026-09 Shopee Yumna incident: refresh_token
// silently expired 45 days past grace period → seller had to re-auth
// through Partner Center + OTP recovery. This notifier scans every
// tenant's credential_connections once a day and fans a warning
// notification when any connection's refresh_token will expire inside
// the configured window (default 7 days).

package services

import (
	"context"
	"fmt"

	"github.com/omni/backend/internal/models"
)

// DefaultRefreshExpiryWindowDays is the default warning horizon. 7 days
// matches the Phase 10.2 partner-key expiry banner (isPartnerKeyExpiring).
const DefaultRefreshExpiryWindowDays = 7

// RefreshExpiryWarning describes a single connection that needs
// re-authorization before its refresh_token dies.
type RefreshExpiryWarning struct {
	TenantID        string
	Platform        string
	StoreIdentifier string
	DaysLeft        int  // 0 = already expired
	AlreadyExpired  bool // convenience flag; equal to conn.IsRefreshTokenExpired()
}

// ScanConnectionsForExpiryWarnings inspects a slice of connections and
// returns those inside the warning window. Pure function — no DB, no
// clock injection needed because CredentialConnection methods handle
// their own time.Now() call.
//
// Skips disabled connections (DisabledAt != nil) because they're
// intentionally off and don't need a re-auth prompt.
func ScanConnectionsForExpiryWarnings(tenantID string, conns []models.CredentialConnection, windowDays int) []RefreshExpiryWarning {
	if windowDays <= 0 {
		windowDays = DefaultRefreshExpiryWindowDays
	}
	warnings := make([]RefreshExpiryWarning, 0)
	for i := range conns {
		c := &conns[i]
		if c.DisabledAt != nil {
			continue
		}
		if !c.IsRefreshExpiringWithin(windowDays) {
			continue
		}
		warnings = append(warnings, RefreshExpiryWarning{
			TenantID:        tenantID,
			Platform:        c.Platform,
			StoreIdentifier: c.StoreIdentifier,
			DaysLeft:        c.RefreshExpiryDaysLeft(),
			AlreadyExpired:  c.IsRefreshTokenExpired(),
		})
	}
	return warnings
}

// FormatExpiryWarningMessage produces the human-readable notification
// message. Kept as a helper so callers (cron + tests + future email
// digest) format identically.
func FormatExpiryWarningMessage(w RefreshExpiryWarning) string {
	if w.AlreadyExpired {
		return fmt.Sprintf("%s connection expired — click Re-authorize in Settings to reconnect.", w.Platform)
	}
	if w.DaysLeft == 1 {
		return fmt.Sprintf("%s connection expires TOMORROW — re-authorize now to avoid a sync outage.", w.Platform)
	}
	return fmt.Sprintf("%s connection expires in %d days — re-authorize soon to keep sync running.", w.Platform, w.DaysLeft)
}

// FormatExpiryWarningTitle produces the notification title. Short so it
// fits in the bell dropdown without truncation.
func FormatExpiryWarningTitle(w RefreshExpiryWarning) string {
	if w.AlreadyExpired {
		return fmt.Sprintf("%s connection expired", w.Platform)
	}
	return fmt.Sprintf("%s expires in %dd", w.Platform, w.DaysLeft)
}

// ExpiryWarningSink is what the cron uses to actually deliver the
// warning. Kept small on purpose so the concrete implementation (in
// cron/refresh_expiry_cron.go) can wire NotificationService.Push +
// realtime.Publisher.PublishNotification without dragging those into
// this file.
type ExpiryWarningSink interface {
	DeliverExpiryWarning(ctx context.Context, w RefreshExpiryWarning) error
}
