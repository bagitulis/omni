// Phase 10.3 — post-OAuth-success backfill trigger.
//
// After a successful re-authorization (Shopee/Lazada/TikTok OAuth callback),
// the freshly-connected shop typically has 30 days of unsynced orders that
// piled up while the token was dead. Waiting for the reguler cron sweep wastes
// operator time. This wire lets the callback handler kick off an immediate
// bounded backfill without pulling the sync package into the services package
// (import cycle safety + testability).

package services

import (
	"context"
	"errors"
	"fmt"
)

// CredentialBackfillTrigger is the seam the callback handler uses to fire an
// auto-backfill after a successful OAuth re-authorization. The concrete
// implementation (wired in app.go) delegates to the existing sync manager.
//
// Kept tiny on purpose: TriggerBackfill takes only the identifiers the
// callback already has and lets the implementation decide the window (default
// 30 days) and error-handling strategy. Best-effort: the callback never fails
// because of a backfill error.
type CredentialBackfillTrigger interface {
	TriggerBackfill(ctx context.Context, tenantID, platform, storeIdentifier string) error
}

// errBackfillNotWired is returned when the CredentialApiService has no
// backfill trigger injected. Not a hard failure: the callback handler treats
// it as "backfill skipped" and continues.
var errBackfillNotWired = errors.New("backfill trigger not wired")

// invokeBackfill is the safe dispatcher used by the callback handler. Returns
// nil if the trigger is not wired (opt-in behaviour) so an unconfigured
// deployment still succeeds at OAuth without silently masking real errors.
func invokeBackfill(ctx context.Context, t CredentialBackfillTrigger, tenantID, platform, storeIdentifier string) error {
	if t == nil {
		return errBackfillNotWired
	}
	if tenantID == "" || platform == "" || storeIdentifier == "" {
		return fmt.Errorf("invokeBackfill: missing identifier (tenant=%q platform=%q store=%q)", tenantID, platform, storeIdentifier)
	}
	return t.TriggerBackfill(ctx, tenantID, platform, storeIdentifier)
}
