// Package products provides clone conflict detection functionality
package products

import (
	"context"
	"fmt"
)

// CheckConflict checks if cloning would create a conflict (product already exists)
func (s *CloneService) CheckConflict(ctx context.Context, req CloneRequest) (*ConflictResult, error) {
	result := &ConflictResult{
		HasConflict: false,
	}

	// 1. Fetch source product data
	sourceData, err := s.fetchProductData(ctx, req.SourcePlatform, req.SourceItemID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch source product: %w", err)
	}

	// Build source summary
	sku := req.SKU
	if sku == "" && len(sourceData.Variants) > 0 {
		sku = sourceData.Variants[0].SKU
	}

	result.SourceProduct = &ProductSummary{
		Platform: req.SourcePlatform,
		ItemID:   req.SourceItemID,
		SKU:      sku,
		Name:     sourceData.Name,
		Price:    sourceData.Price,
		Stock:    sourceData.Stock,
		Images:   sourceData.Images,
	}

	// 2. Check adjustments that will be made
	titleTrunc, descTrunc := WillTruncate(sourceData, req.TargetPlatform)
	if titleTrunc || descTrunc {
		titleLimit, descLimit := GetPlatformLimits(req.TargetPlatform)
		adjusted := AdjustForPlatform(sourceData, req.TargetPlatform)

		result.Adjustments = &AdjustmentInfo{
			TitleWillTruncate: titleTrunc,
			DescWillTruncate:  descTrunc,
			OriginalTitle:     sourceData.Name,
			AdjustedTitle:     adjusted.Name,
			TitleLimit:        titleLimit,
			DescLimit:         descLimit,
		}
	}

	// 3. Check if SKU already exists on target platform
	if sku != "" && s.checkSkuOnPlatform(ctx, req.TargetPlatform, sku) {
		result.HasConflict = true

		// Try to fetch target product for comparison
		targetData, err := s.fetchProductBySKU(ctx, req.TargetPlatform, sku)
		if err == nil && targetData != nil {
			result.TargetProduct = &ProductSummary{
				Platform: req.TargetPlatform,
				ItemID:   targetData.ItemID,
				SKU:      sku,
				Name:     targetData.Name,
				Price:    targetData.Price,
				Stock:    targetData.Stock,
				Images:   targetData.Images,
			}

			// Build differences list
			result.Differences = buildDifferences(result.SourceProduct, result.TargetProduct)
		}
	}

	return result, nil
}

// fetchProductBySKU fetches product data by SKU from a platform
func (s *CloneService) fetchProductBySKU(ctx context.Context, platform, sku string) (*ProductData, error) {
	// Use the existing fetch mechanism but search by SKU
	platformIds, err := s.idFetcher.FetchBySku(ctx, sku)
	if err != nil {
		return nil, err
	}

	var itemID string
	switch platform {
	case "shopee":
		if platformIds.Shopee != nil {
			itemID = platformIds.Shopee.ItemID
		}
	case "lazada":
		if platformIds.Lazada != nil {
			itemID = platformIds.Lazada.ItemID
		}
	case "tiktok":
		if platformIds.Tiktok != nil {
			itemID = platformIds.Tiktok.ProductID
		}
	}

	if itemID == "" {
		return nil, fmt.Errorf("product not found on %s", platform)
	}

	return s.fetchProductData(ctx, platform, itemID)
}

// buildDifferences compares source and target products
func buildDifferences(source, target *ProductSummary) []Difference {
	var diffs []Difference

	if source.Name != target.Name {
		diffs = append(diffs, Difference{
			Field:       "name",
			SourceValue: source.Name,
			TargetValue: target.Name,
		})
	}

	if source.Price != target.Price {
		diffs = append(diffs, Difference{
			Field:       "price",
			SourceValue: fmt.Sprintf("%.2f", source.Price),
			TargetValue: fmt.Sprintf("%.2f", target.Price),
		})
	}

	if source.Stock != target.Stock {
		diffs = append(diffs, Difference{
			Field:       "stock",
			SourceValue: fmt.Sprintf("%d", source.Stock),
			TargetValue: fmt.Sprintf("%d", target.Stock),
		})
	}

	if len(source.Images) != len(target.Images) {
		diffs = append(diffs, Difference{
			Field:       "image_count",
			SourceValue: fmt.Sprintf("%d", len(source.Images)),
			TargetValue: fmt.Sprintf("%d", len(target.Images)),
		})
	}

	return diffs
}
