package models

import "time"

// ============================================
// Order Today Item - Processed Orders with Tracking Info
// Matches Prisma schema: OrderTodayItem
// ============================================
type OrderTodayItem struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Platform      string    `gorm:"column:platform;index;not null" json:"platform"` // shopee, lazada, tiktok
	OrderSN       string    `gorm:"column:order_sn;index;not null" json:"order_sn"`
	TrackingNo    string    `gorm:"column:tracking_no" json:"tracking_no,omitempty"`
	Courier       string    `gorm:"column:courier" json:"courier,omitempty"`
	SellerSku     string    `gorm:"column:seller_sku" json:"seller_sku,omitempty"`
	ProductName   string    `gorm:"column:product_name" json:"product_name,omitempty"`
	VariationName string    `gorm:"column:variation_name" json:"variation_name,omitempty"`
	Quantity      int       `gorm:"column:quantity;default:1" json:"quantity"`
	SyncedAt      time.Time `gorm:"column:synced_at;index" json:"synced_at"`
	CreatedAt     time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (OrderTodayItem) TableName() string { return GetTableName("OrderTodayItem") }

// ============================================
// Locked Order - SKUs locked for processing
// Matches Prisma schema: LockedOrder
// ============================================
type LockedOrder struct {
	ID            string    `gorm:"primaryKey" json:"id"`
	TenantID      string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Sku           string    `gorm:"column:sku;index;not null" json:"sku"`
	ProductName   string    `gorm:"column:product_name" json:"product_name"`
	VariationName string    `gorm:"column:variation_name" json:"variation_name,omitempty"`
	Qty           int       `gorm:"column:qty" json:"qty"`
	CreatedAt     time.Time `gorm:"column:created_at;index" json:"created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (LockedOrder) TableName() string { return GetTableName("LockedOrder") }

// ============================================
// Platform Config - Per-tenant platform configuration
// Matches Prisma schema: PlatformConfig
// ============================================
type PlatformConfig struct {
	ID          string    `gorm:"primaryKey" json:"id"`
	Platform    string    `gorm:"column:platform;index;not null" json:"platform"`
	ConfigKey   string    `gorm:"column:config_key;index;not null" json:"config_key"`
	ConfigValue string    `gorm:"column:config_value" json:"config_value"`
	DataType    string    `gorm:"column:data_type;default:string" json:"data_type"`
	IsEncrypted bool      `gorm:"column:is_encrypted;default:true" json:"is_encrypted"`
	Metadata    string    `gorm:"column:metadata" json:"metadata,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (PlatformConfig) TableName() string { return GetTableName("PlatformConfig") }
