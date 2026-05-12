package handlers

import (
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/services/cache"
	"github.com/omni/backend/internal/services/webhooks"
	"github.com/rs/zerolog/log"
)

	const (
	// MaxWebhookBodySize is the maximum allowed webhook request body (1MB)
	MaxWebhookBodySize = 1 * 1024 * 1024
	// WebhookTimestampTolerance is the max age of a webhook timestamp for replay protection
	WebhookTimestampTolerance = 5 * 60 // 5 minutes in seconds
)

// webhookTenantPattern validates tenant_id format to prevent injection attacks.
// Webhooks receive tenant_id from query params (external platforms don't have JWT),
// so we MUST validate the format here as defense-in-depth.
var webhookTenantPattern = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)

// extractAndValidateWebhookTenant extracts tenant_id from query/header and validates format.
// Returns empty string if invalid (caller should still process webhook but skip tenant-specific logic).
func extractAndValidateWebhookTenant(c *gin.Context) string {
	tenantID := c.Query("tenant_id")
	if tenantID == "" {
		tenantID = c.GetHeader("x-tenant-id")
	}
	if tenantID != "" && !webhookTenantPattern.MatchString(tenantID) {
		log.Warn().Str("tenant_id", tenantID).Msg("[Webhook] Invalid tenant_id format rejected")
		return ""
	}
	return tenantID
}

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
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, MaxWebhookBodySize))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	// Get signature and validate tenant from headers/params
	signature := c.GetHeader("Authorization")
	tenantID := extractAndValidateWebhookTenant(c)

	log.Debug().Str("tenant_id", tenantID).Msg("Shopee webhook received")

	// Fail-closed: reject if no processor configured
	if h.shopeeProcessor == nil {
		log.Error().Msg("[Webhook] Shopee processor not configured")
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "webhook processor not configured"})
		return
	}

	// Fail-closed: reject if no tenant
	if tenantID == "" {
		log.Warn().Msg("[Webhook] Shopee webhook missing or invalid tenant_id")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "missing or invalid tenant_id"})
		return
	}

	// Process webhook
	requestURL := c.Request.URL.String()
	if err := h.shopeeProcessor.Process(c.Request.Context(), tenantID, requestURL, string(body), signature); err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Shopee webhook processing error")
		if isSignatureError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid webhook signature"})
			return
		}
		// Non-signature errors: still return 200 to prevent platform retries
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	// Success
	h.invalidateAnalyticsCache(tenantID, "shopee")
	c.JSON(http.StatusOK, gin.H{"success": true})
}
// LazadaWebhook handles Lazada webhook events
func (h *WebhookHandler) LazadaWebhook(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, MaxWebhookBodySize))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	signature := c.GetHeader("Authorization")
	tenantID := extractAndValidateWebhookTenant(c)

	log.Debug().Str("tenant_id", tenantID).Msg("Lazada webhook received")

	// Fail-closed: reject if no processor configured
	if h.lazadaProcessor == nil {
		log.Error().Msg("[Webhook] Lazada processor not configured")
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "webhook processor not configured"})
		return
	}

	// Fail-closed: reject if no tenant
	if tenantID == "" {
		log.Warn().Msg("[Webhook] Lazada webhook missing or invalid tenant_id")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "missing or invalid tenant_id"})
		return
	}

	// Process webhook
	if err := h.lazadaProcessor.Process(c.Request.Context(), tenantID, string(body), signature); err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("Lazada webhook processing error")
		if isSignatureError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid webhook signature"})
			return
		}
		// Non-signature errors: still return 200 to prevent platform retries
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	// Success
	h.invalidateAnalyticsCache(tenantID, "lazada")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// TiktokWebhook handles TikTok webhook events
func (h *WebhookHandler) TiktokWebhook(c *gin.Context) {
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, MaxWebhookBodySize))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Failed to read body"))
		return
	}

	signature := c.GetHeader("x-tts-signature")
	timestamp := c.GetHeader("x-tts-timestamp")
	tenantID := extractAndValidateWebhookTenant(c)

	log.Debug().Str("tenant_id", tenantID).Msg("TikTok webhook received")

	// Replay protection: validate timestamp is within tolerance
	if timestamp != "" {
		ts, err := strconv.ParseInt(timestamp, 10, 64)
		if err != nil {
			log.Warn().Str("timestamp", timestamp).Msg("[Webhook] TikTok invalid timestamp format")
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "invalid timestamp"})
			return
		}
		now := time.Now().Unix()
		if abs(now-ts) > WebhookTimestampTolerance {
			log.Warn().Int64("timestamp", ts).Int64("now", now).Msg("[Webhook] TikTok webhook timestamp outside tolerance (replay?)")
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "webhook timestamp expired"})
			return
		}
	}

	// Fail-closed: reject if no processor configured
	if h.tiktokProcessor == nil {
		log.Error().Msg("[Webhook] TikTok processor not configured")
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "webhook processor not configured"})
		return
	}

	// Fail-closed: reject if no tenant
	if tenantID == "" {
		log.Warn().Msg("[Webhook] TikTok webhook missing or invalid tenant_id")
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "missing or invalid tenant_id"})
		return
	}

	// Process webhook
	if err := h.tiktokProcessor.Process(c.Request.Context(), tenantID, string(body), timestamp, signature); err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("TikTok webhook processing error")
		if isSignatureError(err) {
			c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "invalid webhook signature"})
			return
		}
		// Non-signature errors: still return 200 to prevent platform retries
		c.JSON(http.StatusOK, gin.H{"success": true})
		return
	}

	// Success
	h.invalidateAnalyticsCache(tenantID, "tiktok")
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// abs returns absolute value of int64
func abs(n int64) int64 {
	if n < 0 {
		return -n
	}
	return n
}
