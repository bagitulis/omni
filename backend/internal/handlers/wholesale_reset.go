package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/wholesale"
	"github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type wholesaleResetItemData struct {
	skus    []string
	price   float64
	modelID *int64
}

// BatchWholesaleReset handles POST /api/wholesale/shopee/batch-reset
// Also used for /api/wholesale/shopee/batch-wholesale-reset
// Body: { items: [{ sku, price }] }
func (h *WholesaleExtendedHandler) BatchWholesaleReset(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	var req BatchWholesaleResetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, response.Error("items array is required: [{ sku, price }]"))
		return
	}

	ctx := c.Request.Context()

	log.Info().
		Str("tenant_id", tenantID).
		Int("item_count", len(req.Items)).
		Msg("Batch wholesale reset")

	db, err := h.getDB(c, tenantID)
	if err != nil {
		log.Error().Err(err).Msg("Database connection failed")
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	shopeeClient, err := config.GetShopeeClient(tenantID, h.basePath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get Shopee client")
		c.JSON(http.StatusInternalServerError, response.Error("Shopee API configuration failed"))
		return
	}

	shopeeAPI := shopee.NewProductAPI(shopeeClient)
	settingsService := wholesale.NewWholesaleService(db, tenantID)
	settings, err := settingsService.GetSettings(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get wholesale settings")
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get wholesale settings"))
		return
	}

	mpqService := wholesale.NewShopeeMpqService(db, tenantID, shopeeAPI)
	wholesaleService := wholesale.NewShopeeWholesaleService(db, tenantID, shopeeAPI)

	itemMap := make(map[int64]*wholesaleResetItemData)
	skipped := make([]string, 0)
	results := make([]wholesale.SingleWholesaleResult, 0)
	failed := 0

	for _, item := range req.Items {
		sku := strings.TrimSpace(item.SKU)
		if sku == "" {
			skipped = append(skipped, sku)
			continue
		}

		var skuModel models.ShopeeSku
		err := db.WithContext(ctx).
			Where("tenant_id = ? AND seller_sku = ?", tenantID, sku).
			First(&skuModel).Error

		if err == gorm.ErrRecordNotFound {
			skipped = append(skipped, sku)
			continue
		}
		if err != nil {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{
				SKU:     sku,
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		if skuModel.ItemID == 0 {
			skipped = append(skipped, sku)
			continue
		}

		if existing, exists := itemMap[skuModel.ItemID]; exists {
			existing.skus = append(existing.skus, sku)
			if existing.price <= 0 && item.Price > 0 {
				existing.price = item.Price
			}
			if existing.modelID == nil {
				existing.modelID = skuModel.ModelID
			}
			continue
		}

		itemMap[skuModel.ItemID] = &wholesaleResetItemData{
			skus:    []string{sku},
			price:   item.Price,
			modelID: skuModel.ModelID,
		}
	}

	uniqueItems := len(itemMap)
	processed := 0

	for itemID, data := range itemMap {
		if data.price <= 0 {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{
				ItemID:  itemID,
				SKU:     data.skus[0],
				Success: false,
				Error:   "price must be > 0 for tier calculation",
			})
			continue
		}

		if err := mpqService.SetMpq(ctx, itemID, 1); err != nil {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{
				ItemID:  itemID,
				SKU:     data.skus[0],
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		tiers := settingsService.CalculateTiersFromSettings(data.price, settings)
		if err := wholesaleService.UpdateWholesaleTiers(ctx, itemID, tiers); err != nil {
			failed++
			results = append(results, wholesale.SingleWholesaleResult{
				ItemID:  itemID,
				SKU:     data.skus[0],
				Success: false,
				Error:   err.Error(),
			})
			continue
		}

		processed++
		results = append(results, wholesale.SingleWholesaleResult{
			ItemID:  itemID,
			SKU:     data.skus[0],
			Success: true,
			Message: fmt.Sprintf("Reset MPQ to 1 and applied %d tiers for %d SKUs", len(tiers), len(data.skus)),
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total_skus":    len(req.Items),
		"unique_items":  uniqueItems,
		"processed":     processed,
		"failed":        failed,
		"skipped":       skipped,
		"results":       results,
		"settings_used": settings,
		"success":       failed == 0,
		"message":       fmt.Sprintf("Wholesale reset: %d/%d items", processed, uniqueItems),
	}))
}
