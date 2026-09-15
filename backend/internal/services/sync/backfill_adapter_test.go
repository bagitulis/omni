package sync

import (
	"context"
	"testing"
)

// Phase 10.3 — unit tests for the concrete adapter's guard rails.
// The full happy-path is covered by integration through the callback flow
// (see credential_backfill_wire_test.go for the wire contract).

func TestBackfillAdapter_RejectsMissingTenant(t *testing.T) {
	err := NewCredentialBackfillAdapter().TriggerBackfill(context.Background(), "", "shopee", "s1")
	if err == nil {
		t.Fatalf("expected error for empty tenant")
	}
}

func TestBackfillAdapter_RejectsMissingPlatform(t *testing.T) {
	err := NewCredentialBackfillAdapter().TriggerBackfill(context.Background(), "t1", "", "s1")
	if err == nil {
		t.Fatalf("expected error for empty platform")
	}
}

func TestBackfillAdapter_RejectsMissingStore(t *testing.T) {
	err := NewCredentialBackfillAdapter().TriggerBackfill(context.Background(), "t1", "shopee", "")
	if err == nil {
		t.Fatalf("expected error for empty store identifier")
	}
}

func TestBackfillAdapter_WindowIsThirtyDays(t *testing.T) {
	if BackfillWindowDays != 30 {
		t.Fatalf("expected 30-day window, got %d", BackfillWindowDays)
	}
}

func TestBackfillCategories_CoversUnpaidUnprocessProcessed(t *testing.T) {
	want := map[OrderStatusCategory]bool{
		StatusUnpaid:    true,
		StatusUnprocess: true,
		StatusProcessed: true,
	}
	if len(backfillCategories) != len(want) {
		t.Fatalf("expected %d categories, got %d", len(want), len(backfillCategories))
	}
	for _, c := range backfillCategories {
		if !want[c] {
			t.Fatalf("unexpected category %q in backfill list", c)
		}
	}
}
