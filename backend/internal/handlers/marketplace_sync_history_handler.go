package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// MarketplaceSyncHistoryHandler handles marketplace sync history endpoints
type MarketplaceSyncHistoryHandler struct {
	repo *repositories.MarketplaceSyncHistoryRepo
}

// NewMarketplaceSyncHistoryHandler creates a new marketplace sync history handler
func NewMarketplaceSyncHistoryHandler(repo *repositories.MarketplaceSyncHistoryRepo) *MarketplaceSyncHistoryHandler {
	return &MarketplaceSyncHistoryHandler{repo: repo}
}

// List handles GET /api/marketplace-sync-history
// Query params: page, page_size, platform, operation, status, sku_search, date_from, date_to
func (h *MarketplaceSyncHistoryHandler) List(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := models.MarketplaceSyncHistoryFilter{
		TenantID:  tenantID,
		Platform:  c.Query("platform"),
		Operation: c.Query("operation"),
		Status:    c.Query("status"),
		SKUSearch: c.Query("sku_search"),
		DateFrom:  c.Query("date_from"),
		DateTo:    c.Query("date_to"),
		Page:      page,
		PageSize:  pageSize,
	}

	result, err := h.repo.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// createSyncHistoryRequest is the POST body for creating a sync history entry
type createSyncHistoryRequest struct {
	SKU          string  `json:"sku" binding:"required"`
	Platform     string  `json:"platform" binding:"required"`
	Operation    string  `json:"operation" binding:"required"`
	Status       string  `json:"status" binding:"required"`
	RequestData  *string `json:"request_data,omitempty"`
	ResponseData *string `json:"response_data,omitempty"`
	ErrorMessage *string `json:"error_message,omitempty"`
}

// Create handles POST /api/marketplace-sync-history
func (h *MarketplaceSyncHistoryHandler) Create(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenant_id",
		})
		return
	}

	var req createSyncHistoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate enum fields
	if !models.IsValidSyncPlatform(req.Platform) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid platform: %q. Allowed values: %s", req.Platform, strings.Join(models.AllowedSyncPlatforms, ", ")),
		})
		return
	}
	if !models.IsValidSyncOperation(req.Operation) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid operation: %q. Allowed values: %s", req.Operation, strings.Join(models.AllowedSyncOperations, ", ")),
		})
		return
	}
	if !models.IsValidSyncStatus(req.Status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Invalid status: %q. Allowed values: %s", req.Status, strings.Join(models.AllowedSyncStatuses, ", ")),
		})
		return
	}

	entry := &models.MarketplaceSyncHistory{
		TenantID:     tenantID,
		SKU:          req.SKU,
		Platform:     req.Platform,
		Operation:    req.Operation,
		Status:       req.Status,
		RequestData:  req.RequestData,
		ResponseData: req.ResponseData,
		ErrorMessage: req.ErrorMessage,
	}

	if err := h.repo.Create(c.Request.Context(), entry); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    entry,
	})
}
