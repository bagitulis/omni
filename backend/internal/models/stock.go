package models

// StockDTO represents stock data for API response
type StockDTO struct {
	SKU          string              `json:"sku"`
	ProductName  string              `json:"product_name"`
	CurrentStock int                 `json:"current_stock"`
	MinStock     int                 `json:"min_stock"`
	IsLowStock   bool                `json:"is_low_stock"`
	Platforms    []PlatformStockInfo `json:"platforms"`
}

// PlatformStockInfo represents stock info per platform
type PlatformStockInfo struct {
	Platform       string `json:"platform"`
	PlatformItemID string `json:"platform_item_id"`
	PlatformSKU    string `json:"platform_sku"`
	Stock          int    `json:"stock"`
	Status         string `json:"status"`
	LastSyncedAt   string `json:"last_synced_at,omitempty"`
}

// StockUpdateRequest represents request to update stock
type StockUpdateRequest struct {
	SKU       string   `json:"sku" binding:"required"`
	Quantity  int      `json:"quantity" binding:"required"`
	Platforms []string `json:"platforms"` // empty = all platforms
}

// BulkStockUpdateRequest represents bulk stock update request
type BulkStockUpdateRequest struct {
	Updates []StockUpdateRequest `json:"updates" binding:"required"`
}

// StockUpdateResult represents result of stock update
type StockUpdateResult struct {
	SKU             string                 `json:"sku"`
	Success         bool                   `json:"success"`
	Message         string                 `json:"message,omitempty"`
	PlatformResults []PlatformUpdateResult `json:"platform_results,omitempty"`
}

// PlatformUpdateResult represents result per platform
type PlatformUpdateResult struct {
	Platform string `json:"platform"`
	Success  bool   `json:"success"`
	Message  string `json:"message,omitempty"`
	OldStock int    `json:"old_stock,omitempty"`
	NewStock int    `json:"new_stock,omitempty"`
}

// BulkStockUpdateResult represents bulk update result
type BulkStockUpdateResult struct {
	TotalRequested int                 `json:"total_requested"`
	TotalSuccess   int                 `json:"total_success"`
	TotalFailed    int                 `json:"total_failed"`
	Results        []StockUpdateResult `json:"results"`
}

// StockAlert represents a low stock alert
type StockAlert struct {
	SKU          string `json:"sku"`
	ProductName  string `json:"product_name"`
	CurrentStock int    `json:"current_stock"`
	MinStock     int    `json:"min_stock"`
	Shortage     int    `json:"shortage"`
	Platform     string `json:"platform,omitempty"`
}
