package shopee

import (
	"context"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// SyncProductsWithDetails fetches products from Shopee API, saves to DB, and returns them.
// Used by GET /api/shopee/products to sync and return products.
func (s *ProductSyncService) SyncProductsWithDetails(ctx context.Context, itemStatus string, offset, limit int) ([]map[string]interface{}, int, error) {
	_ = offset
	_ = limit

	zlog := zerolog.Ctx(ctx)
	zlog.Debug().Str("item_status", itemStatus).Msg("Starting product sync")

	allItemIDs, err := s.fetchAllItemIDs(ctx)
	if err != nil {
		zlog.Error().Err(err).Msg("GetProductList error")
		return nil, 0, err
	}

	zlog.Debug().Int("total_items", len(allItemIDs)).Msg("Total item IDs collected")
	if len(allItemIDs) == 0 {
		zlog.Warn().Msg("Shopee API returned 0 products - skipping sync to prevent data loss")
		return []map[string]interface{}{}, 0, nil
	}

	products := make([]map[string]interface{}, 0)
	savedCount := 0

	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := s.clearShopeeProductCacheWithDB(ctx, tx); err != nil {
			return err
		}

		txProdRepo := repositories.NewShopeeProductRepository(tx)
		txSkuRepo := repositories.NewShopeeSkuRepository(tx)
		batchSize := 50

		for i := 0; i < len(allItemIDs); i += batchSize {
			end := i + batchSize
			if end > len(allItemIDs) {
				end = len(allItemIDs)
			}

			batchIDs := allItemIDs[i:end]
			detailResp, err := s.client.GetProductDetail(batchIDs)
			if err != nil {
				zlog.Error().Err(err).Msg("GetProductDetail batch error")
				return err
			}

			for _, prod := range detailResp.Response.ItemList {
				if itemStatus != "" && prod.ItemStatus != itemStatus {
					continue
				}

				dbProd := &models.ShopeeProduct{
					TenantID:    s.tenantID,
					ItemID:      prod.ItemID,
					Name:        prod.ItemName,
					Description: prod.Description,
					Status:      prod.ItemStatus,
					Price:       prod.CurrentPrice,
					Quantity:    prod.Stock,
				}

				if err := txProdRepo.Upsert(ctx, dbProd); err != nil {
					zlog.Warn().Err(err).Int64("item_id", prod.ItemID).Msg("Failed to upsert product")
					continue
				}

				savedCount++
				savedProd, err := txProdRepo.FindByItemID(ctx, prod.ItemID)
				if err == nil {
					if err := s.syncProductSKUs(ctx, txSkuRepo, savedProd, prod.ItemID, prod.ItemSKU); err != nil {
						zlog.Warn().Err(err).Int64("item_id", prod.ItemID).Msg("Failed to sync product skus")
					}

					if len(prod.Images) > 0 {
						localPaths := s.downloadAndSaveProductImages(ctx, prod.ItemID, prod.Images)
						if err := s.updateProductLocalImages(ctx, tx, savedProd.ID, localPaths); err != nil {
							zlog.Warn().Err(err).Uint("product_id", savedProd.ID).Msg("Failed to update local_images")
						}
					}
				}

				productItem := map[string]interface{}{
					"item_id":        prod.ItemID,
					"item_name":      prod.ItemName,
					"item_status":    prod.ItemStatus,
					"description":    prod.Description,
					"category_id":    prod.CategoryID,
					"original_price": prod.OriginalPrice,
					"current_price":  prod.CurrentPrice,
					"stock":          prod.Stock,
				}

				if len(prod.Images) > 0 {
					productItem["image"] = prod.Images[0]
				}

				products = append(products, productItem)
			}
		}

		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	zlog.Info().Int("total_products", len(products)).Int("saved_count", savedCount).Msg("Product sync completed")
	return products, savedCount, nil
}
