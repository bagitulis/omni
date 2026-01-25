package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
)

// ShopeeWebhookProcessor processes Shopee webhooks
type ShopeeWebhookProcessor struct {
	webhookRepo *repositories.WebhookRepository
	oauthSvc    *oauth.ShopeeOAuthService
}

// NewShopeeWebhookProcessor creates a new Shopee webhook processor
func NewShopeeWebhookProcessor(repo *repositories.WebhookRepository, oauthSvc *oauth.ShopeeOAuthService) *ShopeeWebhookProcessor {
	return &ShopeeWebhookProcessor{
		webhookRepo: repo,
		oauthSvc:    oauthSvc,
	}
}

// ShopeeWebhookPayload represents Shopee webhook payload
type ShopeeWebhookPayload struct {
	Code      int                    `json:"code"`
	ShopID    int64                  `json:"shop_id"`
	Timestamp int64                  `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Process processes a Shopee webhook
func (p *ShopeeWebhookProcessor) Process(ctx context.Context, tenantID, requestURL, body, signature string) error {
	// Verify signature
	if p.oauthSvc != nil && !p.oauthSvc.VerifyWebhookSignature(requestURL, body, signature) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Parse payload
	var payload ShopeeWebhookPayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Determine event type from code
	eventType := p.getEventType(payload.Code)

	// Log webhook
	log, err := p.webhookRepo.CreateLog(ctx, tenantID, models.PlatformShopee, eventType, payload, nil)
	if err != nil {
		return err
	}

	// Process based on event type
	var processErr error
	switch payload.Code {
	case models.ShopeePushOrderStatus, models.ShopeePushTracking:
		processErr = p.processOrderEvent(ctx, tenantID, payload)
	case models.ShopeePushShopeeUpdate, models.ShopeePushPromotion:
		processErr = p.processProductEvent(ctx, tenantID, payload)
	default:
		// Unknown event type, just log it
	}

	// Update log status
	if processErr != nil {
		p.webhookRepo.UpdateLogStatus(ctx, log.ID, models.WebhookStatusFailed, processErr.Error())
		return processErr
	}

	p.webhookRepo.UpdateLogStatus(ctx, log.ID, models.WebhookStatusProcessed, "")
	return nil
}

// getEventType returns event type string from code
func (p *ShopeeWebhookProcessor) getEventType(code int) string {
	switch code {
	case models.ShopeePushOrderStatus:
		return "order_status_update"
	case models.ShopeePushTracking:
		return "tracking_no_update"
	case models.ShopeePushShopeeUpdate:
		return "product_update"
	case models.ShopeePushPromotion:
		return "promotion_update"
	case models.ShopeePushShopUpdate:
		return "shop_update"
	case models.ShopeePushWebchat:
		return "chat_update"
	default:
		return fmt.Sprintf("unknown_%d", code)
	}
}

// processOrderEvent processes order-related webhook events
func (p *ShopeeWebhookProcessor) processOrderEvent(ctx context.Context, tenantID string, payload ShopeeWebhookPayload) error {
	data := payload.Data
	orderSN, _ := data["ordersn"].(string)
	if orderSN == "" {
		return fmt.Errorf("missing ordersn in payload")
	}

	newStatus, _ := data["status"].(string)

	event := &models.WebhookOrderEvent{
		TenantID:  tenantID,
		Platform:  models.PlatformShopee,
		OrderSN:   orderSN,
		EventType: p.getEventType(payload.Code),
		NewStatus: newStatus,
		ShopID:    fmt.Sprintf("%d", payload.ShopID),
		CreatedAt: time.Now(),
	}

	return p.webhookRepo.CreateOrderEvent(ctx, event)
}

// processProductEvent processes product-related webhook events
func (p *ShopeeWebhookProcessor) processProductEvent(ctx context.Context, tenantID string, payload ShopeeWebhookPayload) error {
	data := payload.Data
	itemID, _ := data["item_id"].(float64)

	event := &models.WebhookProductEvent{
		TenantID:  tenantID,
		ItemID:    fmt.Sprintf("%.0f", itemID),
		EventType: p.getEventType(payload.Code),
		ShopID:    fmt.Sprintf("%d", payload.ShopID),
		CreatedAt: time.Now(),
	}

	return p.webhookRepo.CreateProductEvent(ctx, event)
}
