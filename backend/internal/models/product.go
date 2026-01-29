package models

import (
	"time"
)

// Product represents a unified product across all platforms
type Product struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID    string    `gorm:"index;not null;type:varchar(100)" json:"tenant_id"`
	Platform    string    `gorm:"index;not null;type:varchar(20)" json:"platform"` // shopee, lazada, tiktok
	ItemID      string    `gorm:"index;not null;type:varchar(100)" json:"item_id"`
	SKU         string    `gorm:"index;type:varchar(100)" json:"sku"`
	Name        string    `gorm:"type:varchar(500)" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Status      string    `gorm:"type:varchar(20);default:NORMAL" json:"status"` // NORMAL, DELETED, BANNED
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `gorm:"type:varchar(1000)" json:"image_url,omitempty"`
	CategoryID  string    `gorm:"type:varchar(100)" json:"category_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Relations
	SKUs []ProductSKU `gorm:"foreignKey:ProductID" json:"skus,omitempty"`
}

// TableName returns the table name for GORM
func (Product) TableName() string {
	return GetTableName("Product")
}

// ProductSKU represents a product SKU/variation
type ProductSKU struct {
	ID            string  `gorm:"primaryKey;type:varchar(36)" json:"id"`
	ProductID     string  `gorm:"index;not null;type:varchar(36)" json:"product_id"`
	TenantID      string  `gorm:"index;not null;type:varchar(100)" json:"tenant_id"`
	ModelID       string  `gorm:"index;type:varchar(100)" json:"model_id"`
	SKU           string  `gorm:"index;not null;type:varchar(100)" json:"sku"`
	Name          string  `gorm:"type:varchar(500)" json:"name"`
	VariationName string  `gorm:"type:varchar(500)" json:"variation_name,omitempty"`
	Price         float64 `json:"price"`
	Stock         int     `json:"stock"`
}

// TableName returns the table name for GORM
func (ProductSKU) TableName() string {
	return GetTableName("ProductSKU")
}
