package handlers

import "github.com/omni/backend/internal/models"

// =============================================================================
// Wholesale DTOs (Data Transfer Objects)
// All types use Shopee API format: min_count, max_count, unit_price
// =============================================================================

// WholesaleTier is an alias for models.WholesaleTier (DRY — single definition)
type WholesaleTier = models.WholesaleTier
type WholesaleInfo struct {
	ItemID       int64           `json:"item_id"`
	HasWholesale bool            `json:"has_wholesale"`
	Tiers        []WholesaleTier `json:"tiers"`
}

// =============================================================================
// Shopee Wholesale Requests
// =============================================================================

// UpdateWholesaleRequest represents wholesale update request
type UpdateWholesaleRequest struct {
	Tiers []WholesaleTier `json:"tiers" binding:"required"`
}

// BatchDeleteRequest represents batch delete request
type BatchDeleteRequest struct {
	ItemIDs []int64 `json:"item_ids" binding:"required"`
}

// BatchAddRequest represents batch add wholesale request
type BatchAddRequest struct {
	Items []BatchAddItem `json:"items" binding:"required"`
}

// BatchAddItem represents an item for batch wholesale add
type BatchAddItem struct {
	ItemID int64           `json:"item_id"`
	SKU    string          `json:"sku,omitempty"`
	Tiers  []WholesaleTier `json:"tiers"`
}

// PreviewRequest represents preview wholesale request
type PreviewRequest struct {
	SKUs          []string `json:"skus" binding:"required"`
	BasePrice     float64  `json:"base_price"`
	DiscountRates []int    `json:"discount_rates"`
}

// ImportWholesaleRequest represents import wholesale request
type ImportWholesaleRequest struct {
	Data []ImportWholesaleItem `json:"data" binding:"required"`
}

// ImportWholesaleItem represents an item to import
type ImportWholesaleItem struct {
	SKU   string          `json:"sku"`
	Tiers []WholesaleTier `json:"tiers"`
}

// BatchSetMpqRequest represents batch set MPQ request
type BatchSetMpqRequest struct {
	Items []MpqItem `json:"items" binding:"required"`
}

// MpqItem represents an MPQ item
type MpqItem struct {
	ItemID int64  `json:"item_id"`
	SKU    string `json:"sku,omitempty"`
	MPQ    int    `json:"mpq"`
}

// BatchWholesaleResetRequest represents batch wholesale reset request
type BatchWholesaleResetRequest struct {
	Items []BatchUpdateItem `json:"items" binding:"required"`
}

// =============================================================================
// TikTok Wholesale Requests
// =============================================================================

// TiktokBatchMpqRequest represents TikTok batch MPQ request
// Frontend sends: { products, items: [{sku, price}], mpq }
type TiktokBatchMpqRequest struct {
	Products []TiktokMpqItem `json:"products"`
	Items    []struct {
		SKU   string  `json:"sku"`
		Price float64 `json:"price"`
	} `json:"items"`
	MPQ int `json:"mpq" binding:"required,min=1"`
}

// TiktokMpqItem represents a TikTok MPQ product item
type TiktokMpqItem struct {
	ProductID string  `json:"product_id"`
	SKU       string  `json:"sku,omitempty"`
	SKUID     string  `json:"sku_id,omitempty"`
	Price     float64 `json:"price,omitempty"`
	MPQ       int     `json:"mpq"`
}

// TiktokWholesaleRequest represents TikTok wholesale request
type TiktokWholesaleRequest struct {
	Tiers []WholesaleTier `json:"tiers" binding:"required"`
}
