package models

// Constants for notification types
const (
	NotifTypeSuccess = "success"
	NotifTypeError   = "error"
	NotifTypeWarning = "warning"
	NotifTypeInfo    = "info"
)

// Constants for notification categories
const (
	CatSync      = "sync"
	CatOrder     = "order"
	CatProduct   = "product"
	CatInventory = "inventory"
	CatSystem    = "system"
	CatAuth      = "auth"
	CatExport    = "export"
)

// JobTypeToCategory maps background job types to notification categories
func JobTypeToCategory(jobType string) string {
	switch jobType {
	case "shopee_sync", "tiktok_sync", "lazada_sync", "platform_sync_all":
		return CatSync
	case "order_sync", "escrow_sync":
		return CatOrder
	case "product_import", "product_update":
		return CatProduct
	case "inventory_sync":
		return CatInventory
	case "ads_upload", "ads_ml_refresh":
		return CatSystem
	default:
		return CatSystem
	}
}

// JobTypeToTitle provides a human-readable title for a job result
func JobTypeToTitle(jobType string, success bool) string {
	status := "Completed"
	if !success {
		status = "Failed"
	}

	switch jobType {
	case "shopee_sync":
		return "Shopee Sync " + status
	case "tiktok_sync":
		return "TikTok Sync " + status
	case "lazada_sync":
		return "Lazada Sync " + status
	case "inventory_sync":
		return "Inventory Sync " + status
	case "ads_upload":
		return "Ads Upload " + status
	default:
		return "Background Process " + status
	}
}
