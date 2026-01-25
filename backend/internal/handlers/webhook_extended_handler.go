package handlers

import (
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/webhooks"
)

// WebhookExtendedHandler handles extended webhook endpoints
type WebhookExtendedHandler struct {
	shopeeProcessor *webhooks.ShopeeWebhookProcessor
	lazadaProcessor *webhooks.LazadaWebhookProcessor
	tiktokProcessor *webhooks.TiktokWebhookProcessor
	basePath        string
}

// NewWebhookExtendedHandler creates a new webhook extended handler
func NewWebhookExtendedHandler(
	shopeeProcessor *webhooks.ShopeeWebhookProcessor,
	lazadaProcessor *webhooks.LazadaWebhookProcessor,
	tiktokProcessor *webhooks.TiktokWebhookProcessor,
	basePath string,
) *WebhookExtendedHandler {
	return &WebhookExtendedHandler{
		shopeeProcessor: shopeeProcessor,
		lazadaProcessor: lazadaProcessor,
		tiktokProcessor: tiktokProcessor,
		basePath:        basePath,
	}
}

// ShopeeWebhookTenant handles POST /api/webhooks/:tenantId/shopee
func (h *WebhookExtendedHandler) ShopeeWebhookTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing tenantId"))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	signature := c.GetHeader("Authorization")
	requestURL := c.Request.URL.String()

	log.Printf("[Webhook] Shopee tenant-specific: tenant=%s", tenantID)

	if h.shopeeProcessor != nil {
		if err := h.shopeeProcessor.Process(c.Request.Context(), tenantID, requestURL, string(body), signature); err != nil {
			log.Printf("[Webhook] Shopee processing error: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// LazadaWebhookTenant handles POST /api/webhooks/:tenantId/lazada
func (h *WebhookExtendedHandler) LazadaWebhookTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing tenantId"))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	signature := c.GetHeader("Authorization")

	log.Printf("[Webhook] Lazada tenant-specific: tenant=%s", tenantID)

	if h.lazadaProcessor != nil {
		if err := h.lazadaProcessor.Process(c.Request.Context(), tenantID, string(body), signature); err != nil {
			log.Printf("[Webhook] Lazada processing error: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TiktokWebhookTenant handles POST /api/webhooks/:tenantId/tiktok
func (h *WebhookExtendedHandler) TiktokWebhookTenant(c *gin.Context) {
	tenantID := c.Param("tenantId")
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, response.Error("Missing tenantId"))
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	signature := c.GetHeader("x-tts-signature")
	timestamp := c.GetHeader("x-tts-timestamp")

	log.Printf("[Webhook] TikTok tenant-specific: tenant=%s", tenantID)

	if h.tiktokProcessor != nil {
		if err := h.tiktokProcessor.Process(c.Request.Context(), tenantID, string(body), timestamp, signature); err != nil {
			log.Printf("[Webhook] TikTok processing error: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TestWebhookRequest represents test webhook request
type TestWebhookRequest struct {
	Platform  string                 `json:"platform" binding:"required"`
	EventType string                 `json:"eventType" binding:"required"`
	Data      map[string]interface{} `json:"data"`
}

// TestWebhook handles POST /api/webhooks/test
func (h *WebhookExtendedHandler) TestWebhook(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req TestWebhookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request"))
		return
	}

	// Simulate webhook processing
	result := gin.H{
		"platform":  req.Platform,
		"eventType": req.EventType,
		"processed": true,
		"message":   "Test webhook processed successfully",
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// WebhookConfigResponse represents webhook configuration
type WebhookConfigResponse struct {
	Shopee WebhookPlatformConfig `json:"shopee"`
	Lazada WebhookPlatformConfig `json:"lazada"`
	Tiktok WebhookPlatformConfig `json:"tiktok"`
}

// WebhookPlatformConfig represents platform webhook config
type WebhookPlatformConfig struct {
	Enabled    bool   `json:"enabled"`
	WebhookURL string `json:"webhookUrl"`
	Events     []string `json:"events"`
}

// GetWebhookConfig handles GET /api/webhooks/config
func (h *WebhookExtendedHandler) GetWebhookConfig(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	// Get base webhook URL
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	baseURL := scheme + "://" + c.Request.Host

	cfg := WebhookConfigResponse{
		Shopee: WebhookPlatformConfig{
			Enabled:    true,
			WebhookURL: baseURL + "/api/webhooks/" + tenantID + "/shopee",
			Events:     []string{"order_status_update", "item_promotion_update", "shop_authorization"},
		},
		Lazada: WebhookPlatformConfig{
			Enabled:    true,
			WebhookURL: baseURL + "/api/webhooks/" + tenantID + "/lazada",
			Events:     []string{"order_status_update", "seller_product_update"},
		},
		Tiktok: WebhookPlatformConfig{
			Enabled:    true,
			WebhookURL: baseURL + "/api/webhooks/" + tenantID + "/tiktok",
			Events:     []string{"ORDER_STATUS_CHANGE", "PRODUCT_STATUS_CHANGE", "PACKAGE_UPDATE"},
		},
	}

	c.JSON(http.StatusOK, response.Success(cfg))
}
