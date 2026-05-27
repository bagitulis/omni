package dto

import "time"

// AnalyticsSettingsDTO represents analytics settings for a platform
type AnalyticsSettingsDTO struct {
	PriceColumn        string  `json:"price_column"`
	FormulaDeduction   float64 `json:"formula_deduction"`
	FormulaMultiplier  float64 `json:"formula_multiplier"`
}

// SyncStatusDTO represents the sync status for a given month/year
type SyncStatusDTO struct {
	Synced       bool       `json:"synced"`
	TotalOrders  int        `json:"total_orders"`
	FailedOrders int        `json:"failed_orders"`
	SyncedAt     *time.Time `json:"synced_at"`
}

// SyncRequestDTO represents a sync request with month/year parameters
type SyncRequestDTO struct {
	Month       int  `json:"month" form:"month"`
	Year        int  `json:"year" form:"year"`
	ForceResync bool `json:"force_resync" form:"force_resync"`
}

// SyncResultDTO represents the result of a sync operation
type SyncResultDTO struct {
	JobID string `json:"job_id"`
}

// ReconciliationSummaryDTO summarizes reconciliation results
type ReconciliationSummaryDTO struct {
	TotalSku          int `json:"total_sku"`
	TotalTransactions int `json:"total_transactions"`
	SkuOk             int `json:"sku_ok"`
	SkuWithPriceDiff  int `json:"sku_with_price_diff"`
	SkuNoInventory    int `json:"sku_no_inventory"`
}

// SkuGroupDTO represents a SKU group result for Shopee reconciliation.
// Uses per-unit price analysis with inventory lookup status classification.
// Status values: OK, PRICE_DIFF, NO_INVENTORY.
type SkuGroupDTO struct {
	Sku                 string    `json:"sku"`
	ModelSku            string    `json:"model_sku"`
	ItemName            string    `json:"item_name"`
	ModelName           string    `json:"model_name"`
	VariantName         string    `json:"variant_name"`
	InventoryPrice      *float64  `json:"inventory_price"`
	ExpectedIncome      *float64  `json:"expected_income"`
	TotalTransactions   int       `json:"total_transactions"`
	UniqueUnitPrices    []float64 `json:"unique_unit_prices"`
	UniqueActualIncomes []float64 `json:"unique_actual_incomes"`
	HasMultiplePrices   bool      `json:"has_multiple_prices"`
	HasPriceDifference  bool      `json:"has_price_difference"`
	Status              string    `json:"status"`
}

// TiktokSkuGroupDTO represents a SKU group result for TikTok reconciliation.
// Uses per-unit price analysis with inventory lookup status classification.
type TiktokSkuGroupDTO struct {
	Sku                 string    `json:"sku"`
	SellerSku           string    `json:"seller_sku"`
	ProductName         string    `json:"product_name"`
	VariantName         string    `json:"variant_name"`
	InventoryPrice      *float64  `json:"inventory_price"`
	ExpectedIncome      *float64  `json:"expected_income"`
	TotalTransactions   int       `json:"total_transactions"`
	UniqueUnitPrices    []float64 `json:"unique_unit_prices"`
	UniqueActualIncomes []float64 `json:"unique_actual_incomes"`
	HasMultiplePrices   bool      `json:"has_multiple_prices"`
	HasPriceDifference  bool      `json:"has_price_difference"`
	Status              string    `json:"status"`
}

// ReconciliationResultDTO represents the Shopee reconciliation result
type ReconciliationResultDTO struct {
	Summary   ReconciliationSummaryDTO `json:"summary"`
	SkuGroups []SkuGroupDTO            `json:"sku_groups"`
}

// TiktokReconciliationResultDTO represents the TikTok reconciliation result
type TiktokReconciliationResultDTO struct {
	Summary   ReconciliationSummaryDTO `json:"summary"`
	SkuGroups []TiktokSkuGroupDTO      `json:"sku_groups"`
}

