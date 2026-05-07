package handlers

import (
	"context"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	inventoryService "github.com/omni/backend/internal/services/inventory"
	"gorm.io/gorm"
)

// initCredentialService creates a CredentialService using DB_PATH env var.
func initCredentialService() *services.CredentialService {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data"
	}
	return services.NewCredentialService(dbPath)
}

// resolvePlatforms returns the effective platform list from plural/singular fields.
func resolvePlatforms(platforms []string, platform string) []string {
	if len(platforms) > 0 {
		return platforms
	}
	if platform != "" {
		return []string{platform}
	}
	return platforms
}

// resolvePriceFromInventory determines the price value from an inventory record
// considering per-platform pricing and request overrides.
func resolvePriceFromInventory(record models.InventoryRecord, reqPrice *float64, platforms []string, platform string) float64 {
	perPlatformPrices := inventoryService.GetPricePerPlatform(record)
	basePrice := perPlatformPrices["base"]

	var priceValue float64

	// For single-platform sync: use that platform's specific price
	// For multi-platform or no platform specified: use base price
	if len(platforms) == 1 {
		platformKey := platforms[0]
		if pp, ok := perPlatformPrices[platformKey]; ok && pp > 0 {
			priceValue = pp
		} else {
			priceValue = basePrice
		}
	} else if platform != "" {
		if pp, ok := perPlatformPrices[platform]; ok && pp > 0 {
			priceValue = pp
		} else {
			priceValue = basePrice
		}
	} else {
		priceValue = basePrice
	}

	// If inventory has price=0 but request provides a price, prefer request price.
	if priceValue == 0 && reqPrice != nil {
		priceValue = *reqPrice
	}

	return priceValue
}

// updatePriceMultiPlatform handles per-platform price updates when multiple platforms
// are specified and an inventory record exists.
func updatePriceMultiPlatform(
	ctx context.Context,
	orchestrator *inventoryService.PriceUpdateOrchestrator,
	record models.InventoryRecord,
	sku string,
	platforms []string,
	fallbackPrice float64,
) (allResults map[string]*inventoryService.PlatformPriceResult, lastErr error) {
	perPlatformPrices := inventoryService.GetPricePerPlatform(record)
	allResults = make(map[string]*inventoryService.PlatformPriceResult)

	for _, p := range platforms {
		pp := perPlatformPrices[p]
		if pp == 0 {
			pp = fallbackPrice
		}
		result, err := orchestrator.UpdatePrice(ctx, sku, pp, []string{p})
		if err != nil {
			lastErr = err
			continue
		}
		for k, v := range result.Platforms {
			allResults[k] = v
		}
	}
	return allResults, lastErr
}

// newStockOrchestrator creates a StockUpdateOrchestrator with standard credential setup.
func newStockOrchestrator(db *gorm.DB, tenantID string) *inventoryService.StockUpdateOrchestrator {
	credService := initCredentialService()
	return inventoryService.NewStockUpdateOrchestrator(db, tenantID, credService)
}

// newPriceOrchestrator creates a PriceUpdateOrchestrator with standard credential setup.
func newPriceOrchestrator(db *gorm.DB, tenantID string) *inventoryService.PriceUpdateOrchestrator {
	credService := initCredentialService()
	return inventoryService.NewPriceUpdateOrchestrator(db, tenantID, credService)
}

// countBatchResults tallies success/failure counts from raw batch results.
func countBatchResults(rawResults []interface{}) (results []interface{}, successCount, failedCount int) {
	results = make([]interface{}, 0, len(rawResults))
	for _, r := range rawResults {
		results = append(results, r)
		if m, ok := r.(map[string]interface{}); ok {
			if success, exists := m["success"]; exists && success == false {
				failedCount++
				continue
			}
		}
		successCount++
	}
	return results, successCount, failedCount
}

// respondPriceMultiPlatform sends the multi-platform price update response.
func respondPriceMultiPlatform(c *gin.Context, sku string, allResults map[string]*inventoryService.PlatformPriceResult, priceValue float64, lastErr error) bool {
	if len(allResults) == 0 && lastErr != nil {
		c.JSON(500, gin.H{"success": false, "error": lastErr.Error()})
		return true
	}
	c.JSON(200, gin.H{"success": true, "data": gin.H{"sku": sku, "platforms": allResults}, "price_from_inventory": priceValue})
	return true
}
