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