// ShippingFeeSummaryDTO summarizes shipping fee analysis
type ShippingFeeSummaryDTO struct {
	TotalOrders          int     `json:"total_orders"`
	OrdersWithDifference int     `json:"orders_with_difference"`
	TotalProfit          float64 `json:"total_profit"`
	TotalLoss            float64 `json:"total_loss"`
	NetImpact            float64 `json:"net_impact"`
}

// ShopeeShippingOrderDTO represents a Shopee shipping order entry
type ShopeeShippingOrderDTO struct {
	OrderSN     string  `json:"order_sn"`
	BuyerPaid   float64 `json:"buyer_paid"`
	ActualFee   float64 `json:"actual_fee"`
	ShopeeRebate float64 `json:"shopee_rebate"`
	Difference  float64 `json:"difference"`
	Status      string  `json:"status"`
	OrderDate   string  `json:"order_date"`
	BuyerName   string  `json:"buyer_name"`
	PaymentMethod string `json:"payment_method"`
}

// ShopeeShippingFeeResultDTO represents the Shopee shipping fee analysis result
type ShopeeShippingFeeResultDTO struct {
	Summary ShippingFeeSummaryDTO    `json:"summary"`
	Details []ShopeeShippingOrderDTO `json:"details"`
}

// TiktokShippingOrderDTO represents a TikTok shipping order entry
type TiktokShippingOrderDTO struct {
	OrderSN          string  `json:"order_sn"`
	CustomerPaid     float64 `json:"customer_paid"`
	ActualFee        float64 `json:"actual_fee"`
	PlatformDiscount float64 `json:"platform_discount"`
	Difference       float64 `json:"difference"`
	Status           string  `json:"status"`
	OrderDate        string  `json:"order_date"`
	OrderStatus      string  `json:"order_status"`
	Currency         string  `json:"currency"`
}

// TiktokShippingFeeResultDTO represents the TikTok shipping fee analysis result
type TiktokShippingFeeResultDTO struct {
	Summary ShippingFeeSummaryDTO    `json:"summary"`
	Details []TiktokShippingOrderDTO `json:"details"`
}

// ShopeeSkuOrderDTO represents a Shopee order associated with a SKU for drill-down.
type ShopeeSkuOrderDTO struct {
	ID                   string  `json:"id"`
	OrderSN              string  `json:"order_sn"`
	EscrowAmount         float64 `json:"escrow_amount"`
	CommissionFee        float64 `json:"commission_fee"`
	ServiceFee           float64 `json:"service_fee"`
	SellerProcessingFee  float64 `json:"seller_processing_fee"`
	BuyerPaidShippingFee float64 `json:"buyer_paid_shipping_fee"`
	ActualShippingFee    float64 `json:"actual_shipping_fee"`
	ShopeeShippingRebate float64 `json:"shopee_shipping_rebate"`
	EstimatedShippingFee float64 `json:"estimated_shipping_fee"`
	BuyerTotalAmount     float64 `json:"buyer_total_amount"`
	BuyerName            string  `json:"buyer_name"`
	PaymentMethod        string  `json:"payment_method"`
	OrderDate            string  `json:"order_date"`
	ItemName             string  `json:"item_name"`
	ModelName            string  `json:"model_name"`
	Sku                  string  `json:"sku"`
	ModelSku             string  `json:"model_sku"`
	Quantity             int     `json:"quantity"`
	OriginalPrice        float64 `json:"original_price"`
}

// ShopeeSkuOrdersResultDTO wraps Shopee SKU orders result.
type ShopeeSkuOrdersResultDTO struct {
	Orders []ShopeeSkuOrderDTO `json:"orders"`
}

