package webhooks_test

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/services/webhooks"
)

// TestNewShopeeWebhookProcessor verifies constructor returns non-nil with nil deps.
func TestNewShopeeWebhookProcessor(t *testing.T) {
	p := webhooks.NewShopeeWebhookProcessor(nil, nil)
	if p == nil {
		t.Fatal("expected non-nil ShopeeWebhookProcessor")
	}
}

// TestNewLazadaWebhookProcessor verifies constructor returns non-nil with nil deps.
func TestNewLazadaWebhookProcessor(t *testing.T) {
	p := webhooks.NewLazadaWebhookProcessor(nil, nil)
	if p == nil {
		t.Fatal("expected non-nil LazadaWebhookProcessor")
	}
}

// TestNewTiktokWebhookProcessor verifies constructor returns non-nil with nil deps.
func TestNewTiktokWebhookProcessor(t *testing.T) {
	p := webhooks.NewTiktokWebhookProcessor(nil, nil)
	if p == nil {
		t.Fatal("expected non-nil TiktokWebhookProcessor")
	}
}

// TestShopeeProcessor_InvalidJSON verifies JSON parse error is returned before DB access.
// oauthSvc is nil → signature check is skipped entirely.
// webhookRepo is nil; we must not reach CreateLog.
func TestShopeeProcessor_InvalidJSON(t *testing.T) {
	p := webhooks.NewShopeeWebhookProcessor(nil, nil)
	ctx := context.Background()

	err := p.Process(ctx, "tenant1", "http://example.com", "not-valid-json", "sig")
	if err == nil {
		t.Fatal("expected error for invalid JSON body, got nil")
	}
}

// TestLazadaProcessor_InvalidJSON verifies JSON parse error is returned before DB access.
func TestLazadaProcessor_InvalidJSON(t *testing.T) {
	p := webhooks.NewLazadaWebhookProcessor(nil, nil)
	ctx := context.Background()

	err := p.Process(ctx, "tenant1", "not-valid-json", "sig")
	if err == nil {
		t.Fatal("expected error for invalid JSON body, got nil")
	}
}

// TestTiktokProcessor_InvalidJSON verifies JSON parse error is returned before DB access.
func TestTiktokProcessor_InvalidJSON(t *testing.T) {
	p := webhooks.NewTiktokWebhookProcessor(nil, nil)
	ctx := context.Background()

	err := p.Process(ctx, "tenant1", "not-valid-json", "12345", "sig")
	if err == nil {
		t.Fatal("expected error for invalid JSON body, got nil")
	}
}

// TestTiktokWebhookConstants verifies the TikTok webhook type constants have expected values.
func TestTiktokWebhookConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      int
		expected int
	}{
		{"OrderStatusChange", webhooks.TiktokWebhookOrderStatusChange, 1},
		{"OrderShipment", webhooks.TiktokWebhookOrderShipment, 2},
		{"ProductUpdate", webhooks.TiktokWebhookProductUpdate, 3},
		{"ReturnCreated", webhooks.TiktokWebhookReturnCreated, 4},
		{"ReturnStatusChange", webhooks.TiktokWebhookReturnStatusChange, 5},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("TiktokWebhook%s = %d, want %d", tt.name, tt.got, tt.expected)
			}
		})
	}
}
