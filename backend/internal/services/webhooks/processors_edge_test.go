package webhooks

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestShopeeGetEventType_UnknownCode(t *testing.T) {
	p := NewShopeeWebhookProcessor(nil, nil)

	got := p.getEventType(999)
	if got != "unknown_999" {
		t.Fatalf("expected unknown_999, got %s", got)
	}
}

func TestShopeeProcessOrderEvent_MissingOrderSN(t *testing.T) {
	p := NewShopeeWebhookProcessor(nil, nil)

	err := p.processOrderEvent(context.Background(), "tenant-1", ShopeeWebhookPayload{
		Code: models.ShopeePushOrderStatus,
		Data: map[string]interface{}{},
	})

	if err == nil {
		t.Fatal("expected missing ordersn error")
	}
}

func TestTiktokGetEventType_UnknownCode(t *testing.T) {
	p := NewTiktokWebhookProcessor(nil, nil)

	got := p.getEventType(999)
	if got != "unknown_999" {
		t.Fatalf("expected unknown_999, got %s", got)
	}
}

func TestTiktokProcessOrderEvent_MissingOrderID(t *testing.T) {
	p := NewTiktokWebhookProcessor(nil, nil)

	err := p.processOrderEvent(context.Background(), "tenant-1", TiktokWebhookPayload{
		Type: TiktokWebhookOrderStatusChange,
		Data: map[string]interface{}{},
	})

	if err == nil {
		t.Fatal("expected missing order_id error")
	}
}

func TestLazadaProcessOrderEvent_MissingOrderID(t *testing.T) {
	p := NewLazadaWebhookProcessor(nil, nil)

	err := p.processOrderEvent(context.Background(), "tenant-1", LazadaWebhookPayload{
		MessageType: "ORDER_CREATED",
		Data:        map[string]interface{}{},
	})

	if err == nil {
		t.Fatal("expected missing order_id error")
	}
}

func TestLazadaProcessOrderEvent_StringOrderID_PersistsOriginalValue(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.WebhookOrderEvent{})
	repo := repositories.NewWebhookRepository(db)
	p := NewLazadaWebhookProcessor(repo, nil)

	ctx := context.Background()
	tenantID := "tenant-lazada-string-order-id"
	orderID := "LAZADA-ORDER-STR-1001"

	err := p.processOrderEvent(ctx, tenantID, LazadaWebhookPayload{
		MessageType: "ORDER_CREATED",
		SellerID:    "seller-123",
		Data: map[string]interface{}{
			"order_id": orderID,
			"status":   "packed",
		},
	})
	require.NoError(t, err)

	events, findErr := repo.FindOrderEventsByOrder(ctx, tenantID, orderID)
	require.NoError(t, findErr)
	require.Len(t, events, 1)
	require.Equal(t, orderID, events[0].OrderSN)
	require.Equal(t, "packed", events[0].NewStatus)
	require.Equal(t, "seller-123", events[0].ShopID)

	zeroEvents, zeroErr := repo.FindOrderEventsByOrder(ctx, tenantID, "0")
	require.NoError(t, zeroErr)
	require.Len(t, zeroEvents, 0)
}
