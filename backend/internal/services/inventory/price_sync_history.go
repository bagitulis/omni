// Package inventory provides price sync history recording
package inventory

import (
	"context"
	"encoding/json"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// isPriceNotFoundError checks if error is a "not found" error
func isPriceNotFoundError(errMsg string) bool {
	return errMsg == "SKU not found in Shopee" ||
		errMsg == "SKU not found in Lazada" ||
		errMsg == "SKU not found in TikTok"
}

func summarizePriceResults(platformResults map[string]*PlatformPriceResult) (bool, []string) {
	errors := make([]string, 0)
	successCount := 0

	for _, platformResult := range platformResults {
		if platformResult == nil {
			continue
		}

		if platformResult.Success {
			successCount++
			continue
		}

		if platformResult.Error != "" && !isPriceNotFoundError(platformResult.Error) {
			errors = append(errors, platformResult.Error)
		}
	}

	return successCount > 0, errors
}

func recordPriceSyncHistory(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
	sku string,
	price float64,
	result *PriceUpdateOrchestratorResult,
) {
	if result == nil {
		return
	}

	repo := repositories.NewMarketplaceSyncHistoryRepo(db)

	for platform, platformResult := range result.Platforms {
		status := "failed"
		if platformResult.Success {
			status = "success"
		}

		requestDataBytes, _ := json.Marshal(map[string]interface{}{
			"sku":      sku,
			"platform": platform,
			"price":    price,
		})
		responseDataBytes, _ := json.Marshal(platformResult)

		var errorMessage *string
		if platformResult.Error != "" {
			errorMessage = toPtr(platformResult.Error)
		}

		entry := &models.MarketplaceSyncHistory{
			TenantID:     tenantID,
			SKU:          sku,
			Platform:     platform,
			Operation:    "price_update",
			Status:       status,
			RequestData:  toPtr(string(requestDataBytes)),
			ResponseData: toPtr(string(responseDataBytes)),
			ErrorMessage: errorMessage,
		}

		if err := repo.Create(ctx, entry); err != nil {
			log.Error().
				Err(err).
				Str("tenant_id", tenantID).
				Str("sku", sku).
				Str("platform", platform).
				Msg("Failed to record price sync history")
		}
	}
}
