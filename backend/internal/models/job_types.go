// Package models provides job type constants and data structures
package models

// Job type constants for long-running operations
const (
	// Escrow sync jobs
	JobTypeShopeeEscrowSync = "shopee_escrow_sync"
	JobTypeTiktokEscrowSync = "tiktok_escrow_sync"

	// Order sync jobs
	JobTypeShopeeOrderSync = "shopee_order_sync"
	JobTypeLazadaOrderSync = "lazada_order_sync"
	JobTypeTiktokOrderSync = "tiktok_order_sync"

	// Product sync jobs
	JobTypeShopeeProductSync = "shopee_product_sync"
	JobTypeLazadaProductSync = "lazada_product_sync"
	JobTypeTiktokProductSync = "tiktok_product_sync"

	// Inventory jobs
	JobTypeInventorySyncFromSheets = "inventory_sync_from_sheets"
	JobTypeBatchStockUpdate        = "batch_stock_update"
	JobTypeBatchPriceUpdate        = "batch_price_update"

	// Ads jobs
	JobTypeShopeeAdsSync = "shopee_ads_sync"
	JobTypeTiktokAdsSync = "tiktok_ads_sync"

	// Browser-automation jobs. Unlike the sync jobs above, these run in the
	// operator's own browser via a paired extension rather than against a
	// platform API.
	JobTypeShopeeScrape = "shopee_scrape"
)

// EscrowSyncJobData represents escrow sync job payload
type EscrowSyncJobData struct {
	TenantID    string `json:"tenant_id"`
	Platform    string `json:"platform"` // "shopee" or "tiktok"
	Month       int    `json:"month"`
	Year        int    `json:"year"`
	ForceResync bool   `json:"force_resync"`
}

// OrderSyncJobData represents order sync job payload
type OrderSyncJobData struct {
	TenantID  string `json:"tenant_id"`
	Platform  string `json:"platform"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
	Status    string `json:"status,omitempty"`
}

// ProductSyncJobData represents product sync job payload
type ProductSyncJobData struct {
	TenantID string `json:"tenant_id"`
	Platform string `json:"platform"`
}

// InventorySyncJobData represents inventory sync job payload
type InventorySyncJobData struct {
	TenantID      string `json:"tenant_id"`
	SpreadsheetID string `json:"spreadsheet_id,omitempty"`
}

// BatchUpdateJobData represents batch stock/price update job payload
type BatchUpdateJobData struct {
	TenantID string                   `json:"tenant_id"`
	Platform string                   `json:"platform"`
	Items    []map[string]interface{} `json:"items"`
}

// EscrowSyncResult represents escrow sync job result
type EscrowSyncResult struct {
	TotalOrders     int    `json:"total_orders"`
	ProcessedOrders int    `json:"processed_orders"`
	FailedOrders    int    `json:"failed_orders"`
	TotalItems      int    `json:"total_items"`
	Message         string `json:"message"`
}

// ContextKey type for context values
type ContextKey string

// Context keys for job execution
const (
	ContextKeyJobID    ContextKey = "job_id"
	ContextKeyTenantID ContextKey = "tenant_id"
)
