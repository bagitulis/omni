package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	inventoryService "github.com/omni/backend/internal/services/inventory"
)

// DriftedSKU represents a SKU with price drift
type DriftedSKU struct {
	SKU                 string    `json:"sku"`
	Platform            string    `json:"platform"`
	InventoryPrice      float64   `json:"inventory_price"`
	LastSyncedPrice     float64   `json:"last_synced_price"`
	LastSyncAt          time.Time `json:"last_sync_at"`
	InventoryUpdatedAt  time.Time `json:"inventory_updated_at"`
}

// PriceDriftResponse represents the response structure
type PriceDriftResponse struct {
	DriftedSKUs  []DriftedSKU `json:"drifted_skus"`
	TotalDrifted int          `json:"total_drifted"`
	TotalChecked int          `json:"total_checked"`
}

// PriceDrift handles GET /api/inventory/price-drift
// Detects SKUs where inventory price has changed since last marketplace sync
func (h *InventoryHandler) PriceDrift(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Query: Find inventory records with their latest successful price_update sync
	// Using raw SQL for efficiency (LEFT JOIN LATERAL for per-SKU latest sync)
	type QueryResult struct {
		SKU                string     `gorm:"column:sku"`
		InventoryUpdatedAt time.Time  `gorm:"column:inventory_updated_at"`
		InventoryData      string     `gorm:"column:inventory_data"`
		LastSyncAt         *time.Time `gorm:"column:last_sync_at"`
		LastSyncPlatform   *string    `gorm:"column:last_sync_platform"`
		LastSyncRequest    *string    `gorm:"column:last_sync_request"`
	}

	var results []QueryResult
	query := `
		SELECT 
			ir.key_value as sku,
			ir.updated_at as inventory_updated_at,
			ir.data as inventory_data,
			msh.created_at as last_sync_at,
			msh.platform as last_sync_platform,
			msh.request_data as last_sync_request
		FROM inventory_records ir
		LEFT JOIN LATERAL (
			SELECT created_at, platform, request_data
			FROM marketplace_sync_histories
			WHERE sku = ir.key_value 
				AND tenant_id = ir.tenant_id
				AND operation = 'price_update' 
				AND status = 'success'
			ORDER BY created_at DESC 
			LIMIT 1
		) msh ON true
		WHERE ir.tenant_id = ?
			AND (msh.created_at IS NULL OR ir.updated_at > msh.created_at)
		LIMIT 100
	`

	if err := db.WithContext(c.Request.Context()).Raw(query, tenantID).Scan(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Process results to extract drifted SKUs
	driftedSKUs := []DriftedSKU{}
	totalChecked := len(results)

	for _, result := range results {
		// Parse inventory data to get current prices
		var inventoryRecord models.InventoryRecord
		inventoryRecord.Data = result.InventoryData
		currentPrices := inventoryService.GetPricePerPlatform(inventoryRecord)

		// If no sync exists, report drift for all platforms with non-zero prices
		if result.LastSyncAt == nil {
			for _, platform := range []string{"shopee", "tiktok", "lazada"} {
				if currentPrices[platform] > 0 {
					driftedSKUs = append(driftedSKUs, DriftedSKU{
						SKU:                result.SKU,
						Platform:           platform,
						InventoryPrice:     currentPrices[platform],
						LastSyncedPrice:    0,
						LastSyncAt:         time.Time{}, // Zero time
						InventoryUpdatedAt: result.InventoryUpdatedAt,
					})
				}
			}
			continue
		}

		// Parse last synced price from request_data
		if result.LastSyncRequest != nil && result.LastSyncPlatform != nil {
			var requestData map[string]any
			if err := json.Unmarshal([]byte(*result.LastSyncRequest), &requestData); err == nil {
				lastSyncedPrice := 0.0
				if priceVal, ok := requestData["price"]; ok {
					switch v := priceVal.(type) {
					case float64:
						lastSyncedPrice = v
					case int:
						lastSyncedPrice = float64(v)
					}
				}

				platform := *result.LastSyncPlatform
				currentPrice := currentPrices[platform]

				// Only report drift if prices differ
				if currentPrice != lastSyncedPrice {
					driftedSKUs = append(driftedSKUs, DriftedSKU{
						SKU:                result.SKU,
						Platform:           platform,
						InventoryPrice:     currentPrice,
						LastSyncedPrice:    lastSyncedPrice,
						LastSyncAt:         *result.LastSyncAt,
						InventoryUpdatedAt: result.InventoryUpdatedAt,
					})
				}
			}
		}
	}

	response := PriceDriftResponse{
		DriftedSKUs:  driftedSKUs,
		TotalDrifted: len(driftedSKUs),
		TotalChecked: totalChecked,
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": response})
}
