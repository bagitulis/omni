// Package dto provides Data Transfer Objects for API requests/responses
package dto

// AnalyticsSettingsDTO for settings API response
type AnalyticsSettingsDTO struct {
	PriceColumn       string  `json:"price_column"`
	FormulaDeduction  float64 `json:"formula_deduction"`
	FormulaMultiplier float64 `json:"formula_multiplier"`
}

// SyncStatusDTO represents sync status response
type SyncStatusDTO struct {
	Synced       bool    `json:"synced"`
	TotalOrders  int     `json:"total_orders"`
	FailedOrders int     `json:"failed_orders"`
	SyncedAt     *string `json:"synced_at"`
}

// SyncRequestDTO for sync API request
type SyncRequestDTO struct {
	Month       int  `json:"month" binding:"required,min=1,max=12"`
	Year        int  `json:"year" binding:"required,min=2020"`
	ForceResync bool `json:"force_resync"`
}

// SyncResultDTO represents sync operation result
type SyncResultDTO struct {
	TotalOrders  int    `json:"total_orders"`
	TotalItems   int    `json:"total_items"`
	FailedOrders int    `json:"failed_orders,omitempty"`
	Message      string `json:"message,omitempty"`
}

// ReconciliationSummaryDTO represents reconciliation summary
type ReconciliationSummaryDTO struct {
	TotalSku          int `json:"total_sku"`
	TotalTransactions int `json:"total_transactions"`
	SkuOk             int `json:"sku_ok"`
	SkuWithPriceDiff  int `json:"sku_with_price_diff"`
	SkuNoInventory    int `json:"sku_no_inventory"`
}

// SkuGroupDTO represents a single SKU group in reconciliation
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
	Status              string    `json:"status"` // OK, PRICE_DIFF, NO_INVENTORY
}

// TiktokSkuGroupDTO for TikTok-specific SKU group
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

// ReconciliationResultDTO represents full reconciliation result
type ReconciliationResultDTO struct {
	Summary   ReconciliationSummaryDTO `json:"summary"`
	SkuGroups []SkuGroupDTO            `json:"sku_groups"`
}

// TiktokReconciliationResultDTO for TikTok reconciliation
type TiktokReconciliationResultDTO struct {
	Summary   ReconciliationSummaryDTO `json:"summary"`
	SkuGroups []TiktokSkuGroupDTO      `json:"sku_groups"`
}

// ShippingFeeSummaryDTO represents shipping fee summary
type ShippingFeeSummaryDTO struct {
	TotalOrders          int     `json:"total_orders"`
	OrdersWithDifference int     `json:"orders_with_difference"`
	TotalProfit          float64 `json:"total_profit"`
	TotalLoss            float64 `json:"total_loss"`
	NetImpact            float64 `json:"net_impact"`
}

// ShopeeShippingOrderDTO for Shopee shipping fee analysis
type ShopeeShippingOrderDTO struct {
	OrderSn       string  `json:"order_sn"`
	OrderDate     *string `json:"order_date"`
	BuyerPaid     float64 `json:"buyer_paid"`
	ActualFee     float64 `json:"actual_fee"`
	ShopeeRebate  float64 `json:"shopee_rebate"`
	Difference    float64 `json:"difference"`
	BuyerName     *string `json:"buyer_name"`
	PaymentMethod *string `json:"payment_method"`
}

// TiktokShippingOrderDTO for TikTok shipping fee analysis
type TiktokShippingOrderDTO struct {
	OrderID          string  `json:"order_id"`
	OrderDate        *string `json:"order_date"`
	CustomerPaid     float64 `json:"customer_paid"`
	ActualFee        float64 `json:"actual_fee"`
	PlatformDiscount float64 `json:"platform_discount"`
	Difference       float64 `json:"difference"`
	OrderStatus      string  `json:"order_status"`
	Currency         string  `json:"currency"`
}

// ShopeeShippingFeeResultDTO for Shopee shipping analysis
type ShopeeShippingFeeResultDTO struct {
	Summary ShippingFeeSummaryDTO    `json:"summary"`
	Orders  []ShopeeShippingOrderDTO `json:"orders"`
}

// TiktokShippingFeeResultDTO for TikTok shipping analysis
type TiktokShippingFeeResultDTO struct {
	Summary ShippingFeeSummaryDTO    `json:"summary"`
	Orders  []TiktokShippingOrderDTO `json:"orders"`
}
