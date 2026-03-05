package master_product

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
)

// enrichWithInventoryPrices backfills price from inventory_records (Google Sheets)
// and sets InventoryPrice/InventoryStock virtual fields for reference display.
// NOTE: Stock is NOT backfilled to avoid overwriting marketplace sync results.
func (s *Service) enrichWithInventoryPrices(ctx context.Context, tenantID string, products []models.MasterProduct) {
	allSKUs := collectSKUs(products)
	if len(allSKUs) == 0 {
		return
	}

	// Batch query inventory_records
	inventoryTable := models.GetTableName("InventoryRecord")
	var inventoryRows []struct {
		KeyValue string `gorm:"column:key_value"`
		Data     string `gorm:"column:data"`
	}
	if err := s.db.WithContext(ctx).
		Table(inventoryTable).
		Select("key_value, data").
		Where("tenant_id = ? AND LOWER(key_value) IN (?)", tenantID, allSKUs).
		Find(&inventoryRows).Error; err != nil {
		log.Warn().Err(err).Msg("Failed to query inventory for price enrichment")
		return
	}

	// Build lookup map: lowercase SKU → inventory data
	invMap := make(map[string]map[string]interface{}, len(inventoryRows))
	for _, row := range inventoryRows {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(row.Data), &data); err != nil {
			continue
		}
		invMap[strings.ToLower(row.KeyValue)] = data
	}

	// Track SKUs that need price-only DB backfill
	type backfillEntry struct {
		ID    uint
		Price float64
	}
	var backfills []backfillEntry

	for i := range products {
		for j := range products[i].SKUs {
			sku := &products[i].SKUs[j]
			invData, ok := invMap[strings.ToLower(sku.SellerSku)]
			if !ok {
				continue
			}

			needsUpdate := false

			// Set inventory reference price (virtual field)
			if hargaStr, ok := invData["HARGA"].(string); ok {
				if harga, err := strconv.ParseFloat(strings.TrimSpace(hargaStr), 64); err == nil && harga > 0 {
					sku.InventoryPrice = harga
					// Backfill master price if 0
					if sku.Price == 0 {
						sku.Price = harga
						needsUpdate = true
					}
				}
			}

			// Set inventory reference stock (virtual field only — no DB write)
			if stokStr, ok := invData["Sisa Stok"].(string); ok {
				if stok, err := strconv.Atoi(strings.TrimSpace(stokStr)); err == nil && stok > 0 {
					sku.InventoryStock = stok
				}
			}

			if needsUpdate {
				backfills = append(backfills, backfillEntry{ID: sku.ID, Price: sku.Price})
			}
		}
	}

	// Persist backfilled price to DB (stock is NOT backfilled to avoid
	// overwriting legitimate zero-stock from marketplace sync)
	for _, bf := range backfills {
		s.db.WithContext(ctx).
			Model(&models.MasterProductSku{}).
			Where("id = ?", bf.ID).
			Updates(map[string]interface{}{"price": bf.Price})
	}
}

// enrichWithPlatformPrices populates PlatformPrices virtual field from staging tables.
// Queries shopee_skus, lazada_skus, tiktok_skus for real marketplace prices.
func (s *Service) enrichWithPlatformPrices(ctx context.Context, tenantID string, products []models.MasterProduct) {
	allSKUs := collectSKUs(products)
	if len(allSKUs) == 0 {
		return
	}

	type stagingRow struct {
		SellerSku string  `gorm:"column:seller_sku"`
		Price     float64 `gorm:"column:price"`
		Quantity  int     `gorm:"column:quantity"`
	}

	// platformMap: lowercase_sku → platform → {price, stock}
	platformMap := make(map[string][]models.PlatformPrice)

	// Query each platform staging table
	platforms := []struct {
		name  string
		table string
	}{
		{"shopee", models.GetTableName("ShopeeSku")},
		{"lazada", models.GetTableName("LazadaSku")},
		{"tiktok", models.GetTableName("TiktokSku")},
	}

	for _, p := range platforms {
		var rows []stagingRow
		if err := s.db.WithContext(ctx).
			Table(p.table).
			Select("seller_sku, price, quantity").
			Where("tenant_id = ? AND LOWER(seller_sku) IN (?)", tenantID, allSKUs).
			Find(&rows).Error; err != nil {
			log.Warn().Err(err).Str("platform", p.name).Msg("Failed to query platform staging prices")
			continue
		}

		for _, row := range rows {
			key := strings.ToLower(row.SellerSku)
			platformMap[key] = append(platformMap[key], models.PlatformPrice{
				Platform: p.name,
				Price:    row.Price,
				Stock:    row.Quantity,
			})
		}
	}

	// Attach platform prices to each SKU
	for i := range products {
		for j := range products[i].SKUs {
			sku := &products[i].SKUs[j]
			if prices, ok := platformMap[strings.ToLower(sku.SellerSku)]; ok {
				sku.PlatformPrices = prices
			}
		}
	}
}

// collectSKUs gathers all lowercase seller_skus from products.
func collectSKUs(products []models.MasterProduct) []string {
	var skus []string
	for _, p := range products {
		for _, sku := range p.SKUs {
			if sku.SellerSku != "" {
				skus = append(skus, strings.ToLower(sku.SellerSku))
			}
		}
	}
	return skus
}
