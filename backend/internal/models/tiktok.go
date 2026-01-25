package models

import "time"

// TiktokOrder represents a TikTok order in database
// Matches Prisma schema: TiktokOrder
type TiktokOrder struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       string    `gorm:"index;not null" json:"tenantId"`
	OrderSN        string    `gorm:"uniqueIndex;not null" json:"orderSn"`
	ShopID         *int64    `json:"shopId,omitempty"`
	OrderStatus    string    `json:"orderStatus"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

func (TiktokOrder) TableName() string {
	return GetTableName("TiktokOrder")
}

// TiktokOrderItem represents order items
// Matches Prisma schema: TiktokOrderItem
type TiktokOrderItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"index;not null" json:"tenantId"`
	OrderSN       string    `gorm:"index;not null" json:"orderSn"`
	LineItemID    string    `json:"lineItemId,omitempty"`
	ProductID     int64     `json:"productId"`
	SkuID         string    `json:"skuId,omitempty"`
	SellerSku     string    `json:"sellerSku,omitempty"`
	ProductName   string    `json:"productName,omitempty"`
	VariationName string    `json:"variationName,omitempty"`
	Quantity      *int      `json:"quantity,omitempty"`
	Price         *float64  `json:"price,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func (TiktokOrderItem) TableName() string {
	return GetTableName("TiktokOrderItem")
}

// TiktokProduct represents a TikTok product in database
// Matches Prisma schema: TiktokProduct
type TiktokProduct struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"index;not null" json:"tenantId"`
	ProductID   string    `gorm:"index;not null" json:"productId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Price       float64   `gorm:"default:0" json:"price"`
	Quantity    int       `gorm:"default:0" json:"quantity"`
	Image       string    `json:"image,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (TiktokProduct) TableName() string {
	return GetTableName("TiktokProduct")
}

// TiktokSku represents a TikTok SKU/variant
// Matches Prisma schema: TiktokSku
// NOTE: VariantData is JSONB in PostgreSQL, use serializer for proper handling
type TiktokSku struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	TenantID    string          `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	ProductID   uint            `gorm:"column:product_id;index;not null" json:"product_id"`
	SkuID       string          `gorm:"column:sku_id;uniqueIndex;not null" json:"sku_id"`
	SellerSku   string          `gorm:"column:seller_sku" json:"seller_sku,omitempty"`
	VariantName string          `gorm:"column:variant_name" json:"variant_name,omitempty"`
	VariantData JSONMap         `gorm:"column:variant_data;type:jsonb" json:"variant_data,omitempty"`
	Price       float64         `gorm:"column:price;default:0" json:"price"`
	Quantity    int             `gorm:"column:quantity;default:0" json:"quantity"`
	CreatedAt   time.Time       `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"column:updated_at" json:"updated_at"`
}

func (TiktokSku) TableName() string {
	return GetTableName("TiktokSku")
}

// ProductName returns the product name (alias for Name)
func (p *TiktokProduct) ProductName() string {
	return p.Name
}

// ProductStatus returns the product status (alias for Status)
func (p *TiktokProduct) ProductStatus() string {
	return p.Status
}
