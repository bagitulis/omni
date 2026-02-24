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

// LazadaWebhookProcessor processes Lazada webhooks
type LazadaWebhookProcessor struct {
	webhookRepo *repositories.WebhookRepository
	oauthSvc    *oauth.LazadaOAuthService
}

// NewLazadaWebhookProcessor creates a new Lazada webhook processor
func NewLazadaWebhookProcessor(repo *repositories.WebhookRepository, oauthSvc *oauth.LazadaOAuthService) *LazadaWebhookProcessor {
	return &LazadaWebhookProcessor{
		webhookRepo: repo,
		oauthSvc:    oauthSvc,
	}
}

// LazadaWebhookPayload represents Lazada webhook payload
type LazadaWebhookPayload struct {
	MessageType string                 `json:"message_type"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   int64                  `json:"timestamp"`
	SellerID    string                 `json:"seller_id"`
}

// Process processes a Lazada webhook
func (p *LazadaWebhookProcessor) Process(ctx context.Context, tenantID, body, signature string) error {
	// Verify signature
	if p.oauthSvc != nil && !p.oauthSvc.VerifyWebhookSignature(body, signature) {
		return fmt.Errorf("invalid webhook signature")
	}

	// Parse payload
	var payload LazadaWebhookPayload
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return fmt.Errorf("failed to parse webhook payload: %w", err)
	}

	// Log webhook
	log, err := p.webhookRepo.CreateLog(ctx, tenantID, models.PlatformLazada, payload.MessageType, payload, nil)
	if err != nil {
		return err
	}

	// Process based on message type
	var processErr error
	switch payload.MessageType {
	case "ORDER_CREATED", "ORDER_STATUS_CHANGED", "ORDER_ITEMS_STATUS_CHANGED":
		processErr = p.processOrderEvent(ctx, tenantID, payload)
	case "PRODUCT_CREATED", "PRODUCT_UPDATED", "PRODUCT_REMOVED":
		processErr = p.processProductEvent(ctx, tenantID, payload)
	case "RETURN_CREATED", "RETURN_STATUS_CHANGED":
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

// processOrderEvent processes Lazada order events
func (p *LazadaWebhookProcessor) processOrderEvent(ctx context.Context, tenantID string, payload LazadaWebhookPayload) error {
	data := payload.Data
	orderSN := ""
	if orderID, ok := data["order_id"].(float64); ok && orderID != 0 {
		orderSN = fmt.Sprintf("%.0f", orderID)
	}
	if orderSN == "" {
		orderIDStr, _ := data["order_id"].(string)
		if orderIDStr == "" {
			return fmt.Errorf("missing order_id in payload")
		}
		orderSN = orderIDStr
	}

	status, _ := data["status"].(string)

	event := &models.WebhookOrderEvent{
		TenantID:  tenantID,
		Platform:  models.PlatformLazada,
		OrderSN:   orderSN,
		EventType: payload.MessageType,
		NewStatus: status,
		ShopID:    payload.SellerID,
		CreatedAt: time.Now(),
	}

	return p.webhookRepo.CreateOrderEvent(ctx, event)
}

// processProductEvent processes Lazada product events
func (p *LazadaWebhookProcessor) processProductEvent(ctx context.Context, tenantID string, payload LazadaWebhookPayload) error {
	data := payload.Data
	productID, _ := data["item_id"].(float64)

	event := &models.WebhookProductEvent{
		TenantID:  tenantID,
		ItemID:    fmt.Sprintf("%.0f", productID),
		EventType: payload.MessageType,
		ShopID:    payload.SellerID,
		CreatedAt: time.Now(),
	}

	return p.webhookRepo.CreateProductEvent(ctx, event)
}

// processReturnEvent processes Lazada return events
func (p *LazadaWebhookProcessor) processReturnEvent(ctx context.Context, tenantID string, payload LazadaWebhookPayload) error {
	data := payload.Data
	returnID, _ := data["reverse_order_id"].(float64)
	orderID, _ := data["order_id"].(float64)
	reason, _ := data["reason"].(string)
	status, _ := data["status"].(string)

	event := &models.WebhookReturnEvent{
		TenantID:  tenantID,
		ReturnSN:  fmt.Sprintf("%.0f", returnID),
		OrderSN:   fmt.Sprintf("%.0f", orderID),
		Reason:    reason,
		Status:    status,
		ShopID:    payload.SellerID,
		EventType: payload.MessageType,
		CreatedAt: time.Now(),
	}

	return p.webhookRepo.CreateReturnEvent(ctx, event)
}
