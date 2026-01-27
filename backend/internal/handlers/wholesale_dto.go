package handlers

// =============================================================================
// Wholesale DTOs (Data Transfer Objects)
// =============================================================================

// WholesaleTier represents a wholesale tier
type WholesaleTier struct {
	MinQty int     `json:"minQty"`
	MaxQty int     `json:"maxQty"`
	Price  float64 `json:"price"`
}

// WholesaleInfo represents wholesale info for an item
type WholesaleInfo struct {
	ItemID  int64           `json:"itemId"`
	HasTier bool            `json:"hasTier"`
	Tiers   []WholesaleTier `json:"tiers"`
	MPQ     int             `json:"mpq"`
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
	ItemIDs []int64 `json:"itemIds" binding:"required"`
}

// BatchAddRequest represents batch add wholesale request
type BatchAddRequest struct {
	Items []BatchAddItem `json:"items" binding:"required"`
}

// BatchAddItem represents an item for batch wholesale add
type BatchAddItem struct {
	ItemID int64           `json:"itemId"`
	SKU    string          `json:"sku,omitempty"`
	Tiers  []WholesaleTier `json:"tiers"`
}

// PreviewRequest represents preview wholesale request
type PreviewRequest struct {
	SKUs          []string `json:"skus" binding:"required"`
	DiscountRates []int    `json:"discountRates" binding:"required"`
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
	ItemID int64  `json:"itemId"`
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
type TiktokBatchMpqRequest struct {
	Products []TiktokMpqItem `json:"products" binding:"required"`
}

// TiktokMpqItem represents a TikTok MPQ item
type TiktokMpqItem struct {
	ProductID string `json:"productId"`
	SKUID     string `json:"skuId,omitempty"`
	MPQ       int    `json:"mpq"`
}

// TiktokWholesaleRequest represents TikTok wholesale request
type TiktokWholesaleRequest struct {
	Tiers []WholesaleTier `json:"tiers" binding:"required"`
}

// =============================================================================
// Helper Functions
// =============================================================================

// GenerateTiers generates wholesale tiers based on discount rates
func GenerateTiers(basePrice float64, discountRates []int) []WholesaleTier {
	tiers := make([]WholesaleTier, 0, len(discountRates))
	for i, rate := range discountRates {
		minQty := (i + 1) * 5
		maxQty := (i + 2) * 5
		if i == len(discountRates)-1 {
			maxQty = 999
		}
		tiers = append(tiers, WholesaleTier{
			MinQty: minQty,
			MaxQty: maxQty,
			Price:  basePrice * float64(100-rate) / 100,
		})
	}
	return tiers
}
