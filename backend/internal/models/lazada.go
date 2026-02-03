package models

import "time"

// LazadaOrder represents a Lazada order in database
// Matches Prisma schema: LazadaOrder
// NOTE: JSON tags use snake_case for frontend compatibility
type LazadaOrder struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	OrderSN         string    `gorm:"column:order_sn;uniqueIndex;not null" json:"order_sn"`
	ShopID          *int64    `gorm:"column:shop_id" json:"shop_id,omitempty"`
	OrderStatus     string    `gorm:"column:order_status" json:"order_status"`
	TotalAmount     *float64  `gorm:"column:total_amount" json:"total_amount,omitempty"`
	Currency        string    `gorm:"column:currency;default:'IDR'" json:"currency,omitempty"`
	BuyerUsername   string    `gorm:"column:buyer_username" json:"buyer_username,omitempty"`
	PaymentMethod   string    `gorm:"column:payment_method" json:"payment_method,omitempty"`
	ShippingCarrier string    `gorm:"column:shipping_carrier" json:"shipping_carrier"`
	BuyerMessage    string    `gorm:"column:buyer_message" json:"buyer_message,omitempty"`
	ShipByDate      *int64    `gorm:"column:ship_by_date" json:"ship_by_date"` // Unix timestamp deadline for shipping
	Countdown       *int64    `gorm:"-" json:"countdown"`
	TrackingNumber  string    `gorm:"column:tracking_number" json:"tracking_number,omitempty"` // Shipping tracking number
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (LazadaOrder) TableName() string {
	return GetTableName("LazadaOrder")
}

// LazadaOrderItem represents order items
// Matches Prisma schema: LazadaOrderItem
// NOTE: JSON tags use snake_case for frontend compatibility
type LazadaOrderItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	OrderSN       string    `gorm:"column:order_sn;index;not null" json:"order_sn"`
	ItemID        int64     `gorm:"column:item_id" json:"item_id"`
	SkuID         string    `gorm:"column:sku_id" json:"sku_id,omitempty"`
	SellerSku     string    `gorm:"column:seller_sku" json:"seller_sku,omitempty"`
	ProductName   string    `gorm:"column:product_name" json:"product_name,omitempty"`
	VariationName string    `gorm:"column:variation_name" json:"variation_name,omitempty"`
	Quantity      *int      `gorm:"column:quantity" json:"quantity,omitempty"`
	Price         *float64  `gorm:"column:price" json:"price,omitempty"`
	ProductImage  string    `gorm:"column:product_image" json:"product_image,omitempty"` // Product image URL
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
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
	LocalImages JSONMap   `gorm:"column:local_images;type:jsonb" json:"local_images"` // Local image paths after download/conversion
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
