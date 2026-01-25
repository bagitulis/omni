package handlers

import (
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/webhooks"
)

// WebhookHandler handles webhook endpoints
type WebhookHandler struct {
	shopeeProcessor *webhooks.ShopeeWebhookProcessor
	lazadaProcessor *webhooks.LazadaWebhookProcessor
	tiktokProcessor *webhooks.TiktokWebhookProcessor
	basePath        string
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

	log.Printf("[Webhook] Shopee received: tenant=%s", tenantID)

	// Process webhook
	if h.shopeeProcessor != nil && tenantID != "" {
		requestURL := c.Request.URL.String()
		if err := h.shopeeProcessor.Process(c.Request.Context(), tenantID, requestURL, string(body), signature); err != nil {
			log.Printf("[Webhook] Shopee processing error: %v", err)
			// Still return success to prevent retries
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

	log.Printf("[Webhook] Lazada received: tenant=%s", tenantID)

	if h.lazadaProcessor != nil && tenantID != "" {
		if err := h.lazadaProcessor.Process(c.Request.Context(), tenantID, string(body), signature); err != nil {
			log.Printf("[Webhook] Lazada processing error: %v", err)
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

	log.Printf("[Webhook] TikTok received: tenant=%s", tenantID)

	if h.tiktokProcessor != nil && tenantID != "" {
		if err := h.tiktokProcessor.Process(c.Request.Context(), tenantID, string(body), timestamp, signature); err != nil {
			log.Printf("[Webhook] TikTok processing error: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// GetLogs handles GET /api/webhooks/logs
func (h *WebhookHandler) GetLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewWebhookRepository(db)
	logs, total, err := repo.FindLogsByTenant(c.Request.Context(), tenantID, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch logs"))
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(logs, &response.Meta{
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}))
}

// GetLogsByPlatform handles GET /api/webhooks/logs/:platform
func (h *WebhookHandler) GetLogsByPlatform(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	platform := c.Param("platform")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewWebhookRepository(db)
	events, total, err := repo.FindOrderEventsByPlatform(c.Request.Context(), tenantID, platform, pageSize, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to fetch logs"))
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, response.SuccessWithMeta(events, &response.Meta{
		Total:      int(total),
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}))
}

// WebhookStats holds webhook statistics
type WebhookStats struct {
	Total      int64            `json:"total"`
	ByStatus   map[string]int64 `json:"by_status"`
	ByPlatform map[string]int64 `json:"by_platform"`
}

// GetStats handles GET /api/webhooks/stats
func (h *WebhookHandler) GetStats(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	var stats WebhookStats
	stats.ByStatus = make(map[string]int64)
	stats.ByPlatform = make(map[string]int64)

	ctx := c.Request.Context()
	// Get total
	db.WithContext(ctx).Model(&models.WebhookLog{}).Where("tenant_id = ?", tenantID).Count(&stats.Total)

	// Get by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	db.WithContext(ctx).Model(&models.WebhookLog{}).
		Select("status, count(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("status").
		Scan(&statusCounts)
	for _, sc := range statusCounts {
		stats.ByStatus[sc.Status] = sc.Count
	}

	// Get by platform
	var platformCounts []struct {
		Platform string
		Count    int64
	}
	db.WithContext(ctx).Model(&models.WebhookLog{}).
		Select("platform, count(*) as count").
		Where("tenant_id = ?", tenantID).
		Group("platform").
		Scan(&platformCounts)
	for _, pc := range platformCounts {
		stats.ByPlatform[pc.Platform] = pc.Count
	}

	c.JSON(http.StatusOK, response.Success(stats))
}
