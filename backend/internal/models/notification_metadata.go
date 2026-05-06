package models

// BulkOperationMetadata is stored as JSON in the Notification.Metadata field.
// It provides structured data for bulk stock/price sync results.
type BulkOperationMetadata struct {
	OperationType string                       `json:"operation_type"` // "stock_sync" or "price_sync"
	Total         int                          `json:"total"`
	Succeeded     int                          `json:"succeeded"`
	Failed        int                          `json:"failed"`
	Platforms     map[string]PlatformSyncStats `json:"platforms"`
	FailedItems   []FailedItemDetail           `json:"failed_items,omitempty"`
	RequestID     string                       `json:"request_id,omitempty"` // Correlation ID for the batch
}

// PlatformSyncStats holds per-platform success/failure counts.
type PlatformSyncStats struct {
	Succeeded int    `json:"succeeded"`
	Failed    int    `json:"failed"`
	RequestID string `json:"request_id,omitempty"` // Platform-specific request_id
}

// FailedItemDetail holds info about a single failed item in a bulk operation.
type FailedItemDetail struct {
	SKU       string `json:"sku"`
	Platform  string `json:"platform"`
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}
