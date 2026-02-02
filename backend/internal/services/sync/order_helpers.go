package sync

import (
	"fmt"
	"time"
)

// formatShipByDateFromDB converts Unix timestamp to countdown string for display
// Used by flattenShopeeOrders to show shipping deadline in frontend
func formatShipByDateFromDB(timestamp int64) string {
	shipTime := time.Unix(timestamp, 0)
	now := time.Now()

	if shipTime.Before(now) {
		return "Overdue"
	}

	diff := shipTime.Sub(now)
	days := int(diff.Hours() / 24)
	hours := int(diff.Hours()) % 24

	if days > 0 {
		return fmt.Sprintf("%dd %dh left", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh left", hours)
	}
	return "< 1h left"
}
