package models

// StockDTO represents stock data for API response
type StockDTO struct {
	SKU          string               `json:"sku"`
	ProductName  string               `json:"productName"`
	CurrentStock int                  `json:"currentStock"`
	MinStock     int                  `json:"minStock"`
	IsLowStock   bool                 `json:"isLowStock"`
	Platforms    []PlatformStockInfo  `json:"platforms"`
}

// PlatformStockInfo represents stock info per platform
type PlatformStockInfo struct {
	Platform       string `json:"platform"`
	PlatformItemID string `json:"platformItemId"`
	PlatformSKU    string `json:"platformSku"`
	Stock          int    `json:"stock"`
	Status         string `json:"status"`
	LastSyncedAt   string `json:"lastSyncedAt,omitempty"`
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
	SKU            string                 `json:"sku"`
	Success        bool                   `json:"success"`
	Message        string                 `json:"message,omitempty"`
	PlatformResults []PlatformUpdateResult `json:"platformResults,omitempty"`
}

// PlatformUpdateResult represents result per platform
type PlatformUpdateResult struct {
	Platform   string `json:"platform"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
	OldStock   int    `json:"oldStock,omitempty"`
	NewStock   int    `json:"newStock,omitempty"`
}

// BulkStockUpdateResult represents bulk update result
type BulkStockUpdateResult struct {
	TotalRequested int                 `json:"totalRequested"`
	TotalSuccess   int                 `json:"totalSuccess"`
	TotalFailed    int                 `json:"totalFailed"`
	Results        []StockUpdateResult `json:"results"`
}

// StockAlert represents a low stock alert
type StockAlert struct {
	SKU          string `json:"sku"`
	ProductName  string `json:"productName"`
	CurrentStock int    `json:"currentStock"`
	MinStock     int    `json:"minStock"`
	Shortage     int    `json:"shortage"`
	Platform     string `json:"platform,omitempty"`
}
