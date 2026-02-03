package models

import "time"

// ShopeeOrder represents a Shopee order in database
// Matches Prisma schema: ShopeeOrder
// NOTE: Only fields that exist in PostgreSQL database (from Prisma migration)
type ShopeeOrder struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	OrderSN         string    `gorm:"column:order_sn;uniqueIndex;not null" json:"order_sn"`
	ShopID          *int64    `gorm:"column:shop_id" json:"shop_id,omitempty"`
	OrderStatus     string    `gorm:"column:order_status" json:"order_status"`
	OrderTimestamp  *int      `gorm:"column:order_timestamp" json:"order_timestamp,omitempty"`
	TotalAmount     *float64  `gorm:"column:total_amount" json:"total_amount,omitempty"`
	Currency        string    `gorm:"column:currency;default:'IDR'" json:"currency,omitempty"`
	BuyerUsername   string    `gorm:"column:buyer_username" json:"buyer_username,omitempty"`
	PaymentMethod   string    `gorm:"column:payment_method" json:"payment_method,omitempty"`
	ShippingCarrier string    `gorm:"column:shipping_carrier" json:"shipping_carrier"`
	BuyerMessage    string    `gorm:"column:buyer_message" json:"buyer_message,omitempty"`
	ShipByDate      *int64    `gorm:"column:ship_by_date" json:"ship_by_date"`
	Countdown       *int64    `gorm:"-" json:"countdown"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (ShopeeOrder) TableName() string {
	return GetTableName("ShopeeOrder")
}

// ShopeeProduct represents a Shopee product in database
// Matches PostgreSQL schema: shopee_products
type ShopeeProduct struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	ItemID      int64     `gorm:"column:item_id;index;not null" json:"item_id"`
	Name        string    `gorm:"column:name" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Status      string    `gorm:"column:status" json:"status"`
	Price       float64   `gorm:"column:price" json:"price"`
	Quantity    int       `gorm:"column:quantity" json:"quantity"`
	Image       string    `gorm:"column:image" json:"image"`
	LocalImages JSONMap   `gorm:"column:local_images;type:jsonb" json:"local_images"` // Local image paths after download/conversion
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ShopeeProduct) TableName() string {
	return GetTableName("ShopeeProduct")
}

// ShopeeSku represents a Shopee SKU/variant
// Matches PostgreSQL schema: shopee_skus
// NOTE: VariantData is JSONB in PostgreSQL, use JSONMap for proper handling
type ShopeeSku struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	ProductID   uint      `gorm:"column:product_id;index;not null" json:"product_id"`
	ItemID      int64     `gorm:"column:item_id;index" json:"item_id"`
	ModelID     *int64    `gorm:"column:model_id" json:"model_id,omitempty"`
	SellerSku   string    `gorm:"column:seller_sku" json:"seller_sku,omitempty"`
	VariantName string    `gorm:"column:variant_name" json:"variant_name,omitempty"`
	VariantData JSONMap   `gorm:"column:variant_data;type:jsonb" json:"variant_data,omitempty"`
	Price       float64   `gorm:"column:price;default:0" json:"price"`
	Quantity    int       `gorm:"column:quantity;default:0" json:"quantity"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ShopeeSku) TableName() string {
	return GetTableName("ShopeeSku")
}

// ShopeeOrderItem represents order items
// Matches Prisma schema: ShopeeOrderItem + tenant_id (added in database)
type ShopeeOrderItem struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	OrderSN      string    `gorm:"column:order_sn;index;not null" json:"order_sn"`
	ItemID       int64     `gorm:"column:item_id" json:"item_id"`
	ModelID      *int64    `gorm:"column:model_id" json:"model_id,omitempty"`
	ItemName     string    `gorm:"column:item_name" json:"item_name,omitempty"`
	ModelName    string    `gorm:"column:model_name" json:"model_name,omitempty"`
	ItemSku      string    `gorm:"column:item_sku" json:"item_sku,omitempty"`
	ModelSku     string    `gorm:"column:model_sku" json:"model_sku,omitempty"`
	Quantity     *int      `gorm:"column:quantity" json:"quantity,omitempty"`
	Price        *float64  `gorm:"column:price" json:"price,omitempty"`
	ProductImage string    `gorm:"column:product_image" json:"product_image,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ShopeeOrderItem) TableName() string {
	return GetTableName("ShopeeOrderItem")
}
