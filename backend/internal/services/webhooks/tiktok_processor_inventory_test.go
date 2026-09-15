package webhooks

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/realtime"
)

// TestTiktokWebhookProcessor_InventoryUpdate — TikTok added an Inventory
// Update Webhook in Nov 2025 (partner.tiktokshop.com "New Feature: Inventory
// Update Webhook"). Our processor should recognise the type code and fan the
// event out via realtime so dashboards invalidate stock queries without
// waiting for the next poll.
//
// This is a RED-first test: written before the code path was added; the
// getEventType switch had no case for the code, and no publisher wire
// existed either.
func TestTiktokWebhookProcessor_InventoryUpdate_TypeName(t *testing.T) {
	p := &TiktokWebhookProcessor{}
	got := p.getEventType(TiktokWebhookInventoryUpdate)
	if got != "inventory_update" {
		t.Errorf("getEventType(TiktokWebhookInventoryUpdate) = %q, want inventory_update", got)
	}
}

// TestTiktokWebhookProcessor_InventoryUpdate_PublishesRealtime — assert the
// inventory processor calls the realtime publisher. Uses the SkippedNoHub
// counter (no hub wired in unit test) as the tick.
func TestTiktokWebhookProcessor_InventoryUpdate_PublishesRealtime(t *testing.T) {
	realtime.ResetForTests()
	pub := realtime.Get()

	// Craft a minimal payload — processInventoryEvent only needs
	// product_id / sku_id / stock_qty for its publish. The webhook-repo call
	// on the full Process() path needs a DB; we call the inner method
	// directly for the unit test.
	payload := TiktokWebhookPayload{
		Type:       TiktokWebhookInventoryUpdate,
		ShopID:     "shop-42",
		ShopCipher: "cipher-42",
		Timestamp:  1700000000,
		Data: map[string]interface{}{
			"product_id": "P-123",
			"sku_id":     "S-456",
			"stock_qty":  json.Number("7"),
		},
	}
	p := &TiktokWebhookProcessor{}
	if err := p.processInventoryEvent(context.Background(), "tenant-a", payload); err != nil {
		t.Fatalf("processInventoryEvent errored: %v", err)
	}
	published, skipped := pub.Stats()
	// No hub is wired in a unit test, so the publish becomes a skip. That's
	// exactly the proof we want: the call site was hit.
	if published != 0 {
		t.Errorf("published = %d, want 0 (no hub wired)", published)
	}
	if skipped != 1 {
		t.Errorf("skipped = %d, want 1 (inventory update must reach publisher)", skipped)
	}
}

// TestTiktokWebhookProcessor_InventoryUpdate_MissingProductID — negative
// path: a payload without product_id must return a descriptive error and
// must NOT publish.
func TestTiktokWebhookProcessor_InventoryUpdate_MissingProductID(t *testing.T) {
	realtime.ResetForTests()
	pub := realtime.Get()

	payload := TiktokWebhookPayload{
		Type: TiktokWebhookInventoryUpdate,
		Data: map[string]interface{}{
			// product_id intentionally missing
			"sku_id":    "S-1",
			"stock_qty": json.Number("1"),
		},
	}
	p := &TiktokWebhookProcessor{}
	err := p.processInventoryEvent(context.Background(), "tenant-a", payload)
	if err == nil {
		t.Fatal("expected error for missing product_id, got nil")
	}
	published, skipped := pub.Stats()
	if published != 0 || skipped != 0 {
		t.Errorf("counters non-zero on error path: published=%d skipped=%d", published, skipped)
	}
}
