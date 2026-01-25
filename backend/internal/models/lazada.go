package models

import "time"

// LazadaOrder represents a Lazada order in database
// Matches Prisma schema: LazadaOrder
type LazadaOrder struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"index;not null" json:"tenantId"`
	OrderSN       string    `gorm:"uniqueIndex;not null" json:"orderSn"`
	ShopID        *int64    `json:"shopId,omitempty"`
	OrderStatus   string    `json:"orderStatus"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (LazadaOrder) TableName() string {
	return GetTableName("LazadaOrder")
}

// LazadaOrderItem represents order items
// Matches Prisma schema: LazadaOrderItem
type LazadaOrderItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"index;not null" json:"tenantId"`
	OrderSN       string    `gorm:"index;not null" json:"orderSn"`
	ItemID        int64     `json:"itemId"`
	SkuID         string    `json:"skuId,omitempty"`
	SellerSku     string    `json:"sellerSku,omitempty"`
	ProductName   string    `json:"productName,omitempty"`
	VariationName string    `json:"variationName,omitempty"`
	Quantity      *int      `json:"quantity,omitempty"`
	Price         *float64  `json:"price,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (LazadaOrderItem) TableName() string {
	return GetTableName("LazadaOrderItem")
}

// LazadaProduct represents a Lazada product in database
// Matches Prisma schema: LazadaProduct
type LazadaProduct struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"index;not null" json:"tenant_id"`
	ItemID      string    `gorm:"index;not null" json:"item_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Brand       string    `json:"brand,omitempty"`
	Status      string    `json:"status"`
	Price       float64   `gorm:"default:0" json:"price"`
	Quantity    int       `gorm:"default:0" json:"quantity"`
	Image       string    `json:"image,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (LazadaProduct) TableName() string {
	return GetTableName("LazadaProduct")
}

// LazadaSku represents a Lazada SKU/variant
// Matches Prisma schema: LazadaSku
// NOTE: VariantData is JSONB in PostgreSQL, use JSONMap for proper handling
type LazadaSku struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	ItemID       string    `gorm:"column:item_id;index;not null" json:"item_id"`
	ProductID    uint      `gorm:"column:product_id;index" json:"product_id"`
	SkuID        string    `gorm:"column:sku_id;uniqueIndex;not null" json:"sku_id"`
	ShopSku      string    `gorm:"column:shop_sku" json:"shop_sku,omitempty"`
	SellerSku    string    `gorm:"column:seller_sku" json:"seller_sku,omitempty"`
	Name         string    `gorm:"column:name" json:"name,omitempty"`
	VariantName  string    `gorm:"column:variant_name" json:"variant_name,omitempty"`
	Price        float64   `gorm:"column:price;default:0" json:"price"`
	SpecialPrice float64   `gorm:"column:special_price;default:0" json:"special_price"`
	Quantity     int       `gorm:"column:quantity;default:0" json:"quantity"`
	Available    int       `gorm:"column:available;default:0" json:"available"`
	VariantData  JSONMap   `gorm:"column:variant_data;type:jsonb" json:"variant_data,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (LazadaSku) TableName() string {
	return GetTableName("LazadaSku")
}
