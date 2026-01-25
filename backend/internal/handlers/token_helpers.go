package handlers

import (
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
)

// FormatTokenStatusText formats token status as human-readable text
// Matches Node.js formatTokenStatus function
func FormatTokenStatusText(platform string, status *services.TokenInfo) string {
	platformTitle := getPlatformTitle(platform)

	if status == nil {
		return platformTitle + " Token Status:\n  ❌ Not configured"
	}

	var result string
	result = platformTitle + " Token Status:\n"

	// Access token status
	if status.IsValid {
		result += "  ✅ Valid\n"
	} else {
		result += "  ❌ Expired or Invalid\n"
	}

	// Expiry info
	if !status.ExpiresAt.IsZero() {
		remaining := time.Until(status.ExpiresAt)
		if remaining > 0 {
			days := int(remaining.Hours() / 24)
			hours := int(remaining.Hours()) % 24
			mins := int(remaining.Minutes()) % 60
			result += "  Expires in: " + FormatDurationHuman(days, hours, mins) + "\n"
		} else {
			result += "  ❌ Expired\n"
		}
	}

	// Refresh token expiry
	if !status.RefreshTokenExpires.IsZero() {
		remaining := time.Until(status.RefreshTokenExpires)
		if remaining > 0 {
			days := int(remaining.Hours() / 24)
			hours := int(remaining.Hours()) % 24
			mins := int(remaining.Minutes()) % 60
			result += "  Refresh expires: " + FormatDurationHuman(days, hours, mins)
		} else {
			result += "  ❌ Refresh token expired"
		}
	}

	return result
}

// FormatDurationHuman formats days, hours, minutes as human-readable string
func FormatDurationHuman(days, hours, mins int) string {
	if days > 0 {
		return time.Duration(days*24*int(time.Hour) + hours*int(time.Hour) + mins*int(time.Minute)).String()
	}
	return time.Duration(hours*int(time.Hour) + mins*int(time.Minute)).String()
}

// getPlatformTitle returns the display title for a platform
func getPlatformTitle(platform string) string {
	switch platform {
	case models.PlatformShopee:
		return "Shopee"
	case models.PlatformLazada:
		return "Lazada"
	case models.PlatformTiktok:
		return "TikTok"
	default:
		return platform
	}
}

// FormatISOTimestamp formats time as ISO 8601 UTC timestamp
func FormatISOTimestamp(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05.000Z")
}

// ValidatePlatform checks if a platform string is valid
func ValidatePlatform(platform string) bool {
	return platform == models.PlatformShopee ||
		platform == models.PlatformLazada ||
		platform == models.PlatformTiktok
}

// GetAllPlatforms returns all supported platforms
func GetAllPlatforms() []string {
	return []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}
}
