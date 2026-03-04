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
)

// =============================================================================
// MPQ Batch Handlers (Shopee + TikTok)
// =============================================================================

// BatchSetMpq handles POST /api/wholesale/shopee/batch-mpq
// Body: { items: [{ sku, price }], mpq: number }
func (h *WholesaleExtendedHandler) BatchSetMpq(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
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
// TODO: Implement real TikTok API call (requires TikTok API client integration)
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

	log.Warn().
		Str("tenant_id", tenantID).
		Int("product_count", len(req.Products)).
		Msg("TikTok batch MPQ not implemented - requires TikTok API client integration")

	c.JSON(http.StatusNotImplemented, response.Error("TikTok batch MPQ not yet implemented - requires TikTok API client"))
}
