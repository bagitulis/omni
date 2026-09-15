package services

import (
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/realtime"
)

// TestNotificationBroadcast_WiresRealtimePublisher — proves that the
// notification service now calls realtime.Get().PublishNotification when
// broadcast() runs. Uses the publisher stats counters so no hub / DB is
// needed; test-friendly and CI-portable (no testcontainers).
//
// This is a Phase-5 wire-proof: we don't test the whole broadcast (SSE side
// still needs a live DB), but we DO test the additive wire didn't get
// dropped.
func TestNotificationBroadcast_WiresRealtimePublisher(t *testing.T) {
	// Reset singleton so counters are deterministic.
	realtime.ResetForTests()
	pub := realtime.Get()

	// Baseline: nothing published, nothing skipped.
	before, beforeSkip := pub.Stats()
	if before != 0 || beforeSkip != 0 {
		t.Fatalf("baseline non-zero: pub=%d skip=%d", before, beforeSkip)
	}

	// Build a service with a valid tenant and invoke broadcast. Because no
	// hub is Wire()d, PublishNotification takes the SkippedNoHub path — that
	// is exactly the proof: the call site was hit.
	svc := &NotificationService{tenantID: "tenant-a"}
	notif := &models.Notification{
		ID:       42,
		Type:     models.NotifTypeSuccess,
		Category: "sync",
		Title:    "Test notification",
	}
	// broadcast() early-returns when SSE has no clients; that's fine — the
	// realtime publish runs BEFORE the SSE loop, so the counter must tick.
	svc.broadcast(notif)

	after, afterSkip := pub.Stats()
	if after != 0 {
		t.Errorf("published = %d, want 0 (no hub wired in test)", after)
	}
	if afterSkip != 1 {
		t.Errorf("skipped = %d, want 1 (wire call must reach publisher)", afterSkip)
	}
}

// TestNotificationBroadcast_EmptyTenantSkipsRealtime — negative path:
// when tenantID is empty, the publisher's own guard drops the message
// without incrementing either counter.
func TestNotificationBroadcast_EmptyTenantSkipsRealtime(t *testing.T) {
	realtime.ResetForTests()
	pub := realtime.Get()

	svc := &NotificationService{tenantID: ""}
	notif := &models.Notification{ID: 1, Type: models.NotifTypeError, Title: "x"}
	svc.broadcast(notif)

	published, skipped := pub.Stats()
	if published != 0 || skipped != 0 {
		t.Errorf("empty tenant path leaked: published=%d skipped=%d, want 0/0",
			published, skipped)
	}
}
