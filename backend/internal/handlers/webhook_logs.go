package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// GetLogs handles GET /api/webhooks/logs
func (h *WebhookHandler) GetLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	page, pageSize := parseWebhookPagination(c)
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

	c.JSON(http.StatusOK, response.SuccessWithMeta(logs, buildWebhookMeta(int(total), page, pageSize)))
}

// GetLogsByPlatform handles GET /api/webhooks/logs/:platform
func (h *WebhookHandler) GetLogsByPlatform(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	platform := c.Param("platform")
	page, pageSize := parseWebhookPagination(c)
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

	c.JSON(http.StatusOK, response.SuccessWithMeta(events, buildWebhookMeta(int(total), page, pageSize)))
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
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
	db.WithContext(ctx).Model(&models.WebhookLog{}).Where("tenant_id = ?", tenantID).Count(&stats.Total)

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

// parseWebhookPagination extracts and validates pagination parameters
func parseWebhookPagination(c *gin.Context) (page, pageSize int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ = strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	return
}

// buildWebhookMeta builds pagination metadata
func buildWebhookMeta(total, page, pageSize int) *response.Meta {
	totalPages := total / pageSize
	if total%pageSize > 0 {
		totalPages++
	}
	return &response.Meta{
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}