// ShopeeOrderItemDTO represents a Shopee escrow item in an order.
type ShopeeOrderItemDTO struct {
	ID                        string  `json:"id"`
	EscrowOrderID             string  `json:"escrow_order_id"`
	ItemID                    *int64  `json:"item_id,omitempty"`
	ModelID                   *int64  `json:"model_id,omitempty"`
	Sku                       string  `json:"sku"`
	ModelSku                  string  `json:"model_sku"`
	ItemName                  string  `json:"item_name"`
	ModelName                 string  `json:"model_name"`
	Quantity                  int     `json:"quantity"`
	OriginalPrice             float64 `json:"original_price"`
	SellingPrice              float64 `json:"selling_price"`
	DiscountedPrice           float64 `json:"discounted_price"`
	SellerDiscount            float64 `json:"seller_discount"`
	ShopeeDiscount            float64 `json:"shopee_discount"`
	DiscountFromCoin          float64 `json:"discount_from_coin"`
	DiscountFromVoucherSeller float64 `json:"discount_from_voucher_seller"`
	DiscountFromVoucherShopee float64 `json:"discount_from_voucher_shopee"`
	AmsCommissionFee          float64 `json:"ams_commission_fee"`
	SellerOrderProcessingFee  float64 `json:"seller_order_processing_fee"`
}

// ShopeeOrderItemsResultDTO wraps Shopee order items result.
type ShopeeOrderItemsResultDTO struct {
	Items []ShopeeOrderItemDTO `json:"items"`
}

// TiktokSkuOrderDTO represents a TikTok order associated with a SKU for drill-down.
type TiktokSkuOrderDTO struct {
	ID                          string  `json:"id"`
	OrderID                     string  `json:"order_id"`
	OrderStatus                 string  `json:"order_status"`
	TotalSettlementAmount       float64 `json:"total_settlement_amount"`
	ProductRevenue              float64 `json:"product_revenue"`
	PlatformCommission           float64 `json:"platform_commission"`
	TransactionFee              float64 `json:"transaction_fee"`
	ShippingFeeCustomerPaid     float64 `json:"shipping_fee_customer_paid"`
	ShippingFeeActual           float64 `json:"shipping_fee_actual"`
	ShippingFeePlatformDiscount float64 `json:"shipping_fee_platform_discount"`
	SellerShippingDiscount      float64 `json:"seller_shipping_discount"`
	RefundAmount                float64 `json:"refund_amount"`
	Currency                    string  `json:"currency"`
	BuyerName                   string  `json:"buyer_name"`
	OrderDate                   string  `json:"order_date"`
	ProductName                 string  `json:"product_name"`
	SellerSku                   string  `json:"seller_sku"`
	Quantity                    int     `json:"quantity"`
	SalePrice                   float64 `json:"sale_price"`
	OriginalPrice               float64 `json:"original_price"`
}

// TiktokSkuOrdersResultDTO wraps TikTok SKU orders result.
type TiktokSkuOrdersResultDTO struct {
	Orders []TiktokSkuOrderDTO `json:"orders"`
}

// TiktokOrderItemDTO represents a TikTok escrow item in an order.
type TiktokOrderItemDTO struct {
	ID                          string  `json:"id"`
	EscrowOrderID               string  `json:"escrow_order_id"`
	ProductName                 string  `json:"product_name"`
	SkuID                       string  `json:"sku_id"`
	SellerSku                   string  `json:"seller_sku"`
	Quantity                    int     `json:"quantity"`
	SalePrice                   float64 `json:"sale_price"`
	OriginalPrice               float64 `json:"original_price"`
	SubtotalAfterSellerDiscount float64 `json:"subtotal_after_seller_discount"`
	PlatformDiscount            float64 `json:"platform_discount"`
	SellerDiscount              float64 `json:"seller_discount"`
	Commission                  float64 `json:"commission"`
	TransactionFeeItem           float64 `json:"transaction_fee_item"`
	SettlementAmount            float64 `json:"settlement_amount"`
}

// TiktokOrderItemsResultDTO wraps TikTok order items result.
type TiktokOrderItemsResultDTO struct {
	Items []TiktokOrderItemDTO `json:"items"`
}
