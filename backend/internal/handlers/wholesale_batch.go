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
)

// =============================================================================
// Shopee Batch Wholesale Operations (Delete, Add, Preview, Import)
// =============================================================================

// BatchDeleteByItemIds handles POST /api/wholesale/shopee/batch-delete
func (h *WholesaleExtendedHandler) BatchDeleteByItemIds(c *gin.Context) {
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
	if len(req.ItemIDs) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("item_ids array is required"))
		return
	}

	service, err := h.getShopeeWholesaleService(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	results := make([]wholesale.SingleWholesaleResult, 0, len(req.ItemIDs))
	processed, failed := 0, 0

	for _, itemID := range req.ItemIDs {
		if err := service.DeleteWholesaleTiers(c.Request.Context(), itemID); err != nil {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{ItemID: itemID, Success: false, Error: err.Error()})
		} else {
			processed++
			results = append(results, wholesale.SingleWholesaleResult{ItemID: itemID, Success: true, Message: "Wholesale deleted"})
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total": len(req.ItemIDs), "processed": processed, "failed": failed,
		"success": failed == 0, "results": results,
	}))
}

// BatchAdd handles POST /api/wholesale/shopee/batch-add
func (h *WholesaleExtendedHandler) BatchAdd(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req BatchAddRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}
	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("items array is required"))
		return
	}

	service, err := h.getShopeeWholesaleService(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	results := make([]wholesale.SingleWholesaleResult, 0, len(req.Items))
	processed, failed := 0, 0

	for _, item := range req.Items {
		tiers := convertDTOTiersToService(item.Tiers)
		if err := service.UpdateWholesaleTiers(c.Request.Context(), item.ItemID, tiers); err != nil {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{ItemID: item.ItemID, Success: false, Error: err.Error()})
		} else {
			processed++
			results = append(results, wholesale.SingleWholesaleResult{
				ItemID: item.ItemID, Success: true, Message: fmt.Sprintf("Added %d tier(s)", len(item.Tiers)),
			})
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total": len(req.Items), "processed": processed, "failed": failed,
		"success": failed == 0, "results": results,
	}))
}

// Preview handles POST /api/wholesale/shopee/preview
func (h *WholesaleExtendedHandler) Preview(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req PreviewRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	settings, err := service.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get settings"))
		return
	}

	var results []gin.H
	for _, sku := range req.SKUs {
		basePrice := float64(100000)
		if req.BasePrice > 0 {
			basePrice = req.BasePrice
		}
		tiers := service.CalculateTiersFromSettings(basePrice, settings)
		results = append(results, gin.H{"sku": sku, "base_price": basePrice, "tiers": tiers})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"results": results, "settings_used": settings}))
}

// ImportWholesale handles POST /api/wholesale/shopee/import
func (h *WholesaleExtendedHandler) ImportWholesale(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req ImportWholesaleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}
	if len(req.Data) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("data array is required"))
		return
	}

	service, err := h.getShopeeWholesaleService(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	results := make([]wholesale.SingleWholesaleResult, 0, len(req.Data))
	processed, failed := 0, 0

	for _, item := range req.Data {
		itemID, lookupErr := service.LookupItemIDBySKU(c.Request.Context(), item.SKU)
		if lookupErr != nil {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{SKU: item.SKU, Success: false, Error: "SKU not found: " + item.SKU})
			continue
		}

		tiers := convertDTOTiersToService(item.Tiers)
		if err := service.UpdateWholesaleTiers(c.Request.Context(), itemID, tiers); err != nil {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{ItemID: itemID, SKU: item.SKU, Success: false, Error: err.Error()})
		} else {
			processed++
			results = append(results, wholesale.SingleWholesaleResult{
				ItemID: itemID, SKU: item.SKU, Success: true, Message: fmt.Sprintf("Imported %d tier(s)", len(item.Tiers)),
			})
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total": len(req.Data), "processed": processed, "failed": failed,
		"success": failed == 0, "results": results,
	}))
}

// =============================================================================
// Helpers (DRY)
// =============================================================================

// getShopeeWholesaleService creates a ShopeeWholesaleService from context
func (h *WholesaleExtendedHandler) getShopeeWholesaleService(c *gin.Context, tenantID string) (*wholesale.ShopeeWholesaleService, error) {
	db, err := h.getDB(c, tenantID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed")
	}
	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		return nil, fmt.Errorf("shopee API configuration failed")
	}
	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	return wholesale.NewShopeeWholesaleService(db, tenantID, shopeeAPI), nil
}

// convertDTOTiersToService converts handler DTOs to service tiers
func convertDTOTiersToService(dtoTiers []WholesaleTier) []wholesale.WholesaleTier {
	tiers := make([]wholesale.WholesaleTier, len(dtoTiers))
	for i, t := range dtoTiers {
		tiers[i] = wholesale.WholesaleTier{MinCount: t.MinCount, MaxCount: t.MaxCount, UnitPrice: t.UnitPrice}
	}
	return tiers
}
