package models

import "time"

// TiktokOrder represents a TikTok order in database
// Matches Prisma schema: TiktokOrder
// NOTE: JSON tags use snake_case for frontend compatibility
type TiktokOrder struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	OrderSN         string    `gorm:"column:order_sn;uniqueIndex;not null" json:"order_sn"`
	ShopID          *int64    `gorm:"column:shop_id" json:"shop_id,omitempty"`
	OrderStatus     string    `gorm:"column:order_status" json:"order_status"`
	TotalAmount     *float64  `gorm:"column:total_amount" json:"total_amount,omitempty"`
	Currency        string    `gorm:"column:currency;default:'IDR'" json:"currency,omitempty"`
	BuyerUsername   string    `gorm:"column:buyer_username" json:"buyer_username,omitempty"`
	PaymentMethod   string    `gorm:"column:payment_method" json:"payment_method,omitempty"`
	ShippingCarrier string    `gorm:"column:shipping_carrier" json:"shipping_carrier,omitempty"`
	BuyerMessage    string    `gorm:"column:buyer_message" json:"buyer_message,omitempty"`
	ShipByDate      *int64    `gorm:"column:ship_by_date" json:"ship_by_date,omitempty"`       // Unix timestamp deadline for shipping
	TrackingNumber  string    `gorm:"column:tracking_number" json:"tracking_number,omitempty"` // Shipping tracking number
	CreatedAt       time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (TiktokOrder) TableName() string {
	return GetTableName("TiktokOrder")
}

// TiktokOrderItem represents order items
// Matches Prisma schema: TiktokOrderItem
// NOTE: JSON tags use snake_case for frontend compatibility
type TiktokOrderItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	OrderSN       string    `gorm:"column:order_sn;index;not null" json:"order_sn"`
	LineItemID    string    `gorm:"column:line_item_id" json:"line_item_id,omitempty"`
	ProductID     int64     `gorm:"column:product_id" json:"product_id"`
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

func (TiktokOrderItem) TableName() string {
	return GetTableName("TiktokOrderItem")
}

// TiktokProduct represents a TikTok product in database
// Matches Prisma schema: TiktokProduct
// NOTE: JSON tags use snake_case for frontend compatibility
type TiktokProduct struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	ProductID   string    `gorm:"column:product_id;index;not null" json:"product_id"`
	Name        string    `gorm:"column:name" json:"name"`
	Description string    `gorm:"column:description" json:"description,omitempty"`
	Status      string    `gorm:"column:status" json:"status"`
	Price       float64   `gorm:"column:price;default:0" json:"price"`
	Quantity    int       `gorm:"column:quantity;default:0" json:"quantity"`
	Image       string    `gorm:"column:image" json:"image,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (TiktokProduct) TableName() string {
	return GetTableName("TiktokProduct")
}

// TiktokSku represents a TikTok SKU/variant
// Matches Prisma schema: TiktokSku
// NOTE: VariantData is JSONB in PostgreSQL, use serializer for proper handling
type TiktokSku struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	ProductID   uint      `gorm:"column:product_id;index;not null" json:"product_id"`
	SkuID       string    `gorm:"column:sku_id;uniqueIndex;not null" json:"sku_id"`
	SellerSku   string    `gorm:"column:seller_sku" json:"seller_sku,omitempty"`
	VariantName string    `gorm:"column:variant_name" json:"variant_name,omitempty"`
	VariantData JSONMap   `gorm:"column:variant_data;type:jsonb" json:"variant_data,omitempty"`
	Price       float64   `gorm:"column:price;default:0" json:"price"`
	Quantity    int       `gorm:"column:quantity;default:0" json:"quantity"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
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
