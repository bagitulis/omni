// Package master_product provides import handlers for Master Product
package master_product

// ImportRequest represents the import request body
type ImportRequest struct {
	Platform string `json:"platform" binding:"required"`
	ItemID   string `json:"item_id" binding:"required"`
}

// FileImportRequest represents file-based import payload
type FileImportRequest struct {
	Rows []FileImportRowRequest `json:"rows" binding:"required,min=1"`
}

// FileImportRowRequest represents one row from preview/import UI
type FileImportRowRequest struct {
	RowNumber   int      `json:"row_number"`
	ItemName    string   `json:"item_name"`
	ItemSku     string   `json:"item_sku"`
	VariantName string   `json:"variant_name,omitempty"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	BatchKey    string   `json:"batch_key,omitempty"`
	Description string   `json:"description,omitempty"`
	ImageUrls   []string `json:"image_urls,omitempty"`
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors,omitempty"`
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
