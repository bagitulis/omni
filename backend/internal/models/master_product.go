package models

import "time"

// MasterProduct represents a unified product across all marketplace platforms
// This is the single source of truth for product data
type MasterProduct struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Title       string    `gorm:"column:title;size:120;not null" json:"title"`
	Description string    `gorm:"column:description;type:text" json:"description"`
	Images      JSONArray `gorm:"column:images;type:jsonb;default:'[]'" json:"images"`
	Status      string    `gorm:"column:status;size:50;default:'draft'" json:"status"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`

	// Relations (not stored in DB, for preloading)
	SKUs []MasterProductSku `gorm:"foreignKey:MasterProductID" json:"skus,omitempty"`
}

// TableName specifies the PostgreSQL table name
func (MasterProduct) TableName() string {
	return GetTableName("MasterProduct")
}

// MasterProductSku represents a SKU/variant of a master product
// Max 50 SKUs per product (Shopee limit), max 2 variant tiers
// PlatformPrice holds real marketplace price/stock from platform staging tables.
// Used in API response only (not stored in master DB).
type PlatformPrice struct {
	Platform string  `json:"platform"`       // "shopee", "tiktok", "lazada"
	Price    float64 `json:"platform_price"`
	Stock    int     `json:"platform_stock"`
}

type MasterProductSku struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	MasterProductID uint      `gorm:"column:master_product_id;index;not null" json:"master_product_id"`
	SellerSku       string    `gorm:"column:seller_sku;size:255;not null" json:"seller_sku"`
	VariantName     string    `gorm:"column:variant_name;size:255" json:"variant_name"`
	VariantData     JSONMap   `gorm:"column:variant_data;type:jsonb;default:'{}'" json:"variant_data"`
	Price           float64   `gorm:"column:price;type:decimal(15,2);default:0" json:"price"`
	Stock           int       `gorm:"column:stock;default:0" json:"stock"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`

	// Relations
	PlatformLinks []MasterProductPlatformLink `gorm:"foreignKey:MasterSkuID;references:ID" json:"platform_links,omitempty"`

	// Virtual fields — enriched at API response time, not stored in DB
	PlatformPrices []PlatformPrice `gorm:"-" json:"platform_prices,omitempty"`
	InventoryPrice float64         `gorm:"-" json:"inventory_price,omitempty"`
	InventoryStock int             `gorm:"-" json:"inventory_stock,omitempty"`
}

// TableName specifies the PostgreSQL table name
func (MasterProductSku) TableName() string {
	return GetTableName("MasterProductSku")
}

// MasterProductPlatformLink maps master product/SKU to platform-specific products
// Supports: shopee, tiktok, lazada
type MasterProductPlatformLink struct {
	ID                uint       `gorm:"primaryKey" json:"id"`
	MasterProductID   uint       `gorm:"column:master_product_id;index;not null" json:"master_product_id"`
	MasterSkuID       *uint      `gorm:"column:master_sku_id;index" json:"master_sku_id,omitempty"`
	Platform          string     `gorm:"column:platform;size:50;not null" json:"platform"`
	PlatformProductID string     `gorm:"column:platform_product_id;size:255" json:"platform_product_id"`
	PlatformItemID    string     `gorm:"column:platform_item_id;size:255" json:"platform_item_id"`
	PlatformSkuID     string     `gorm:"column:platform_sku_id;size:255" json:"platform_sku_id"`
	SyncStatus        string     `gorm:"column:sync_status;size:50;default:'pending'" json:"sync_status"`
	LastSyncedAt      *time.Time `gorm:"column:last_synced_at" json:"last_synced_at,omitempty"`
	CreatedAt         time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the PostgreSQL table name
func (MasterProductPlatformLink) TableName() string {
	return GetTableName("MasterProductPlatformLink")
}

// Constants for MasterProduct status
const (
	MasterProductStatusDraft    = "draft"
	MasterProductStatusActive   = "active"
	MasterProductStatusArchived = "archived"
)

// Constants for platform sync status
const (
	SyncStatusPending  = "pending"
	SyncStatusSynced   = "synced"
	SyncStatusError    = "error"
	SyncStatusOutdated = "outdated"
)

// NOTE: Platform constants (PlatformShopee, PlatformTiktok, PlatformLazada)
// are already defined in oauth.go - reuse those

// Field limits (Shopee is the most restrictive)
const (
	MasterProductMaxTitleLength       = 120  // Shopee limit
	MasterProductMaxDescriptionLength = 5000 // Shopee limit
	MasterProductMaxImages            = 8    // Lazada limit (strictest)
	MasterProductMaxSKUs              = 50   // Shopee limit
	MasterProductMaxVariantTiers      = 2    // Shopee limit
)
