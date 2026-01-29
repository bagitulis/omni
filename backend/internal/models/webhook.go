package models

import "time"

// WebhookLog logs all webhook events received
// Matches Prisma schema: WebhookLog
type WebhookLog struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	TenantID    string     `gorm:"index" json:"tenant_id,omitempty"`
	Platform    string     `gorm:"index;not null" json:"platform"`
	EventType   string     `gorm:"index;not null" json:"event_type"`
	Payload     string     `json:"payload"`           // JSON string
	Headers     string     `json:"headers,omitempty"` // JSON string
	Status      string     `gorm:"default:received;index" json:"status"`
	ErrorMsg    string     `json:"error_msg,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for GORM
func (WebhookLog) TableName() string {
	return GetTableName("WebhookLog")
}

// WebhookOrderEvent stores order-related webhook events
// Matches Prisma schema: WebhookOrderEvent
type WebhookOrderEvent struct {
	ID                int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          string     `gorm:"index;not null" json:"tenant_id"`
	Platform          string     `gorm:"index;not null" json:"platform"`
	EventType         string     `gorm:"index;not null" json:"event_type"`
	OrderSN           string     `gorm:"index;not null" json:"order_sn"`
	ShopID            string     `json:"shop_id,omitempty"`
	OldStatus         string     `json:"old_status,omitempty"`
	NewStatus         string     `json:"new_status,omitempty"`
	FulfillmentStatus string     `json:"fulfillment_status,omitempty"`
	PackageNumber     string     `json:"package_number,omitempty"`
	Payload           string     `json:"payload,omitempty"` // JSON string
	WebhookLogID      string     `json:"webhook_log_id,omitempty"`
	ProcessedAt       *time.Time `json:"processed_at,omitempty"`
	CreatedAt         time.Time  `gorm:"index" json:"created_at"`
}

// TableName specifies the table name for GORM
func (WebhookOrderEvent) TableName() string {
	return GetTableName("WebhookOrderEvent")
}

// WebhookProductEvent stores product-related webhook events
// Matches Prisma schema: WebhookProductEvent
type WebhookProductEvent struct {
	ID          int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string     `gorm:"index;not null" json:"tenant_id"`
	EventType   string     `gorm:"index;not null" json:"event_type"`
	ShopID      string     `json:"shop_id,omitempty"`
	ItemID      string     `gorm:"index" json:"item_id,omitempty"`
	VariationID string     `json:"variation_id,omitempty"`
	Action      string     `json:"action,omitempty"`
	Payload     string     `json:"payload,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName specifies the table name for GORM
func (WebhookProductEvent) TableName() string {
	return GetTableName("WebhookProductEvent")
}

// WebhookReturnEvent stores return/refund webhook events
// Matches Prisma schema: WebhookReturnEvent
type WebhookReturnEvent struct {
	ID          int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string     `gorm:"index;not null" json:"tenant_id"`
	EventType   string     `gorm:"index;not null" json:"event_type"`
	ShopID      string     `json:"shop_id,omitempty"`
	OrderSN     string     `gorm:"index" json:"order_sn,omitempty"`
	ReturnSN    string     `gorm:"index" json:"return_sn,omitempty"`
	Status      string     `json:"status,omitempty"`
	Reason      string     `json:"reason,omitempty"`
	Payload     string     `json:"payload,omitempty"`
	ProcessedAt *time.Time `json:"processed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName specifies the table name for GORM
func (WebhookReturnEvent) TableName() string {
	return GetTableName("WebhookReturnEvent")
}

// WebhookMarketingEvent stores marketing/promo webhook events
// Matches Prisma schema: WebhookMarketingEvent
type WebhookMarketingEvent struct {
	ID            int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      string     `gorm:"index;not null" json:"tenant_id"`
	EventType     string     `gorm:"index;not null" json:"event_type"`
	ShopID        string     `json:"shop_id,omitempty"`
	ItemID        string     `json:"item_id,omitempty"`
	PromotionID   string     `gorm:"index" json:"promotion_id,omitempty"`
	PromotionType string     `json:"promotion_type,omitempty"`
	Action        string     `json:"action,omitempty"`
	Payload       string     `json:"payload,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (WebhookMarketingEvent) TableName() string {
	return GetTableName("WebhookMarketingEvent")
}

// WebhookShopeeEvent stores Shopee system events
// Matches Prisma schema: WebhookShopeeEvent
type WebhookShopeeEvent struct {
	ID            int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      string     `gorm:"index;not null" json:"tenant_id"`
	EventType     string     `gorm:"index;not null" json:"event_type"`
	ShopID        string     `gorm:"index" json:"shop_id,omitempty"`
	Action        string     `json:"action,omitempty"`
	ExpiryTime    *time.Time `json:"expiry_time,omitempty"`
	PenaltyPoints *int       `json:"penalty_points,omitempty"`
	Payload       string     `json:"payload,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (WebhookShopeeEvent) TableName() string {
	return GetTableName("WebhookShopeeEvent")
}

// WebhookWebchatEvent stores chat/message webhook events
// Matches Prisma schema: WebhookWebchatEvent
type WebhookWebchatEvent struct {
	ID             int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       string     `gorm:"index;not null" json:"tenant_id"`
	EventType      string     `gorm:"index;not null" json:"event_type"`
	ShopID         string     `json:"shop_id,omitempty"`
	ConversationID string     `gorm:"index" json:"conversation_id,omitempty"`
	MessageType    string     `json:"message_type,omitempty"`
	SenderID       string     `json:"sender_id,omitempty"`
	Payload        string     `json:"payload,omitempty"`
	ProcessedAt    *time.Time `json:"processed_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (WebhookWebchatEvent) TableName() string {
	return GetTableName("WebhookWebchatEvent")
}

// WebhookFBSEvent stores Fulfillment by Shopee events
// Matches Prisma schema: WebhookFBSEvent
type WebhookFBSEvent struct {
	ID            int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      string     `gorm:"index;not null" json:"tenant_id"`
	EventType     string     `gorm:"index;not null" json:"event_type"`
	ShopID        string     `json:"shop_id,omitempty"`
	ItemID        string     `gorm:"index" json:"item_id,omitempty"`
	SkuID         string     `json:"sku_id,omitempty"`
	StockChange   *int       `json:"stock_change,omitempty"`
	InvoiceNumber string     `json:"invoice_number,omitempty"`
	Payload       string     `json:"payload,omitempty"`
	ProcessedAt   *time.Time `json:"processed_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (WebhookFBSEvent) TableName() string {
	return GetTableName("WebhookFBSEvent")
}

// WebhookStatus constants
const (
	WebhookStatusReceived  = "received"
	WebhookStatusProcessed = "processed"
	WebhookStatusFailed    = "failed"
)

// Shopee push code categories
const (
	ShopeePushShopAuth     = 0 // Shop authorization
	ShopeePushOrderStatus  = 1 // Order status update
	ShopeePushTracking     = 2 // Tracking number update
	ShopeePushShopeeUpdate = 3 // Item add/update/delete
	ShopeePushPromotion    = 4 // Item promotion
	ShopeePushShopUpdate   = 6 // Shop update
	ShopeePushWebchat      = 7 // Chat message
)
