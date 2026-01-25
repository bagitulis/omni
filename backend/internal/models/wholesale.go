package models

import "time"

// WholesaleSettings represents wholesale configuration for a tenant
type WholesaleSettings struct {
	ID        string    `gorm:"primaryKey;size:50" json:"id"`
	TenantID  string    `gorm:"index;not null" json:"tenant_id"`
	MinQty1   int       `json:"min_qty_1"`
	Discount1 float64   `json:"discount_1"` // Percentage discount for tier 1
	MinQty2   int       `json:"min_qty_2"`
	Discount2 float64   `json:"discount_2"` // Percentage discount for tier 2
	MinQty3   int       `json:"min_qty_3"`
	Discount3 float64   `json:"discount_3"` // Percentage discount for tier 3
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName returns the table name
func (WholesaleSettings) TableName() string {
	return "wholesale_settings"
}

// WholesaleTier represents a single wholesale tier
type WholesaleTier struct {
	MinQty   int     `json:"min_qty"`
	MaxQty   int     `json:"max_qty,omitempty"` // 0 means unlimited
	Discount float64 `json:"discount"`          // Percentage
	Price    float64 `json:"price"`             // Calculated price
}

// WholesaleCalculateRequest represents wholesale calculation request
type WholesaleCalculateRequest struct {
	OriginalPrice float64 `json:"original_price" binding:"required"`
	SKUs          []string `json:"skus,omitempty"` // Optional: specific SKUs
}

// WholesaleCalculateResponse represents calculation result
type WholesaleCalculateResponse struct {
	OriginalPrice float64         `json:"original_price"`
	Tiers         []WholesaleTier `json:"tiers"`
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
