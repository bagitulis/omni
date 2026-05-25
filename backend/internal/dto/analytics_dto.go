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
	TotalSKU          int `json:"total_sku"`
	TotalTransactions int `json:"total_transactions"`
	SKUOk             int `json:"sku_ok"`
	SKUWithPriceDiff  int `json:"sku_with_price_diff"`
	SKUNoInventory    int `json:"sku_no_inventory"`
}

// SkuGroupDTO represents a SKU group result for Shopee reconciliation
type SkuGroupDTO struct {
	SKU              string  `json:"sku"`
	ItemName         string  `json:"item_name"`
	TotalQuantity    int     `json:"total_quantity"`
	TotalAmount      float64 `json:"total_amount"`
	SystemAmount     float64 `json:"system_amount"`
	PriceDiff        float64 `json:"price_diff"`
	PriceDiffPercent float64 `json:"price_diff_percent"`
	OrderCount       int     `json:"order_count"`
}

// TiktokSkuGroupDTO represents a SKU group result for TikTok reconciliation
type TiktokSkuGroupDTO struct {
	SKU              string  `json:"sku"`
	ItemName         string  `json:"item_name"`
	TotalQuantity    int     `json:"total_quantity"`
	TotalAmount      float64 `json:"total_amount"`
	SystemAmount     float64 `json:"system_amount"`
	PriceDiff        float64 `json:"price_diff"`
	PriceDiffPercent float64 `json:"price_diff_percent"`
	OrderCount       int     `json:"order_count"`
}

// ReconciliationResultDTO represents the Shopee reconciliation result
type ReconciliationResultDTO struct {
	Summary ReconciliationSummaryDTO `json:"summary"`
	Details []SkuGroupDTO            `json:"details"`
}

// TiktokReconciliationResultDTO represents the TikTok reconciliation result
type TiktokReconciliationResultDTO struct {
	Summary ReconciliationSummaryDTO `json:"summary"`
	Details []TiktokSkuGroupDTO      `json:"details"`
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
	PlatformFee float64 `json:"platform_fee"`
	ActualFee   float64 `json:"actual_fee"`
	Difference  float64 `json:"difference"`
	Status      string  `json:"status"`
	OrderDate   string  `json:"order_date"`
}

// ShopeeShippingFeeResultDTO represents the Shopee shipping fee analysis result
type ShopeeShippingFeeResultDTO struct {
	Summary ShippingFeeSummaryDTO   `json:"summary"`
	Details []ShopeeShippingOrderDTO `json:"details"`
}

// TiktokShippingOrderDTO represents a TikTok shipping order entry
type TiktokShippingOrderDTO struct {
	OrderSN     string  `json:"order_sn"`
	ShippingFee float64 `json:"shipping_fee"`
	ActualFee   float64 `json:"actual_fee"`
	Difference  float64 `json:"difference"`
	Status      string  `json:"status"`
	OrderDate   string  `json:"order_date"`
}

// TiktokShippingFeeResultDTO represents the TikTok shipping fee analysis result
type TiktokShippingFeeResultDTO struct {
	Summary ShippingFeeSummaryDTO    `json:"summary"`
	Details []TiktokShippingOrderDTO `json:"details"`
}
