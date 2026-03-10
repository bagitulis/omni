// Package products provides product cloning - common helper functions
package products

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// =============================================================================
// Platform Router Functions - Single Responsibility
// =============================================================================

// fetchProductData routes fetch to appropriate platform handler
func (s *CloneService) fetchProductData(ctx context.Context, platform, itemID string) (*ProductData, error) {
	switch platform {
	case "shopee":
		return s.fetchShopeeProductFromDB(ctx, itemID)
	case "lazada":
		return s.fetchLazadaProductFromDB(ctx, itemID)
	case "tiktok":
		return s.fetchTiktokProductFromDB(ctx, itemID)
	default:
		return nil, fmt.Errorf("unsupported platform: %s", platform)
	}
}

// createProduct routes creation to appropriate platform handler
func (s *CloneService) createProduct(ctx context.Context, platform string, data *ProductData) (string, error) {
	if s.credService == nil {
		return "", fmt.Errorf("credential service not initialized")
	}

	switch platform {
	case "shopee":
		return s.createShopeeProduct(ctx, data)
	case "lazada":
		return s.createLazadaProduct(ctx, data)
	case "tiktok":
		return s.createTiktokProduct(ctx, data)
	default:
		return "", fmt.Errorf("unsupported platform: %s", platform)
	}
}

// =============================================================================
// Category Mapping
// =============================================================================

// mapCategory handles cross-platform category mapping
// For TikTok, Lazada, and Shopee as target, we support auto-recommendation
func (s *CloneService) mapCategory(_ context.Context, sourcePlatform, targetPlatform, sourceCategory string) (string, error) {
	if sourcePlatform == targetPlatform {
		return sourceCategory, nil
	}

	// For all platforms as target, we use auto-recommendation in createProduct
	// Empty string triggers auto-recommendation
	return "", nil // Will trigger auto-recommendation
}

// =============================================================================
// Image Parsing Functions
// =============================================================================

// parseImagesFromDB parses image string (JSON array or single URL) to []string
func parseImagesFromDB(imageData interface{}) []string {
	if imageData == nil {
		return []string{}
	}

	switch v := imageData.(type) {
	case string:
		if v == "" {
			return []string{}
		}
		// Try parsing as JSON array
		var images []string
		if err := json.Unmarshal([]byte(v), &images); err == nil {
			return images
		}
		// Return as single image
		return []string{v}
	case []string:
		return v
	case []interface{}:
		images := make([]string, 0, len(v))
		for _, img := range v {
			if str, ok := img.(string); ok {
				images = append(images, str)
			}
		}
		return images
	}
	return []string{}
}

// =============================================================================
// TikTok Image Builders
// =============================================================================

// buildTiktokImageInfos builds TikTok ImageInfo array for CreateProductRequest
func buildTiktokImageInfos(images []string) []tiktokPkg.ImageInfo {
	result := make([]tiktokPkg.ImageInfo, 0, len(images))
	for _, img := range images {
		result = append(result, tiktokPkg.ImageInfo{URI: img})
	}
	return result
}

// =============================================================================
// Utility Functions
// =============================================================================

// minInt returns the minimum of two integers
func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// getStringOrEmpty returns first non-empty string value
func getStringOrEmpty(values ...interface{}) string {
	for _, v := range values {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case string:
			if val != "" {
				return val
			}
		case *string:
			if val != nil && *val != "" {
				return *val
			}
		}
	}
	return ""
}

// getFloat64 returns first valid float64 value
func getFloat64(values ...interface{}) float64 {
	for _, v := range values {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case float64:
			if val > 0 {
				return val
			}
		case *float64:
			if val != nil && *val > 0 {
				return *val
			}
		case int:
			if val > 0 {
				return float64(val)
			}
		case *int:
			if val != nil && *val > 0 {
				return float64(*val)
			}
		}
	}
	return 0
}

// getInt returns first valid int value
func getInt(values ...interface{}) int {
	for _, v := range values {
		if v == nil {
			continue
		}
		switch val := v.(type) {
		case int:
			if val >= 0 {
				return val
			}
		case *int:
			if val != nil && *val >= 0 {
				return *val
			}
		case int64:
			if val >= 0 {
				return int(val)
			}
		case *int64:
			if val != nil && *val >= 0 {
				return int(*val)
			}
		case float64:
			if val >= 0 {
				return int(val)
			}
		}
	}
	return 0
}

// =============================================================================
// Result Helpers
// =============================================================================

func (s *CloneService) failResult(result *CloneResult, format string, args ...interface{}) (*CloneResult, error) {
	result.Status = "failed"
	result.Message = fmt.Sprintf(format, args...)
	return result, nil
}

func (s *CloneService) successResult(result *CloneResult, targetItemID string) (*CloneResult, error) {
	result.Progress = 100
	result.Status = "completed"
	result.TargetItemID = targetItemID
	result.Message = "Product cloned successfully"
	now := time.Now()
	result.CompletedAt = &now
	return result, nil
}

func (s *CloneService) successResultWithSync(result *CloneResult, targetItemID, syncResult string) (*CloneResult, error) {
	result.Progress = 100
	result.TargetItemID = targetItemID
	result.SyncTriggered = true
	result.SyncResult = syncResult
	now := time.Now()
	result.CompletedAt = &now

	if strings.Contains(syncResult, "failed") || strings.Contains(syncResult, "error") {
		result.Status = "completed_with_warnings"
		result.Message = "Product cloned but sync had issues"
	} else {
		result.Status = "completed"
		result.Message = "Product cloned and synced successfully"
	}

	log.Printf("[Clone] Product cloned to %s, sync result: %s", result.TargetPlatform, syncResult)
	return result, nil
}

func (s *CloneService) determineBatchStatus(failed, success int) string {
	if failed == 0 {
		return "completed"
	}
	if success == 0 {
		return "failed"
	}
	return "partial"
}

func generateCloneID() string {
	return fmt.Sprintf("clone-%d", time.Now().UnixNano())
}
