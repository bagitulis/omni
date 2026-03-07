package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/wholesale"
	"github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// WholesaleBatchHandler handles batch wholesale operations
type WholesaleBatchHandler struct {
	basePath   string
	fallbackDB *gorm.DB
}

// NewWholesaleBatchHandler creates a new batch wholesale handler
func NewWholesaleBatchHandler(basePath string, db *gorm.DB) *WholesaleBatchHandler {
	return &WholesaleBatchHandler{
		basePath:   basePath,
		fallbackDB: db,
	}
}

// getDB returns the appropriate database for the current request
func (h *WholesaleBatchHandler) getDB(_ *gin.Context, tenantID string) (*gorm.DB, error) {
	if h.fallbackDB != nil {
		return h.fallbackDB, nil
	}
	return config.GetTenantDB(tenantID, h.basePath)
}

// BatchUpdateBySkusRequest represents batch update request
type BatchUpdateBySkusRequest struct {
	Items []BatchUpdateItem `json:"items" binding:"required"`
}

// BatchUpdateItem represents a single item for batch update
type BatchUpdateItem struct {
	SKU   string  `json:"sku" binding:"required"`
	Price float64 `json:"price" binding:"required,gt=0"`
}

// BatchUpdateBySkus handles POST /api/wholesale/shopee/batch-update-skus
func (h *WholesaleBatchHandler) BatchUpdateBySkus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req BatchUpdateBySkusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("items array is required: [{ sku, price }]"))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("item_count", len(req.Items)).
		Msg("Batch update wholesale by SKUs")

	db, err := h.getDB(c, tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Database connection failed")
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Get Shopee API client
	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Shopee client")
		c.JSON(http.StatusInternalServerError, response.Error("Shopee API configuration failed"))
		return
	}

	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	service := wholesale.NewShopeeWholesaleService(db, tenantID, shopeeAPI)

	// Get wholesale settings
	settingsService := wholesale.NewWholesaleService(db, tenantID)
	settings, err := settingsService.GetSettings(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("Failed to get wholesale settings")
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get wholesale settings"))
		return
	}

	// Build SKU → Price map
	skuPriceMap := make(map[string]float64)
	for _, item := range req.Items {
		skuPriceMap[item.SKU] = item.Price
	}

	// Create tier calculator
	calculateTiers := func(basePrice float64) []wholesale.WholesaleTier {
		tiers := settingsService.CalculateTiersFromSettings(basePrice, settings)
		return tiers
	}

	result, err := service.BatchUpdateBySkus(c.Request.Context(), skuPriceMap, calculateTiers)
	if err != nil {
		log.Error().Err(err).Msg("Batch update failed")
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("processed", result.Processed).
		Int("unique_items", result.UniqueItems).
		Msg("Batch update completed")

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total_skus":    result.TotalSKUs,
		"unique_items":  result.UniqueItems,
		"processed":     result.Processed,
		"failed":        result.Failed,
		"skipped":       result.Skipped,
		"success":       result.Success,
		"results":       result.Results,
		"settings_used": settings,
	}))
}

// BatchDeleteByItemIds handles POST /api/wholesale/shopee/batch-delete
func (h *WholesaleBatchHandler) BatchDeleteByItemIds(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req BatchDeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("item_count", len(req.ItemIDs)).
		Msg("Batch delete wholesale by itemIds")

	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Get Shopee API client
	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Shopee API configuration failed"))
		return
	}

	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	service := wholesale.NewShopeeWholesaleService(db, tenantID, shopeeAPI)

	// Delete each item
	results := make([]wholesale.SingleWholesaleResult, 0)
	processed := 0
	failed := 0

	for _, itemID := range req.ItemIDs {
		err := service.DeleteWholesaleTiers(c.Request.Context(), itemID)
		if err != nil {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{
				ItemID:  itemID,
				Success: false,
				Error:   err.Error(),
			})
		} else {
			processed++
			results = append(results, wholesale.SingleWholesaleResult{
				ItemID:  itemID,
				Success: true,
				Message: "Wholesale deleted successfully",
			})
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":     len(req.ItemIDs),
		"processed": processed,
		"failed":    failed,
		"success":   failed == 0,
		"results":   results,
	}))
}

// BatchDeleteBySkus handles POST /api/wholesale/shopee/batch-delete-skus
// Body: { skus: string[] }
func (h *WholesaleBatchHandler) BatchDeleteBySkus(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req struct {
		SKUs []string `json:"skus" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if len(req.SKUs) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("skus array is required"))
		return
	}

	log.Info().
		Str("tenant_id", tenantID).
		Int("sku_count", len(req.SKUs)).
		Msg("Batch delete wholesale by SKUs")

	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Shopee API configuration failed"))
		return
	}

	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	service := wholesale.NewShopeeWholesaleService(db, tenantID, shopeeAPI)

	result, err := service.BatchDeleteBySkus(c.Request.Context(), req.SKUs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total_skus":   result.TotalSKUs,
		"unique_items": result.UniqueItems,
		"processed":    result.Processed,
		"failed":       result.Failed,
		"skipped":      result.Skipped,
		"success":      result.Success,
		"results":      result.Results,
		"message":      fmt.Sprintf("Deleted wholesale for %d/%d items", result.Processed, result.UniqueItems),
	}))
}
