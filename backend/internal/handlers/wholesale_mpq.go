package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/services/wholesale"
	"github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// =============================================================================
// MPQ Batch Handlers (Shopee + TikTok)
// =============================================================================

// BatchSetMpq handles POST /api/wholesale/shopee/batch-mpq
// Body: { items: [{ sku, price }], mpq: number }
func (h *WholesaleExtendedHandler) BatchSetMpq(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

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

	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		log.Error().Str("tenant_id", tenantID).Err(err).Msg("Failed to get Shopee client")
		c.JSON(http.StatusInternalServerError, response.Error(fmt.Sprintf("Shopee API configuration failed: %v", err)))
		return
	}

	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	mpqService := wholesale.NewShopeeMpqService(db, tenantID, shopeeAPI)

	skuPriceMap := make(map[string]float64)
	for _, item := range reqBody.Items {
		if item.SKU != "" && item.Price > 0 {
			skuPriceMap[item.SKU] = item.Price
		}
	}

	result, err := mpqService.BatchSetMpqBySkus(c.Request.Context(), skuPriceMap, reqBody.MPQ)
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
		"results":      result.Results,
		"mpq":          reqBody.MPQ,
		"success":      result.Success,
		"message":      fmt.Sprintf("MPQ=%d: %d/%d items", reqBody.MPQ, result.Processed, result.UniqueItems),
	}))
}

// BatchSetTiktokMpq handles POST /api/wholesale/tiktok/batch-mpq
func (h *WholesaleExtendedHandler) BatchSetTiktokMpq(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req TiktokBatchMpqRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	db, err := h.getDB(c, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Get TikTok client from platform coordination service
	coordService := platform.GetPlatformCoordinationService(tenantID)
	if err := coordService.InitializePlatforms(c.Request.Context()); err != nil {
		log.Error().Str("tenant_id", tenantID).Err(err).Msg("Failed to initialize platform clients")
		c.JSON(http.StatusInternalServerError, response.Error("Failed to initialize platform clients"))
		return
	}
	tiktokClient := coordService.GetTiktokClient()
	if tiktokClient == nil || !tiktokClient.IsInitialized() {
		c.JSON(http.StatusServiceUnavailable, response.Error("TikTok API not configured for this tenant"))
		return
	}

	mpqService := wholesale.NewTiktokMpqService(db, tenantID, tiktokClient)

	// Build SKU→price map from both items and products
	skuPriceMap := make(map[string]float64)
	for _, item := range req.Items {
		if item.SKU != "" {
			skuPriceMap[item.SKU] = item.Price
		}
	}
	for _, p := range req.Products {
		if p.SKU != "" && p.Price > 0 {
			skuPriceMap[p.SKU] = p.Price
		}
	}

	result, err := mpqService.BatchUpdateMpq(c.Request.Context(), skuPriceMap, req.MPQ)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total_skus":      result.TotalSKUs,
		"unique_products": result.UniqueProducts,
		"processed":       result.Processed,
		"failed":          result.Failed,
		"skipped":         result.Skipped,
		"results":         result.Results,
		"mpq":             req.MPQ,
		"success":         result.Success,
	}))
}
