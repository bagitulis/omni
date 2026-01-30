// Package master_product provides import service for Master Product
package master_product

import (
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/pkg/shopee"
)

// ImportResult represents the result of an import operation
type ImportResult struct {
	MasterProduct *models.MasterProduct `json:"master_product"`
	SkusImported  int                   `json:"skus_imported"`
	SkusSkipped   int                   `json:"skus_skipped"`
	PlatformLinks int                   `json:"platform_links"`
}

// PreviewResult represents what will be imported (without creating)
type PreviewResult struct {
	Title         string       `json:"title"`
	Description   string       `json:"description"`
	Images        []string     `json:"images"`
	SKUs          []PreviewSku `json:"skus"`
	TotalSKUs     int          `json:"total_skus"`
	ValidSKUs     int          `json:"valid_skus"`
	SkippedSKUs   int          `json:"skipped_skus"`
	AlreadyExists bool         `json:"already_exists"`
}

// PreviewSku represents a SKU in preview
type PreviewSku struct {
	SellerSku   string  `json:"seller_sku"`
	VariantName string  `json:"variant_name"`
	Price       float64 `json:"price"`
	Stock       int     `json:"stock"`
	Valid       bool    `json:"valid"`
	SkipReason  string  `json:"skip_reason,omitempty"`
}

// Helper functions

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}

func limitImages(images []string, max int) []string {
	if len(images) <= max {
		return images
	}
	return images[:max]
}

func buildVariantName(tiers []shopee.TierVariation, indices []int) string {
	if len(tiers) == 0 || len(indices) == 0 {
		return ""
	}

	var parts []string
	for i, tierIdx := range indices {
		if i < len(tiers) && tierIdx < len(tiers[i].OptionList) {
			parts = append(parts, tiers[i].OptionList[tierIdx].Option)
		}
	}

	if len(parts) == 0 {
		return ""
	}

	result := parts[0]
	for i := 1; i < len(parts); i++ {
		result += " / " + parts[i]
	}
	return result
}
