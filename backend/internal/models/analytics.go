package models

import "time"

// AnalyticsSettings stores analytics configuration per tenant
// Matches Prisma schema: AnalyticsSettings
type AnalyticsSettings struct {
	ID                string    `gorm:"primaryKey" json:"id"`
	TenantID          string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Platform          string    `gorm:"column:platform;default:shopee" json:"platform"`
	PriceColumn       string    `gorm:"column:price_column;default:HARGA" json:"price_column"`
	FormulaDeduction  float64   `gorm:"column:formula_deduction;default:1500" json:"formula_deduction"`
	FormulaMultiplier float64   `gorm:"column:formula_multiplier;default:0.84" json:"formula_multiplier"`
	CreatedAt         time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (AnalyticsSettings) TableName() string {
	return GetTableName("AnalyticsSettings")
}

// ShopeeEscrowSync tracks Shopee escrow sync status
// Matches Prisma schema: ShopeeEscrowSync
type ShopeeEscrowSync struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	TenantID       string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Month          int       `gorm:"column:month;index" json:"month"`
	Year           int       `gorm:"column:year;index" json:"year"`
	TotalOrders    int       `gorm:"column:total_orders;default:0" json:"total_orders"`
	FailedOrders   int       `gorm:"column:failed_orders;default:0" json:"failed_orders"`
	FailedOrderIDs *string   `gorm:"column:failed_order_ids;type:text" json:"failed_order_ids,omitempty"`
	SyncedAt       time.Time `gorm:"column:synced_at" json:"synced_at"`
}

// TableName specifies the table name for GORM
func (ShopeeEscrowSync) TableName() string {
	return GetTableName("ShopeeEscrowSync")
}

// TiktokEscrowSync tracks TikTok escrow sync status
// Matches Prisma schema: TiktokEscrowSync
type TiktokEscrowSync struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	TenantID       string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Month          int       `gorm:"column:month;index" json:"month"`
	Year           int       `gorm:"column:year;index" json:"year"`
	TotalOrders    int       `gorm:"column:total_orders;default:0" json:"total_orders"`
	FailedOrders   int       `gorm:"column:failed_orders;default:0" json:"failed_orders"`
	FailedOrderIDs *string   `gorm:"column:failed_order_ids;type:text" json:"failed_order_ids,omitempty"`
	SyncedAt       time.Time `gorm:"column:synced_at" json:"synced_at"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the table name for GORM
func (TiktokEscrowSync) TableName() string {
	return GetTableName("TiktokEscrowSync")
}

// AnalyticsFilter represents filter for analytics queries
type AnalyticsFilter struct {
	TenantID  string
	Platform  string
	StartDate time.Time
	EndDate   time.Time
	GroupBy   string // daily, weekly, monthly
}

// ShopeeEscrowOrder represents Shopee escrow order data
// Matches Prisma schema: ShopeeEscrowOrder
type ShopeeEscrowOrder struct {
	ID                   string             `gorm:"primaryKey" json:"id"`
	TenantID             string             `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	OrderSN              string             `gorm:"column:order_sn;index;not null" json:"order_sn"`
	Month                int                `gorm:"column:month;index" json:"month"`
	Year                 int                `gorm:"column:year;index" json:"year"`
	OrderDate            *time.Time         `gorm:"column:order_date" json:"order_date,omitempty"`
	BuyerUserName        *string            `gorm:"column:buyer_user_name" json:"buyer_user_name,omitempty"`
	EscrowAmount         float64            `gorm:"column:escrow_amount;default:0" json:"escrow_amount"`
	CommissionFee        float64            `gorm:"column:commission_fee;default:0" json:"commission_fee"`
	ServiceFee           float64            `gorm:"column:service_fee;default:0" json:"service_fee"`
	SellerProcessingFee  float64            `gorm:"column:seller_processing_fee;default:0" json:"seller_processing_fee"`
	BuyerPaidShippingFee float64            `gorm:"column:buyer_paid_shipping_fee;default:0" json:"buyer_paid_shipping_fee"`
	ActualShippingFee    float64            `gorm:"column:actual_shipping_fee;default:0" json:"actual_shipping_fee"`
	ShopeeShippingRebate float64            `gorm:"column:shopee_shipping_rebate;default:0" json:"shopee_shipping_rebate"`
	EstimatedShippingFee float64            `gorm:"column:estimated_shipping_fee;default:0" json:"estimated_shipping_fee"`
	BuyerTotalAmount     float64            `gorm:"column:buyer_total_amount;default:0" json:"buyer_total_amount"`
	BuyerPaymentMethod   *string            `gorm:"column:buyer_payment_method" json:"buyer_payment_method,omitempty"`
	RawOrderIncome       *string            `gorm:"column:raw_order_income" json:"raw_order_income,omitempty"`
	RawBuyerPaymentInfo  *string            `gorm:"column:raw_buyer_payment_info" json:"raw_buyer_payment_info,omitempty"`
	SyncedAt             time.Time          `gorm:"column:synced_at" json:"synced_at"`
	CreatedAt            time.Time          `gorm:"column:created_at" json:"created_at"`
	UpdatedAt            time.Time          `gorm:"column:updated_at" json:"updated_at"`
	Items                []ShopeeEscrowItem `gorm:"foreignKey:EscrowOrderID" json:"items,omitempty"`
}

