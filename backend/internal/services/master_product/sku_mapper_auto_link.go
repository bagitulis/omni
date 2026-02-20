package master_product

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// AutoMappedLink represents a successfully created auto-link.
type AutoMappedLink struct {
	MasterSkuID    uint   `json:"master_sku_id"`
	SellerSku      string `json:"seller_sku"`
	Platform       string `json:"platform"`
	PlatformItemID string `json:"platform_item_id"`
	PlatformSkuID  string `json:"platform_sku_id,omitempty"`
	SyncStatus     string `json:"sync_status"`
}

// AutoMapBatchResult summarizes batch auto-map + link execution.
type AutoMapBatchResult struct {
	MappedCount  int              `json:"mapped_count"`
	SkippedCount int              `json:"skipped_count"`
	Mappings     []AutoMappedLink `json:"mappings"`
	Errors       []string         `json:"errors,omitempty"`
}

// AutoMapAndLinkBySkus finds platform matches and creates SKU platform links.
func (m *SkuMapper) AutoMapAndLinkBySkus(ctx context.Context, sellerSkus []string) (*AutoMapBatchResult, error) {
	result := &AutoMapBatchResult{
		Mappings: make([]AutoMappedLink, 0),
		Errors:   make([]string, 0),
	}

	seen := make(map[string]struct{}, len(sellerSkus))

	for _, rawSku := range sellerSkus {
		sellerSku := strings.TrimSpace(rawSku)
		normalizedSku := normalizeSkuForLookup(rawSku)
		if sellerSku == "" {
			result.SkippedCount++
			result.Errors = append(result.Errors, "seller_sku is empty")
			continue
		}
		if _, ok := seen[normalizedSku]; ok {
			continue
		}
		seen[normalizedSku] = struct{}{}

		var masterSku models.MasterProductSku
		err := m.db.WithContext(ctx).
			Preload("PlatformLinks").
			Where("tenant_id = ? AND LOWER(seller_sku) = LOWER(?)", m.tenantID, normalizedSku).
			First(&masterSku).Error
		if err != nil {
			result.SkippedCount++
			if errors.Is(err, gorm.ErrRecordNotFound) {
				result.Errors = append(result.Errors, fmt.Sprintf("master SKU not found: %s", sellerSku))
				continue
			}
			return nil, fmt.Errorf("failed to load master SKU %s: %w", sellerSku, err)
		}

		platformMatches, err := m.AutoMapBySku(ctx, normalizedSku)
		if err != nil {
			return nil, fmt.Errorf("failed to auto-map SKU %s: %w", sellerSku, err)
		}

		existingLinks, err := m.repo.FindPlatformLinks(ctx, masterSku.MasterProductID)
		if err != nil {
			return nil, fmt.Errorf("failed to load existing platform links for SKU %s: %w", sellerSku, err)
		}

		mappedThisSku := false
		hasAmbiguousMatch := false

		for _, match := range platformMatches {
			if match.Ambiguous {
				hasAmbiguousMatch = true
				continue
			}
			if !match.Found || match.PlatformItemID == "" {
				continue
			}
			if hasPlatformLink(existingLinks, masterSku.ID, match.Platform) {
				continue
			}

			now := time.Now()
			masterSkuID := masterSku.ID
			link := &models.MasterProductPlatformLink{
				MasterProductID:   masterSku.MasterProductID,
				MasterSkuID:       &masterSkuID,
				Platform:          match.Platform,
				PlatformItemID:    match.PlatformItemID,
				PlatformProductID: match.PlatformItemID,
				PlatformSkuID:     match.PlatformSkuID,
				SyncStatus:        models.SyncStatusSynced,
				LastSyncedAt:      &now,
				CreatedAt:         now,
				UpdatedAt:         now,
			}

			if err := m.repo.CreatePlatformLink(ctx, link); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("failed to create link for SKU %s on %s: %v", sellerSku, match.Platform, err))
				continue
			}

			existingLinks = append(existingLinks, *link)
			mappedThisSku = true
			result.MappedCount++
			result.Mappings = append(result.Mappings, AutoMappedLink{
				MasterSkuID:    masterSku.ID,
				SellerSku:      sellerSku,
				Platform:       match.Platform,
				PlatformItemID: match.PlatformItemID,
				PlatformSkuID:  match.PlatformSkuID,
				SyncStatus:     models.SyncStatusSynced,
			})
		}

		if !mappedThisSku {
			result.SkippedCount++
			if hasAmbiguousMatch {
				result.Errors = append(result.Errors, fmt.Sprintf("ambiguous platform match for SKU %s", sellerSku))
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("no platform match found for SKU %s", sellerSku))
			}
		}
	}

	return result, nil
}

func hasPlatformLink(links []models.MasterProductPlatformLink, masterSkuID uint, platform string) bool {
	for _, link := range links {
		if link.MasterSkuID == nil {
			continue
		}
		if *link.MasterSkuID == masterSkuID && link.Platform == platform {
			return true
		}
	}
	return false
}
