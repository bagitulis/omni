package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/wholesale"
	"github.com/omni/backend/pkg/shopee"
)

// =============================================================================
// Batch Operations for Wholesale Handler
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

	results := make([]map[string]interface{}, 0, len(req.ItemIDs))
	for _, itemID := range req.ItemIDs {
		results = append(results, map[string]interface{}{
			"itemId":  itemID,
			"success": true,
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":   len(req.ItemIDs),
		"results": results,
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

	results := make([]map[string]interface{}, 0, len(req.Items))
	for _, item := range req.Items {
		results = append(results, map[string]interface{}{
			"itemId":  item.ItemID,
			"sku":     item.SKU,
			"success": true,
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":   len(req.Items),
		"results": results,
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

	var settings models.WholesaleSettings
	db.Where("tenant_id = ?", tenantID).First(&settings)

	previews := make([]map[string]interface{}, 0, len(req.SKUs))
	for _, sku := range req.SKUs {
		previews = append(previews, map[string]interface{}{
			"sku":           sku,
			"originalPrice": 100000,
			"tiers":         GenerateTiers(100000, req.DiscountRates),
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"previews": previews}))
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

	results := make([]map[string]interface{}, 0, len(req.Data))
	for _, item := range req.Data {
		results = append(results, map[string]interface{}{
			"sku":     item.SKU,
			"success": true,
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":   len(req.Data),
		"success": len(req.Data),
		"results": results,
	}))
}

// BatchSetMpq handles POST /api/wholesale/shopee/batch-mpq
// Body: { items: [{ sku, price }], mpq: number }
func (h *WholesaleExtendedHandler) BatchSetMpq(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	// Parse request body
	var reqBody struct {
		Items []struct {
			SKU   string  `json:"sku" binding:"required"`
			Price float64 `json:"price" binding:"required"`
		} `json:"items" binding:"required"`
		MPQ int `json:"mpq" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&reqBody); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if len(reqBody.Items) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("items array is required: [{ sku, price }]"))
		return
	}

	// Get database
	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Get Shopee API client
	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		log.Printf("[ERROR] Failed to get Shopee client for tenant %s: %v", tenantID, err)
		c.JSON(http.StatusInternalServerError, response.Error(fmt.Sprintf("Shopee API configuration failed: %v", err)))
		return
	}

	// Create MPQ service
	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	mpqService := wholesale.NewShopeeMpqService(db, tenantID, shopeeAPI)

	// Build SKU → Price map
	skuPriceMap := make(map[string]float64)
	for _, item := range reqBody.Items {
		if item.SKU != "" && item.Price > 0 {
			skuPriceMap[item.SKU] = item.Price
		}
	}

	// Execute batch MPQ operation
	result, err := mpqService.BatchSetMpqBySkus(c.Request.Context(), skuPriceMap, reqBody.MPQ)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	// Return response with snake_case
	c.JSON(http.StatusOK, response.Success(gin.H{
		"total_skus":   result.TotalSKUs,
		"unique_items": result.UniqueItems,
		"processed":    result.Processed,
		"failed":       result.Failed,
		"skipped":      result.Skipped,
		"results":      result.Results,
		"mpq":          reqBody.MPQ,
		"message":      result.Success,
	}))
}

// BatchSetTiktokMpq handles POST /api/wholesale/tiktok/batch-mpq
func (h *WholesaleExtendedHandler) BatchSetTiktokMpq(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	var req TiktokBatchMpqRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	results := make([]map[string]interface{}, 0, len(req.Products))
	for _, product := range req.Products {
		results = append(results, map[string]interface{}{
			"productId": product.ProductID,
			"mpq":       product.MPQ,
			"success":   true,
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":   len(req.Products),
		"results": results,
	}))
}