func (ShopeeEscrowOrder) TableName() string { return GetTableName("ShopeeEscrowOrder") }

// ShopeeEscrowItem represents Shopee escrow item data
// Matches Prisma schema: ShopeeEscrowItem
type ShopeeEscrowItem struct {
	ID                        string    `gorm:"primaryKey" json:"id"`
	TenantID                  string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	EscrowOrderID             string    `gorm:"column:escrow_order_id;index;not null" json:"escrow_order_id"`
	OrderID                   *string   `gorm:"column:order_id" json:"order_id,omitempty"`
	OrderSN                   *string   `gorm:"column:order_sn" json:"order_sn,omitempty"`
	Month                     *int      `gorm:"column:month" json:"month,omitempty"`
	Year                      *int      `gorm:"column:year" json:"year,omitempty"`
	ItemID                    *int64    `gorm:"column:item_id" json:"item_id,omitempty"`
	ModelID                   *int64    `gorm:"column:model_id" json:"model_id,omitempty"`
	Sku                       *string   `gorm:"column:sku" json:"sku,omitempty"`
	ModelSku                  *string   `gorm:"column:model_sku;index" json:"model_sku,omitempty"`
	ItemName                  *string   `gorm:"column:item_name" json:"item_name,omitempty"`
	ModelName                 *string   `gorm:"column:model_name" json:"model_name,omitempty"`
	Quantity                  int       `gorm:"column:quantity;default:0" json:"quantity"`
	OriginalPrice             float64   `gorm:"column:original_price;default:0" json:"original_price"`
	SellingPrice              float64   `gorm:"column:selling_price;default:0" json:"selling_price"`
	DiscountedPrice           float64   `gorm:"column:discounted_price;default:0" json:"discounted_price"`
	SellerDiscount            float64   `gorm:"column:seller_discount;default:0" json:"seller_discount"`
	ShopeeDiscount            float64   `gorm:"column:shopee_discount;default:0" json:"shopee_discount"`
	DiscountFromCoin          float64   `gorm:"column:discount_from_coin;default:0" json:"discount_from_coin"`
	DiscountFromVoucherSeller float64   `gorm:"column:discount_from_voucher_seller;default:0" json:"discount_from_voucher_seller"`
	DiscountFromVoucherShopee float64   `gorm:"column:discount_from_voucher_shopee;default:0" json:"discount_from_voucher_shopee"`
	AmsCommissionFee          float64   `gorm:"column:ams_commission_fee;default:0" json:"ams_commission_fee"`
	SellerOrderProcessingFee  float64   `gorm:"column:seller_order_processing_fee;default:0" json:"seller_order_processing_fee"`
	RawItemData               *string   `gorm:"column:raw_item_data" json:"raw_item_data,omitempty"`
	CreatedAt                 time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                 time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (ShopeeEscrowItem) TableName() string { return GetTableName("ShopeeEscrowItem") }

// TiktokEscrowOrder represents TikTok escrow order data
// Matches Prisma schema: TiktokEscrowOrder
type TiktokEscrowOrder struct {
	ID                          string             `gorm:"primaryKey" json:"id"`
	TenantID                    string             `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	OrderID                     string             `gorm:"column:order_id;index;not null" json:"order_id"`
	Month                       int                `gorm:"column:month;index" json:"month"`
	Year                        int                `gorm:"column:year;index" json:"year"`
	OrderStatus                 *string            `gorm:"column:order_status" json:"order_status,omitempty"`
	OrderDate                   *time.Time         `gorm:"column:order_date" json:"order_date,omitempty"`
	BuyerName                   *string            `gorm:"column:buyer_name" json:"buyer_name,omitempty"`
	TransactionID               *string            `gorm:"column:transaction_id" json:"transaction_id,omitempty"`
	TransactionType             *string            `gorm:"column:transaction_type" json:"transaction_type,omitempty"`
	StatementTime               *time.Time         `gorm:"column:statement_time" json:"statement_time,omitempty"`
	TotalSettlementAmount       float64            `gorm:"column:total_settlement_amount;default:0" json:"total_settlement_amount"`
	ProductRevenue              float64            `gorm:"column:product_revenue;default:0" json:"product_revenue"`
	PlatformCommission          float64            `gorm:"column:platform_commission;default:0" json:"platform_commission"`
	TransactionFee              float64            `gorm:"column:transaction_fee;default:0" json:"transaction_fee"`
	ShippingFeeCustomerPaid     float64            `gorm:"column:shipping_fee_customer_paid;default:0" json:"shipping_fee_customer_paid"`
	ShippingFeeActual           float64            `gorm:"column:shipping_fee_actual;default:0" json:"shipping_fee_actual"`
	ShippingFeePlatformDiscount float64            `gorm:"column:shipping_fee_platform_discount;default:0" json:"shipping_fee_platform_discount"`
	SellerShippingDiscount      float64            `gorm:"column:seller_shipping_discount;default:0" json:"seller_shipping_discount"`
	RefundAmount                float64            `gorm:"column:refund_amount;default:0" json:"refund_amount"`
	Adjustment                  float64            `gorm:"column:adjustment;default:0" json:"adjustment"`
	BuyerTotalAmount            float64            `gorm:"column:buyer_total_amount;default:0" json:"buyer_total_amount"`
	Currency                    string             `gorm:"column:currency;default:IDR" json:"currency"`
	RawTransactionData          *string            `gorm:"column:raw_transaction_data" json:"raw_transaction_data,omitempty"`
	RawOrderData                *string            `gorm:"column:raw_order_data" json:"raw_order_data,omitempty"`
	SyncedAt                    time.Time          `gorm:"column:synced_at" json:"synced_at"`
	CreatedAt                   time.Time          `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                   time.Time          `gorm:"column:updated_at" json:"updated_at"`
	Items                       []TiktokEscrowItem `gorm:"foreignKey:EscrowOrderID" json:"items,omitempty"`
}

func (TiktokEscrowOrder) TableName() string { return GetTableName("TiktokEscrowOrder") }

// TiktokEscrowItem represents TikTok escrow item data
// Matches Prisma schema: TiktokEscrowItem
type TiktokEscrowItem struct {
	ID                          string    `gorm:"primaryKey" json:"id"`
	TenantID                    string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	EscrowOrderID               string    `gorm:"column:escrow_order_id;index;not null" json:"escrow_order_id"`
	OrderID                     string    `gorm:"column:order_id;index" json:"order_id"`
	ProductID                   *string   `gorm:"column:product_id" json:"product_id,omitempty"`
	ProductName                 *string   `gorm:"column:product_name" json:"product_name,omitempty"`
	SkuID                       *string   `gorm:"column:sku_id" json:"sku_id,omitempty"`
	SellerSku                   *string   `gorm:"column:seller_sku;index" json:"seller_sku,omitempty"`
	Quantity                    int       `gorm:"column:quantity;default:0" json:"quantity"`
	SalePrice                   float64   `gorm:"column:sale_price;default:0" json:"sale_price"`
	OriginalPrice               float64   `gorm:"column:original_price;default:0" json:"original_price"`
	SubtotalAfterSellerDiscount float64   `gorm:"column:subtotal_after_seller_discount;default:0" json:"subtotal_after_seller_discount"`
	PlatformDiscount            float64   `gorm:"column:platform_discount;default:0" json:"platform_discount"`
	SellerDiscount              float64   `gorm:"column:seller_discount;default:0" json:"seller_discount"`
	Commission                  float64   `gorm:"column:commission;default:0" json:"commission"`
	TransactionFeeItem          float64   `gorm:"column:transaction_fee_item;default:0" json:"transaction_fee_item"`
	SettlementAmount            float64   `gorm:"column:settlement_amount;default:0" json:"settlement_amount"`
	RawItemData                 *string   `gorm:"column:raw_item_data" json:"raw_item_data,omitempty"`
	SyncedAt                    time.Time `gorm:"column:synced_at" json:"synced_at"`
	CreatedAt                   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt                   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (TiktokEscrowItem) TableName() string { return GetTableName("TiktokEscrowItem") }
