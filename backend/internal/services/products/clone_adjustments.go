// Package products provides title/description auto-truncation for cloning
package products

import (
	"strings"
	"unicode"
)

// Platform character limits
const (
	ShopeeTitleLimit = 120
	ShopeeDescLimit  = 3000

	TiktokTitleLimit = 80
	TiktokDescLimit  = 10000

	LazadaTitleLimit = 255
	LazadaDescLimit  = 25000
)

// AdjustForPlatform adjusts product data for target platform limits
// Returns a new ProductData with adjusted title and description
func AdjustForPlatform(productData *ProductData, targetPlatform string) *ProductData {
	if productData == nil {
		return nil
	}

	// Create copy to avoid mutating original
	adjusted := *productData

	switch targetPlatform {
	case "tiktok":
		adjusted.Name = truncateTitle(adjusted.Name, TiktokTitleLimit)
		adjusted.Description = truncateDescription(adjusted.Description, TiktokDescLimit)
	case "lazada":
		adjusted.Name = truncateTitle(adjusted.Name, LazadaTitleLimit)
		adjusted.Description = truncateDescription(adjusted.Description, LazadaDescLimit)
	case "shopee":
		adjusted.Name = truncateTitle(adjusted.Name, ShopeeTitleLimit)
		adjusted.Description = truncateDescription(adjusted.Description, ShopeeDescLimit)
	}

	return &adjusted
}

// GetPlatformLimits returns the character limits for a platform
func GetPlatformLimits(platform string) (titleLimit, descLimit int) {
	switch platform {
	case "shopee":
		return ShopeeTitleLimit, ShopeeDescLimit
	case "tiktok":
		return TiktokTitleLimit, TiktokDescLimit
	case "lazada":
		return LazadaTitleLimit, LazadaDescLimit
	default:
		return 255, 10000 // Safe defaults
	}
}

// truncateTitle truncates at word boundary, adds "..." if truncated
func truncateTitle(title string, limit int) string {
	title = strings.TrimSpace(title)

	// If within limit, return as-is
	if len(title) <= limit {
		return title
	}

	// Account for "..." suffix (3 chars)
	maxLen := limit - 3
	if maxLen < 10 {
		maxLen = 10 // Minimum reasonable length
	}

	// Find last space before limit
	truncated := title[:maxLen]
	lastSpace := strings.LastIndexFunc(truncated, unicode.IsSpace)

	if lastSpace > maxLen/2 { // Only use word boundary if it's reasonable
		truncated = truncated[:lastSpace]
	}

	return strings.TrimSpace(truncated) + "..."
}

// truncateDescription truncates at sentence boundary if possible
func truncateDescription(desc string, limit int) string {
	desc = strings.TrimSpace(desc)

	// If within limit, return as-is
	if len(desc) <= limit {
		return desc
	}

	// Account for "..." suffix
	maxLen := limit - 3
	if maxLen < 50 {
		maxLen = 50 // Minimum reasonable length
	}

	truncated := desc[:maxLen]

	// Try to find sentence boundary (last . ! or ?)
	sentenceEnd := -1
	for i := len(truncated) - 1; i >= maxLen/2; i-- {
		if truncated[i] == '.' || truncated[i] == '!' || truncated[i] == '?' {
			sentenceEnd = i + 1
			break
		}
	}

	if sentenceEnd > maxLen/2 {
		return strings.TrimSpace(truncated[:sentenceEnd])
	}

	// Fallback: find last space (word boundary)
	lastSpace := strings.LastIndexFunc(truncated, unicode.IsSpace)
	if lastSpace > maxLen/2 {
		truncated = truncated[:lastSpace]
	}

	return strings.TrimSpace(truncated) + "..."
}

// WillTruncate checks if content will be truncated for target platform
func WillTruncate(productData *ProductData, targetPlatform string) (titleTruncated, descTruncated bool) {
	titleLimit, descLimit := GetPlatformLimits(targetPlatform)

	if productData == nil {
		return false, false
	}

	titleTruncated = len(productData.Name) > titleLimit
	descTruncated = len(productData.Description) > descLimit

	return titleTruncated, descTruncated
}
