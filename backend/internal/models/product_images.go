// Package models contains database models for the application
package models

// MasterProductImage links master products to images
type MasterProductImage struct {
	ProductID uint   `gorm:"primaryKey" json:"product_id"`
	ImageID   uint   `gorm:"primaryKey" json:"image_id"`
	SortOrder int    `gorm:"default:0" json:"sort_order"`
	Role      string `gorm:"size:50" json:"role"` // "main", "gallery", etc.

	// Foreign keys
	Product MasterProduct `gorm:"foreignKey:ProductID" json:"-"`
	Image   Image         `gorm:"foreignKey:ImageID" json:"-"`
}

// TableName specifies the PostgreSQL table name
func (MasterProductImage) TableName() string {
	return GetTableName("MasterProductImage")
}

// ShopeeProductImage links Shopee products to images
type ShopeeProductImage struct {
	ShopeeProductID uint `gorm:"primaryKey;column:shopee_product_id" json:"shopee_product_id"`
	ImageID         uint `gorm:"primaryKey" json:"image_id"`
	SortOrder       int  `gorm:"default:0" json:"sort_order"`

	// Foreign keys
	ShopeeProduct ShopeeProduct `gorm:"foreignKey:ShopeeProductID" json:"-"`
	Image         Image         `gorm:"foreignKey:ImageID" json:"-"`
}

// TableName specifies the PostgreSQL table name
func (ShopeeProductImage) TableName() string {
	return GetTableName("ShopeeProductImage")
}

// TiktokProductImage links TikTok products to images
type TiktokProductImage struct {
	TiktokProductID uint `gorm:"primaryKey;column:tiktok_product_id" json:"tiktok_product_id"`
	ImageID         uint `gorm:"primaryKey" json:"image_id"`
	SortOrder       int  `gorm:"default:0" json:"sort_order"`

	// Foreign keys
	TiktokProduct TiktokProduct `gorm:"foreignKey:TiktokProductID" json:"-"`
	Image         Image         `gorm:"foreignKey:ImageID" json:"-"`
}

// TableName specifies the PostgreSQL table name
func (TiktokProductImage) TableName() string {
	return GetTableName("TiktokProductImage")
}

// LazadaProductImage links Lazada products to images
type LazadaProductImage struct {
	LazadaProductID uint `gorm:"primaryKey;column:lazada_product_id" json:"lazada_product_id"`
	ImageID         uint `gorm:"primaryKey" json:"image_id"`
	SortOrder       int  `gorm:"default:0" json:"sort_order"`

	// Foreign keys
	LazadaProduct LazadaProduct `gorm:"foreignKey:LazadaProductID" json:"-"`
	Image         Image         `gorm:"foreignKey:ImageID" json:"-"`
}

// TableName specifies the PostgreSQL table name
func (LazadaProductImage) TableName() string {
	return GetTableName("LazadaProductImage")
}
