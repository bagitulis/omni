// Phase 10.3 — CredentialBackfillTrigger implementation.
//
// After a successful OAuth callback, the credential service fires
// TriggerBackfill(tenantID, platform, storeIdentifier). This adapter
// translates that call into a bounded H-30 SyncPlatformOrders sweep across
// the "unpaid" + "unprocess" + "processed" categories that fresh sellers
// most need to see immediately.

package sync

import (
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

// BackfillWindowDays is the default lookback for a post-reauth backfill.
// 30 days matches the operational SLA: any order beyond 30 days is out of
// the shipping/refund window and can be picked up by the reguler cron.
const BackfillWindowDays = 30

// backfillCategories lists the categories worth force-syncing right after a
// re-authorization. Kept small on purpose: we want a fast operator feedback
// loop, not a full-history rebuild (which is what the wallet-tx / escrow
// syncer is for).
var backfillCategories = []OrderStatusCategory{
	StatusUnpaid,
	StatusUnprocess,
	StatusProcessed,
}

// CredentialBackfillAdapter implements the services.CredentialBackfillTrigger
// contract without introducing a services↔sync import cycle. Wired at app
// startup via CredentialApiService.SetBackfillTrigger.
type CredentialBackfillAdapter struct{}

// NewCredentialBackfillAdapter returns a ready-to-wire trigger.
func NewCredentialBackfillAdapter() *CredentialBackfillAdapter {
	return &CredentialBackfillAdapter{}
}

// TriggerBackfill runs a bounded H-30 sync for the given tenant/platform.
// Best-effort: partial failures are logged and downgraded to a non-fatal
// error so the caller can decide whether to alert.
func (a *CredentialBackfillAdapter) TriggerBackfill(ctx context.Context, tenantID, platform, storeIdentifier string) error {
	if tenantID == "" || platform == "" || storeIdentifier == "" {
		return fmt.Errorf("TriggerBackfill: missing identifier (tenant=%q platform=%q store=%q)", tenantID, platform, storeIdentifier)
	}
	svc, err := GetOrderSyncService(tenantID)
	if err != nil {
		return fmt.Errorf("resolve order sync service: %w", err)
	}
	if !svc.IsInitialized() {
		return fmt.Errorf("order sync service not initialized for tenant %s", tenantID)
	}
	pt := PlatformType(platform)
	var firstErr error
	for _, cat := range backfillCategories {
		if _, syncErr := svc.SyncPlatformOrders(ctx, pt, cat, BackfillWindowDays); syncErr != nil {
			log.Warn().Err(syncErr).Str("tenant_id", tenantID).Str("platform", platform).
				Str("category", string(cat)).Msg("post-reauth backfill: category sync failed (continuing)")
			if firstErr == nil {
				firstErr = syncErr
			}
		}
	}
	if firstErr != nil {
		return fmt.Errorf("post-reauth backfill completed with errors: %w", firstErr)
	}
	log.Info().Str("tenant_id", tenantID).Str("platform", platform).Str("store", storeIdentifier).
		Int("window_days", BackfillWindowDays).Msg("Post-reauth backfill completed")
	return nil
}
