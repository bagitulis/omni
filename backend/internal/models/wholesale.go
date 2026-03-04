package models

import "time"

// WholesaleSettings represents wholesale configuration for a tenant
// Schema matches Node.js backend (admin_fee-based pricing)
type WholesaleSettings struct {
	ID            string    `gorm:"primaryKey;size:50" json:"id"`
	TenantID      string    `gorm:"index;not null" json:"tenant_id"`
	Platform      string    `gorm:"default:shopee" json:"platform"`
	AdminFee      int       `json:"admin_fee"`
	MinOrder1     int       `json:"min_order_1"`
	MaxOrder1     int       `json:"max_order_1"`
	MaxOrderTier3 int       `json:"max_order_tier_3"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// TableName returns the table name
func (WholesaleSettings) TableName() string {
	return "wholesale_settings"
}

// WholesaleTier represents a single wholesale tier (for Shopee API format)
type WholesaleTier struct {
	MinCount  int     `json:"min_count"`
	MaxCount  int     `json:"max_count"`
	UnitPrice float64 `json:"unit_price"`
}

// WholesaleCalculateRequest represents wholesale calculation request
type WholesaleCalculateRequest struct {
	BasePrice float64 `json:"base_price" binding:"required"`
}

// WholesaleCalculateResponse represents calculation result
type WholesaleCalculateResponse struct {
	BasePrice float64         `json:"base_price"`
	AdminFee  int             `json:"admin_fee"`
	Tiers     []WholesaleTier `json:"tiers"`
}

// WholesaleApplyRequest represents request to apply wholesale to products
type WholesaleApplyRequest struct {
	Platform string   `json:"platform" binding:"required"` // shopee, lazada, tiktok
	SKUs     []string `json:"skus" binding:"required"`
}

// WholesaleApplyResult represents apply operation result
type WholesaleApplyResult struct {
	SKU     string `json:"sku"`
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
