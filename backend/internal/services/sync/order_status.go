package sync

// PlatformType represents supported e-commerce platforms
type PlatformType string

const (
	PlatformShopee PlatformType = "shopee"
	PlatformLazada PlatformType = "lazada"
	PlatformTiktok PlatformType = "tiktok"
)

// OrderStatusCategory represents order status categories
// These match the Node.js ORDER_STATUS_MAPPINGS exactly
type OrderStatusCategory string

const (
	// Node.js categories from orderSyncOperations.ts
	StatusUnpaid    OrderStatusCategory = "unpaid"    // Orders waiting for payment
	StatusUnprocess OrderStatusCategory = "unprocess" // Orders ready to ship/pack
	StatusProcessed OrderStatusCategory = "processed" // Orders being processed/shipped
	// Additional standard categories
	StatusShipped   OrderStatusCategory = "shipped"
	StatusCompleted OrderStatusCategory = "completed"
	StatusCancelled OrderStatusCategory = "cancelled"
)

// OrderStatusMapping maps categories to platform-specific statuses
type OrderStatusMapping map[OrderStatusCategory]string

// OrderStatusMappings contains all platform status mappings
// These match Node.js ORDER_STATUS_MAPPINGS exactly from orderSyncOperations.ts
var OrderStatusMappings = map[PlatformType]OrderStatusMapping{
	PlatformShopee: {
		// Node.js: shopee: { unpaid: "UNPAID", unprocess: "READY_TO_SHIP", processed: "PROCESSED" }
		StatusUnpaid:    "UNPAID",
		StatusUnprocess: "READY_TO_SHIP",
		StatusProcessed: "PROCESSED",
		StatusShipped:   "SHIPPED",
		StatusCompleted: "COMPLETED",
		StatusCancelled: "CANCELLED",
	},
	PlatformLazada: {
		// Node.js: lazada: { unprocess: "topack", processed: "toship" }
		// Note: Lazada tidak punya unpaid status
		StatusUnpaid:    "",       // Lazada tidak support unpaid
		StatusUnprocess: "topack", // Ready to pack
		StatusProcessed: "toship", // Ready to ship
		StatusShipped:   "shipped",
		StatusCompleted: "delivered",
		StatusCancelled: "canceled",
	},
	PlatformTiktok: {
		// Node.js: tiktok: { unpaid: "UNPAID", unprocess: "AWAITING_SHIPMENT", processed: "AWAITING_COLLECTION" }
		StatusUnpaid:    "UNPAID",
		StatusUnprocess: "AWAITING_SHIPMENT",
		StatusProcessed: "AWAITING_COLLECTION",
		StatusShipped:   "IN_TRANSIT",
		StatusCompleted: "COMPLETED",
		StatusCancelled: "CANCELLED",
	},
}

// GetPlatformStatus returns the platform-specific status for a category
func GetPlatformStatus(platform PlatformType, category OrderStatusCategory) string {
	if mapping, ok := OrderStatusMappings[platform]; ok {
		if status, exists := mapping[category]; exists {
			return status
		}
	}
	return ""
}

// GetCategoryFromStatus converts platform status to category
func GetCategoryFromStatus(platform PlatformType, status string) OrderStatusCategory {
	if mapping, ok := OrderStatusMappings[platform]; ok {
		for category, s := range mapping {
			if s == status {
				return category
			}
		}
	}
	return ""
}

// AllPlatforms returns all supported platforms
func AllPlatforms() []PlatformType {
	return []PlatformType{PlatformShopee, PlatformLazada, PlatformTiktok}
}

// AllCategories returns all status categories
func AllCategories() []OrderStatusCategory {
	return []OrderStatusCategory{
		StatusUnpaid,
		StatusUnprocess,
		StatusProcessed,
		StatusShipped,
		StatusCompleted,
		StatusCancelled,
	}
}
