package models

import (
	"time"
)

// Product represents a unified product across all platforms
type Product struct {
	ID          string    `gorm:"primaryKey;type:varchar(36)" json:"id"`
	TenantID    string    `gorm:"index;not null;type:varchar(100)" json:"tenantId"`
	Platform    string    `gorm:"index;not null;type:varchar(20)" json:"platform"` // shopee, lazada, tiktok
	ItemID      string    `gorm:"index;not null;type:varchar(100)" json:"itemId"`
	SKU         string    `gorm:"index;type:varchar(100)" json:"sku"`
	Name        string    `gorm:"type:varchar(500)" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Status      string    `gorm:"type:varchar(20);default:NORMAL" json:"status"` // NORMAL, DELETED, BANNED
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	ImageURL    string    `gorm:"type:varchar(1000)" json:"imageUrl,omitempty"`
	CategoryID  string    `gorm:"type:varchar(100)" json:"categoryId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
	
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
	ProductID     string  `gorm:"index;not null;type:varchar(36)" json:"productId"`
	TenantID      string  `gorm:"index;not null;type:varchar(100)" json:"tenantId"`
	ModelID       string  `gorm:"index;type:varchar(100)" json:"modelId"`
	SKU           string  `gorm:"index;not null;type:varchar(100)" json:"sku"`
	Name          string  `gorm:"type:varchar(500)" json:"name"`
	VariationName string  `gorm:"type:varchar(500)" json:"variationName,omitempty"`
	Price         float64 `json:"price"`
	Stock         int     `json:"stock"`
}

// TableName returns the table name for GORM
func (ProductSKU) TableName() string {
	return GetTableName("ProductSKU")
}
