// Package master_product provides import handlers for Master Product
package master_product

// ImportRequest represents the import request body
type ImportRequest struct {
	Platform string `json:"platform" binding:"required"`
	ItemID   string `json:"item_id" binding:"required"`
}

// AutoMapRequest represents the auto-map request body
type AutoMapRequest struct {
	SellerSku string `json:"seller_sku" binding:"required"`
}

// AutoMapBatchRequest represents batch auto-map request body
type AutoMapBatchRequest struct {
	Skus []string `json:"skus" binding:"required"`
}

// ManualLinkRequest represents the manual link request body
type ManualLinkRequest struct {
	MasterSkuID    uint   `json:"master_sku_id" binding:"required"`
	Platform       string `json:"platform" binding:"required"`
	PlatformItemID string `json:"platform_item_id" binding:"required"`
	PlatformSkuID  string `json:"platform_sku_id"`
}

// UnlinkRequest represents the unlink request body
type UnlinkRequest struct {
	MasterSkuID uint   `json:"master_sku_id" binding:"required"`
	Platform    string `json:"platform" binding:"required"`
}
