package tiktok

import (
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

func parsePrice(salePrice, originalPrice, taxExclusivePrice string) float64 {
	price := 0.0
	if salePrice != "" {
		_, _ = fmt.Sscanf(salePrice, "%f", &price)
		if price > 0 {
			return price
		}
	}

	if originalPrice != "" {
		_, _ = fmt.Sscanf(originalPrice, "%f", &price)
		if price > 0 {
			return price
		}
	}

	if taxExclusivePrice != "" {
		_, _ = fmt.Sscanf(taxExclusivePrice, "%f", &price)
	}

	return price
}

func sumInventory(inventory []struct {
	WarehouseID string `json:"warehouse_id"`
	Quantity    int    `json:"quantity"`
}) int {
	totalQty := 0
	for _, inv := range inventory {
		totalQty += inv.Quantity
	}

	return totalQty
}

func extractPreferredImageURLs(images []tiktokPkg.ProductImage) []string {
	imageURLs := make([]string, 0, len(images))

	for _, img := range images {
		if len(img.URLs) == 0 {
			continue
		}

		webpURL := ""
		fallbackURL := ""
		for _, u := range img.URLs {
			if u == "" {
				continue
			}

			if strings.HasSuffix(strings.ToLower(u), ".webp") || strings.Contains(strings.ToLower(u), "webp") {
				webpURL = u
				break
			}

			if fallbackURL == "" {
				fallbackURL = u
			}
		}

		if webpURL != "" {
			imageURLs = append(imageURLs, webpURL)
			continue
		}

		if fallbackURL != "" {
			imageURLs = append(imageURLs, fallbackURL)
		}
	}

	return imageURLs
}

func buildVariantName(attrs []tiktokPkg.ProductSalesAttr) string {
	if len(attrs) == 0 {
		return ""
	}

	parts := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		if attr.Name != "" && attr.ValueName != "" {
			parts = append(parts, attr.Name+":"+attr.ValueName)
		}
	}

	return strings.Join(parts, " | ")
}

func buildVariantData(attrs []tiktokPkg.ProductSalesAttr) models.JSONMap {
	if len(attrs) == 0 {
		return nil
	}

	result := make(models.JSONMap)
	for _, attr := range attrs {
		if attr.Name != "" && attr.ValueName != "" {
			result[attr.Name] = attr.ValueName
		}
	}

	if len(result) == 0 {
		return nil
	}

	return result
}
