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

// TiktokWebhookProcessor processes TikTok webhooks
type TiktokWebhookProcessor struct {
	webhookRepo *repositories.WebhookRepository
	oauthSvc    *oauth.TiktokOAuthService
}

// NewTiktokWebhookProcessor creates a new TikTok webhook processor
func NewTiktokWebhookProcessor(repo *repositories.WebhookRepository, oauthSvc *oauth.TiktokOAuthService) *TiktokWebhookProcessor {
	return &TiktokWebhookProcessor{
		webhookRepo: repo,
		oauthSvc:    oauthSvc,
	}
}

// TiktokWebhookPayload represents TikTok webhook payload
type TiktokWebhookPayload struct {
	Type       int                    `json:"type"`
	ShopID     string                 `json:"shop_id"`
	ShopCipher string                 `json:"shop_cipher"`
	Timestamp  int64                  `json:"timestamp"`
	Data       map[string]interface{} `json:"data"`
}

// TikTok webhook types
const (
	TiktokWebhookOrderStatusChange  = 1
	TiktokWebhookOrderShipment      = 2
	TiktokWebhookProductUpdate      = 3
	TiktokWebhookReturnCreated      = 4
	TiktokWebhookReturnStatusChange = 5
)

// Process processes a TikTok webhook
func (p *TiktokWebhookProcessor) Process(ctx context.Context, tenantID, body, timestamp, signature string) error {
	// Verify signature
	if p.oauthSvc != nil && !p.oauthSvc.VerifyWebhookSignature(body, timestamp, signature) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Parse payload
	var payload TiktokWebhookPayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Determine event type
	eventType := p.getEventType(payload.Type)

	// Log webhook
	log, err := p.webhookRepo.CreateLog(ctx, tenantID, models.PlatformTiktok, eventType, payload, nil)
	if err != nil {
		return err
	}

	// Process based on type
	var processErr error
	switch payload.Type {
	case TiktokWebhookOrderStatusChange, TiktokWebhookOrderShipment:
		processErr = p.processOrderEvent(ctx, tenantID, payload)
	case TiktokWebhookProductUpdate:
		processErr = p.processProductEvent(ctx, tenantID, payload)
	case TiktokWebhookReturnCreated, TiktokWebhookReturnStatusChange:
		processErr = p.processReturnEvent(ctx, tenantID, payload)
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

// getEventType returns event type string from type code
func (p *TiktokWebhookProcessor) getEventType(typeCode int) string {
	switch typeCode {
	case TiktokWebhookOrderStatusChange:
		return "order_status_change"
	case TiktokWebhookOrderShipment:
		return "order_shipment"
	case TiktokWebhookProductUpdate:
		return "product_update"
	case TiktokWebhookReturnCreated:
		return "return_created"
	case TiktokWebhookReturnStatusChange:
		return "return_status_change"
	default:
		return fmt.Sprintf("unknown_%d", typeCode)
	}
}

// processOrderEvent processes TikTok order events
func (p *TiktokWebhookProcessor) processOrderEvent(ctx context.Context, tenantID string, payload TiktokWebhookPayload) error {
	data := payload.Data
	orderID, _ := data["order_id"].(string)
	if orderID == "" {
		return fmt.Errorf("missing order_id in payload")
	}

	status, _ := data["order_status"].(float64)
	statusStr := fmt.Sprintf("%.0f", status)

	event := &models.WebhookOrderEvent{
		TenantID:  tenantID,
		Platform:  models.PlatformTiktok,
		OrderSN:   orderID,
		EventType: p.getEventType(payload.Type),
		NewStatus: statusStr,
		ShopID:    payload.ShopID,
		CreatedAt: time.Now(),
	}

	return p.webhookRepo.CreateOrderEvent(ctx, event)
}

// processProductEvent processes TikTok product events
func (p *TiktokWebhookProcessor) processProductEvent(ctx context.Context, tenantID string, payload TiktokWebhookPayload) error {
	data := payload.Data
	productID, _ := data["product_id"].(string)

	event := &models.WebhookProductEvent{
		TenantID:  tenantID,
		ItemID:    productID,
		EventType: p.getEventType(payload.Type),
		ShopID:    payload.ShopID,
		CreatedAt: time.Now(),
	}

	return p.webhookRepo.CreateProductEvent(ctx, event)
}

// processReturnEvent processes TikTok return events
func (p *TiktokWebhookProcessor) processReturnEvent(ctx context.Context, tenantID string, payload TiktokWebhookPayload) error {
	data := payload.Data
	returnID, _ := data["return_id"].(string)
	orderID, _ := data["order_id"].(string)
	reason, _ := data["reason"].(string)
	status, _ := data["status"].(string)

	event := &models.WebhookReturnEvent{
		TenantID:  tenantID,
		ReturnSN:  returnID,
		OrderSN:   orderID,
		Reason:    reason,
		Status:    status,
		ShopID:    payload.ShopID,
		EventType: p.getEventType(payload.Type),
		CreatedAt: time.Now(),
	}

	return p.webhookRepo.CreateReturnEvent(ctx, event)
}
