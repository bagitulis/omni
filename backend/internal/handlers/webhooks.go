package handlers

import (
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/services/cache"
	"github.com/omni/backend/internal/services/webhooks"
	"github.com/rs/zerolog/log"
)

// isSignatureError checks if the error is related to webhook signature verification
func isSignatureError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "signature") ||
		strings.Contains(errMsg, "verification failed") ||
		strings.Contains(errMsg, "unauthorized")
}

// WebhookHandler handles webhook endpoints
type WebhookHandler struct {
	shopeeProcessor *webhooks.ShopeeWebhookProcessor
	lazadaProcessor *webhooks.LazadaWebhookProcessor
	tiktokProcessor *webhooks.TiktokWebhookProcessor
	basePath        string
	cacheService    cache.CacheManager
}

// NewWebhookHandler creates a new webhook handler
func NewWebhookHandler(
	shopeeProcessor *webhooks.ShopeeWebhookProcessor,
	lazadaProcessor *webhooks.LazadaWebhookProcessor,
	tiktokProcessor *webhooks.TiktokWebhookProcessor,
	basePath string,
) *WebhookHandler {
	return &WebhookHandler{
		shopeeProcessor: shopeeProcessor,
		lazadaProcessor: lazadaProcessor,
		tiktokProcessor: tiktokProcessor,
		basePath:        basePath,
	}
}

// NewWebhookHandlerWithCache creates a new webhook handler with cache service
func NewWebhookHandlerWithCache(
	shopeeProcessor *webhooks.ShopeeWebhookProcessor,
	lazadaProcessor *webhooks.LazadaWebhookProcessor,
	tiktokProcessor *webhooks.TiktokWebhookProcessor,
	basePath string,
	cacheService cache.CacheManager,
) *WebhookHandler {
	return &WebhookHandler{
		shopeeProcessor: shopeeProcessor,
		lazadaProcessor: lazadaProcessor,
		tiktokProcessor: tiktokProcessor,
		basePath:        basePath,
		cacheService:    cacheService,
	}
}

// invalidateAnalyticsCache invalidates analytics cache after webhook processing
func (h *WebhookHandler) invalidateAnalyticsCache(tenantID, platform string) {
	if h.cacheService == nil || tenantID == "" {
		return
	}

	// Invalidate relevant cache keys
	keys := []string{
		"analytics:unified:summary",
	}

	// Add platform-specific keys
	if platform == "shopee" {
		keys = append(keys, "analytics:ml:portfolio-health:shopee")
	} else if platform == "tiktok" {
		keys = append(keys, "analytics:ml:portfolio-health:tiktok")
	}

	for _, key := range keys {
		if err := h.cacheService.Delete(tenantID, key); err != nil {
			log.Warn().
				Err(err).
				Str("tenant_id", tenantID).
				Str("cache_key", key).
				Msg("Failed to invalidate cache after webhook")
		}
	}

	log.Debug().
		Str("tenant_id", tenantID).
		Str("platform", platform).
		Msg("Analytics cache invalidated after webhook")
}

// ShopeeWebhook handles Shopee webhook events
func (h *WebhookHandler) ShopeeWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	// Get signature and tenant from headers/params
	signature := c.GetHeader("Authorization")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("x-tenant-id")
	}

	log.Debug().Str("tenant_id", tenantID).Msg("Shopee webhook received")

	// Process webhook
	if h.shopeeProcessor != nil && tenantID != "" {
		requestURL := c.Request.URL.String()
		if err := h.shopeeProcessor.Process(c.Request.Context(), tenantID, requestURL, string(body), signature); err != nil {
			log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Shopee webhook processing error")
			if isSignatureError(err) {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid webhook signature"})
				return
			}
			// Still return success to prevent retries
		} else {
			// Invalidate cache on successful processing
			h.invalidateAnalyticsCache(tenantID, "shopee")
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// LazadaWebhook handles Lazada webhook events
func (h *WebhookHandler) LazadaWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	signature := c.GetHeader("Authorization")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("x-tenant-id")
	}

	log.Debug().Str("tenant_id", tenantID).Msg("Lazada webhook received")

	if h.lazadaProcessor != nil && tenantID != "" {
		if err := h.lazadaProcessor.Process(c.Request.Context(), tenantID, string(body), signature); err != nil {
			log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Lazada webhook processing error")
			if isSignatureError(err) {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid webhook signature"})
				return
			}
		} else {
			// Invalidate cache on successful processing
			h.invalidateAnalyticsCache(tenantID, "lazada")
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TiktokWebhook handles TikTok webhook events
func (h *WebhookHandler) TiktokWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	signature := c.GetHeader("x-tts-signature")
	timestamp := c.GetHeader("x-tts-timestamp")
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("x-tenant-id")
	}

	log.Debug().Str("tenant_id", tenantID).Msg("TikTok webhook received")

	if h.tiktokProcessor != nil && tenantID != "" {
		if err := h.tiktokProcessor.Process(c.Request.Context(), tenantID, string(body), timestamp, signature); err != nil {
			log.Warn().Err(err).Str("tenant_id", tenantID).Msg("TikTok webhook processing error")
			if isSignatureError(err) {
				c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid webhook signature"})
				return
			}
		} else {
			// Invalidate cache on successful processing
			h.invalidateAnalyticsCache(tenantID, "tiktok")
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
