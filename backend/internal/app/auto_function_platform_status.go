package app

import (
	"context"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// refreshInventoryPlatformStatus populates InventorySkuPlatformStatus from staging tables.
// This ensures the inventory page shows real platform stock/price data.
func refreshInventoryPlatformStatus(ctx context.Context, db *gorm.DB, tenantID string) (int, error) {
	now := time.Now()
	updated := 0

	type stagingSku struct {
		SellerSku string
		Price     float64
		Quantity  int
		ItemID    string
		SkuID     string
	}

	// Helper to upsert platform status records
	upsertStatus := func(platform string, skus []stagingSku) {
		for _, sku := range skus {
			if sku.SellerSku == "" {
				continue
			}
			status := models.InventorySkuPlatformStatus{
				TenantID:      tenantID,
				SKU:           sku.SellerSku,
				Platform:      platform,
				PlatformItemID: sku.ItemID,
				PlatformSKU:   sku.SkuID,
				Status:        "active",
				Stock:         sku.Quantity,
				Price:         sku.Price,
				LastCheckedAt: now,
			}

			result := db.WithContext(ctx).
				Where("tenant_id = ? AND sku = ? AND platform = ?", tenantID, sku.SellerSku, platform).
				Assign(map[string]interface{}{
					"stock":            sku.Quantity,
					"price":            sku.Price,
					"platform_item_id": sku.ItemID,
					"platform_sku":     sku.SkuID,
					"status":           "active",
					"last_checked_at":  now,
					"updated_at":       now,
				}).
				FirstOrCreate(&status)
			if result.Error == nil {
				updated++
			}
		}
	}

	// Shopee SKUs
	var shopeeSkus []stagingSku
	db.WithContext(ctx).Table("shopee_skus").
		Where("tenant_id = ? AND seller_sku != ''", tenantID).
		Select("seller_sku, price, quantity, CAST(item_id AS TEXT) as item_id, CAST(sku_id AS TEXT) as sku_id").
		Limit(500).
		Scan(&shopeeSkus)
	upsertStatus("shopee", shopeeSkus)

	// TikTok SKUs
	var tiktokSkus []stagingSku
	db.WithContext(ctx).Table("tiktok_skus").
		Where("tenant_id = ? AND seller_sku != ''", tenantID).
		Select("seller_sku, price, quantity, CAST(product_id AS TEXT) as item_id, CAST(sku_id AS TEXT) as sku_id").
		Limit(500).
		Scan(&tiktokSkus)
	upsertStatus("tiktok", tiktokSkus)

	// Lazada SKUs (uses shop_sku as seller_sku)
	var lazadaSkus []stagingSku
	db.WithContext(ctx).Table("lazada_skus").
		Where("tenant_id = ? AND shop_sku != ''", tenantID).
		Select("shop_sku as seller_sku, price, quantity, CAST(item_id AS TEXT) as item_id, CAST(sku_id AS TEXT) as sku_id").
		Limit(500).
		Scan(&lazadaSkus)
	upsertStatus("lazada", lazadaSkus)

	log.Info().Msgf("[AutoFunction] Platform status refreshed for %s: %d SKUs updated", tenantID, updated)
	return updated, nil
}
