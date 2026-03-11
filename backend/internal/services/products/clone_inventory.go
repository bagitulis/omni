// Package products provides inventory data lookup for cloning
package products

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"strconv"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// InventoryData holds price and stock from inventory_records
type InventoryData struct {
	Found   bool                   `json:"found"`
	SKU     string                 `json:"sku"`
	Price   float64                `json:"price"`
	Stock   int                    `json:"stock"`
	RawData map[string]interface{} `json:"raw_data,omitempty"`
}

// InventoryFetcher fetches inventory data for cloning
type InventoryFetcher struct {
	db       *gorm.DB
	tenantID string
}

// NewInventoryFetcher creates a new inventory fetcher
func NewInventoryFetcher(db *gorm.DB, tenantID string) *InventoryFetcher {
	return &InventoryFetcher{db: db, tenantID: tenantID}
}

// GetInventoryBySKU fetches inventory data by SKU from inventory_records
func (f *InventoryFetcher) GetInventoryBySKU(ctx context.Context, sku string) (*InventoryData, error) {
	var record models.InventoryRecord
	err := f.db.WithContext(ctx).
		Where("tenant_id = ? AND key_value = ?", f.tenantID, sku).
		First(&record).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Info().Msgf("[InventoryFetcher] SKU %s not found in inventory_records", sku)
			return &InventoryData{Found: false, SKU: sku}, nil
		}
		return nil, err
	}

	// Parse JSONB data
	data := f.parseData(record.Data)
	price := f.getPrice(data)
	stock := f.getStock(data)

	log.Info().Msgf("[InventoryFetcher] Found inventory for SKU %s: price=%.2f, stock=%d", sku, price, stock)

	return &InventoryData{
		Found:   true,
		SKU:     sku,
		Price:   price,
		Stock:   stock,
		RawData: data,
	}, nil
}

// GetInventoryBySellerSKU tries to find inventory by searching various SKU column names
func (f *InventoryFetcher) GetInventoryBySellerSKU(ctx context.Context, sellerSku string) (*InventoryData, error) {
	// First try direct lookup by key_value
	result, err := f.GetInventoryBySKU(ctx, sellerSku)
	if err != nil {
		return nil, err
	}
	if result.Found {
		return result, nil
	}

	// If not found, try searching in JSONB data (slower but catches variations)
	var record models.InventoryRecord
	skuColumns := []string{"SKU", "sku", "Sku", "KODE", "kode", "Code", "code", "SellerSku", "seller_sku"}

	for _, col := range skuColumns {
		// PostgreSQL JSONB query: data->>col = 'value'
		err := f.db.WithContext(ctx).
			Where("tenant_id = ? AND data->>? = ?", f.tenantID, col, sellerSku).
			First(&record).Error

		if err == nil {
			data := f.parseData(record.Data)
			return &InventoryData{
				Found:   true,
				SKU:     sellerSku,
				Price:   f.getPrice(data),
				Stock:   f.getStock(data),
				RawData: data,
			}, nil
		}
	}

	return &InventoryData{Found: false, SKU: sellerSku}, nil
}

// parseData parses JSONB data string to map
func (f *InventoryFetcher) parseData(dataStr string) map[string]interface{} {
	data := make(map[string]interface{})
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		log.Info().Msgf("[InventoryFetcher] Failed to parse JSONB data: %v", err)
	}
	return data
}

// getPrice extracts price from data using common column names
func (f *InventoryFetcher) getPrice(data map[string]interface{}) float64 {
	priceColumns := []string{"HARGA", "Harga", "harga", "Price", "price", "PRICE", "HargaJual", "harga_jual"}

	for _, col := range priceColumns {
		if val, ok := data[col]; ok {
			switch v := val.(type) {
			case float64:
				return v
			case int:
				return float64(v)
			case string:
				f, _ := strconv.ParseFloat(v, 64)
				return f
			}
		}
	}
	return 0
}

// getStock extracts stock from data using common column names
func (f *InventoryFetcher) getStock(data map[string]interface{}) int {
	stockColumns := []string{"Stock", "stock", "STOCK", "Quantity", "quantity", "QTY", "qty", "Stok", "stok", "Jumlah", "jumlah"}

	for _, col := range stockColumns {
		if val, ok := data[col]; ok {
			switch v := val.(type) {
			case float64:
				return int(v)
			case int:
				return v
			case string:
				i, _ := strconv.Atoi(v)
				return i
			}
		}
	}
	return 0
}

// GetProductNameFromInventory gets product name from inventory data
func (f *InventoryFetcher) GetProductNameFromInventory(data map[string]interface{}) string {
	nameColumns := []string{"Nama Barang", "NamaBarang", "nama_barang", "Nama", "nama", "Name", "name", "ProductName", "product_name", "Title", "title"}

	for _, col := range nameColumns {
		if val, ok := data[col]; ok {
			if str, ok := val.(string); ok && str != "" {
				return str
			}
		}
	}
	return ""
}
