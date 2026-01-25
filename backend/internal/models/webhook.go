package models

import "time"

// WebhookLog logs all webhook events received
// Matches Prisma schema: WebhookLog
type WebhookLog struct {
	ID          string     `gorm:"primaryKey" json:"id"`
	TenantID    string     `gorm:"index" json:"tenantId,omitempty"`
	Platform    string     `gorm:"index;not null" json:"platform"`
	EventType   string     `gorm:"index;not null" json:"eventType"`
	Payload     string     `json:"payload"`           // JSON string
	Headers     string     `json:"headers,omitempty"` // JSON string
	Status      string     `gorm:"default:received;index" json:"status"`
	ErrorMsg    string     `json:"errorMsg,omitempty"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
	CreatedAt   time.Time  `gorm:"index" json:"createdAt"`
}

// TableName specifies the table name for GORM
func (WebhookLog) TableName() string {
	return GetTableName("WebhookLog")
}

// WebhookOrderEvent stores order-related webhook events
// Matches Prisma schema: WebhookOrderEvent
type WebhookOrderEvent struct {
	ID                int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID          string     `gorm:"index;not null" json:"tenantId"`
	Platform          string     `gorm:"index;not null" json:"platform"`
	EventType         string     `gorm:"index;not null" json:"eventType"`
	OrderSN           string     `gorm:"index;not null" json:"orderSn"`
	ShopID            string     `json:"shopId,omitempty"`
	OldStatus         string     `json:"oldStatus,omitempty"`
	NewStatus         string     `json:"newStatus,omitempty"`
	FulfillmentStatus string     `json:"fulfillmentStatus,omitempty"`
	PackageNumber     string     `json:"packageNumber,omitempty"`
	Payload           string     `json:"payload,omitempty"` // JSON string
	WebhookLogID      string     `json:"webhookLogId,omitempty"`
	ProcessedAt       *time.Time `json:"processedAt,omitempty"`
	CreatedAt         time.Time  `gorm:"index" json:"createdAt"`
}

// TableName specifies the table name for GORM
func (WebhookOrderEvent) TableName() string {
	return GetTableName("WebhookOrderEvent")
}

// WebhookProductEvent stores product-related webhook events
// Matches Prisma schema: WebhookProductEvent
type WebhookProductEvent struct {
	ID          int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string     `gorm:"index;not null" json:"tenantId"`
	EventType   string     `gorm:"index;not null" json:"eventType"`
	ShopID      string     `json:"shopId,omitempty"`
	ItemID      string     `gorm:"index" json:"itemId,omitempty"`
	VariationID string     `json:"variationId,omitempty"`
	Action      string     `json:"action,omitempty"`
	Payload     string     `json:"payload,omitempty"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// TableName specifies the table name for GORM
func (WebhookProductEvent) TableName() string {
	return GetTableName("WebhookProductEvent")
}

// WebhookReturnEvent stores return/refund webhook events
// Matches Prisma schema: WebhookReturnEvent
type WebhookReturnEvent struct {
	ID          int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID    string     `gorm:"index;not null" json:"tenantId"`
	EventType   string     `gorm:"index;not null" json:"eventType"`
	ShopID      string     `json:"shopId,omitempty"`
	OrderSN     string     `gorm:"index" json:"orderSn,omitempty"`
	ReturnSN    string     `gorm:"index" json:"returnSn,omitempty"`
	Status      string     `json:"status,omitempty"`
	Reason      string     `json:"reason,omitempty"`
	Payload     string     `json:"payload,omitempty"`
	ProcessedAt *time.Time `json:"processedAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
}

// TableName specifies the table name for GORM
func (WebhookReturnEvent) TableName() string {
	return GetTableName("WebhookReturnEvent")
}

// WebhookMarketingEvent stores marketing/promo webhook events
// Matches Prisma schema: WebhookMarketingEvent
type WebhookMarketingEvent struct {
	ID            int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      string     `gorm:"index;not null" json:"tenantId"`
	EventType     string     `gorm:"index;not null" json:"eventType"`
	ShopID        string     `json:"shopId,omitempty"`
	ItemID        string     `json:"itemId,omitempty"`
	PromotionID   string     `gorm:"index" json:"promotionId,omitempty"`
	PromotionType string     `json:"promotionType,omitempty"`
	Action        string     `json:"action,omitempty"`
	Payload       string     `json:"payload,omitempty"`
	ProcessedAt   *time.Time `json:"processedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func (WebhookMarketingEvent) TableName() string {
	return GetTableName("WebhookMarketingEvent")
}

// WebhookShopeeEvent stores Shopee system events
// Matches Prisma schema: WebhookShopeeEvent
type WebhookShopeeEvent struct {
	ID            int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      string     `gorm:"index;not null" json:"tenantId"`
	EventType     string     `gorm:"index;not null" json:"eventType"`
	ShopID        string     `gorm:"index" json:"shopId,omitempty"`
	Action        string     `json:"action,omitempty"`
	ExpiryTime    *time.Time `json:"expiryTime,omitempty"`
	PenaltyPoints *int       `json:"penaltyPoints,omitempty"`
	Payload       string     `json:"payload,omitempty"`
	ProcessedAt   *time.Time `json:"processedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
}

func (WebhookShopeeEvent) TableName() string {
	return GetTableName("WebhookShopeeEvent")
}

// WebhookWebchatEvent stores chat/message webhook events
// Matches Prisma schema: WebhookWebchatEvent
type WebhookWebchatEvent struct {
	ID             int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID       string     `gorm:"index;not null" json:"tenantId"`
	EventType      string     `gorm:"index;not null" json:"eventType"`
	ShopID         string     `json:"shopId,omitempty"`
	ConversationID string     `gorm:"index" json:"conversationId,omitempty"`
	MessageType    string     `json:"messageType,omitempty"`
	SenderID       string     `json:"senderId,omitempty"`
	Payload        string     `json:"payload,omitempty"`
	ProcessedAt    *time.Time `json:"processedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

func (WebhookWebchatEvent) TableName() string {
	return GetTableName("WebhookWebchatEvent")
}

// WebhookFBSEvent stores Fulfillment by Shopee events
// Matches Prisma schema: WebhookFBSEvent
type WebhookFBSEvent struct {
	ID            int        `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID      string     `gorm:"index;not null" json:"tenantId"`
	EventType     string     `gorm:"index;not null" json:"eventType"`
	ShopID        string     `json:"shopId,omitempty"`
	ItemID        string     `gorm:"index" json:"itemId,omitempty"`
	SkuID         string     `json:"skuId,omitempty"`
	StockChange   *int       `json:"stockChange,omitempty"`
	InvoiceNumber string     `json:"invoiceNumber,omitempty"`
	Payload       string     `json:"payload,omitempty"`
	ProcessedAt   *time.Time `json:"processedAt,omitempty"`
	CreatedAt     time.Time  `json:"createdAt"`
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
