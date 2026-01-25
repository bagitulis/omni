// Package dto provides Data Transfer Objects for API requests/responses
package dto

// AnalyticsSettingsDTO for settings API response
type AnalyticsSettingsDTO struct {
	PriceColumn       string  `json:"priceColumn"`
	FormulaDeduction  float64 `json:"formulaDeduction"`
	FormulaMultiplier float64 `json:"formulaMultiplier"`
}

// SyncStatusDTO represents sync status response
type SyncStatusDTO struct {
	Synced      bool    `json:"synced"`
	TotalOrders int     `json:"totalOrders"`
	SyncedAt    *string `json:"syncedAt"`
}

// SyncRequestDTO for sync API request
type SyncRequestDTO struct {
	Month       int  `json:"month" binding:"required,min=1,max=12"`
	Year        int  `json:"year" binding:"required,min=2020"`
	ForceResync bool `json:"forceResync"`
}

// SyncResultDTO represents sync operation result
type SyncResultDTO struct {
	TotalOrders  int    `json:"totalOrders"`
	TotalItems   int    `json:"totalItems"`
	FailedOrders int    `json:"failedOrders,omitempty"`
	Message      string `json:"message,omitempty"`
}

// ReconciliationSummaryDTO represents reconciliation summary
type ReconciliationSummaryDTO struct {
	TotalSku          int `json:"totalSku"`
	TotalTransactions int `json:"totalTransactions"`
	SkuOk             int `json:"skuOk"`
	SkuWithPriceDiff  int `json:"skuWithPriceDiff"`
	SkuNoInventory    int `json:"skuNoInventory"`
}

// SkuGroupDTO represents a single SKU group in reconciliation
type SkuGroupDTO struct {
	Sku                 string    `json:"sku"`
	ModelSku            string    `json:"modelSku"`
	ItemName            string    `json:"itemName"`
	ModelName           string    `json:"modelName"`
	InventoryPrice      *float64  `json:"inventoryPrice"`
	ExpectedIncome      *float64  `json:"expectedIncome"`
	TotalTransactions   int       `json:"totalTransactions"`
	UniqueUnitPrices    []float64 `json:"uniqueUnitPrices"`
	UniqueActualIncomes []float64 `json:"uniqueActualIncomes"`
	HasMultiplePrices   bool      `json:"hasMultiplePrices"`
	HasPriceDifference  bool      `json:"hasPriceDifference"`
	Status              string    `json:"status"` // OK, PRICE_DIFF, NO_INVENTORY
}

// TiktokSkuGroupDTO for TikTok-specific SKU group
type TiktokSkuGroupDTO struct {
	Sku                 string    `json:"sku"`
	SellerSku           string    `json:"sellerSku"`
	ProductName         string    `json:"productName"`
	VariantName         string    `json:"variantName"`
	InventoryPrice      *float64  `json:"inventoryPrice"`
	ExpectedIncome      *float64  `json:"expectedIncome"`
	TotalTransactions   int       `json:"totalTransactions"`
	UniqueUnitPrices    []float64 `json:"uniqueUnitPrices"`
	UniqueActualIncomes []float64 `json:"uniqueActualIncomes"`
	HasMultiplePrices   bool      `json:"hasMultiplePrices"`
	HasPriceDifference  bool      `json:"hasPriceDifference"`
	Status              string    `json:"status"`
}

// ReconciliationResultDTO represents full reconciliation result
type ReconciliationResultDTO struct {
	Summary   ReconciliationSummaryDTO `json:"summary"`
	SkuGroups []SkuGroupDTO            `json:"skuGroups"`
}

// TiktokReconciliationResultDTO for TikTok reconciliation
type TiktokReconciliationResultDTO struct {
	Summary   ReconciliationSummaryDTO `json:"summary"`
	SkuGroups []TiktokSkuGroupDTO      `json:"skuGroups"`
}

// ShippingFeeSummaryDTO represents shipping fee summary
type ShippingFeeSummaryDTO struct {
	TotalOrders          int     `json:"totalOrders"`
	OrdersWithDifference int     `json:"ordersWithDifference"`
	TotalProfit          float64 `json:"totalProfit"`
	TotalLoss            float64 `json:"totalLoss"`
	NetImpact            float64 `json:"netImpact"`
}

// ShopeeShippingOrderDTO for Shopee shipping fee analysis
type ShopeeShippingOrderDTO struct {
	OrderSn       string  `json:"orderSn"`
	OrderDate     *string `json:"orderDate"`
	BuyerPaid     float64 `json:"buyerPaid"`
	ActualFee     float64 `json:"actualFee"`
	ShopeeRebate  float64 `json:"shopeeRebate"`
	Difference    float64 `json:"difference"`
	BuyerName     *string `json:"buyerName"`
	PaymentMethod *string `json:"paymentMethod"`
}

// TiktokShippingOrderDTO for TikTok shipping fee analysis
type TiktokShippingOrderDTO struct {
	OrderID         string  `json:"order_id"`
	OrderDate       *string `json:"order_date"`
	CustomerPaid    float64 `json:"customer_paid"`
	ActualCost      float64 `json:"actual_cost"`
	PlatformSubsidy float64 `json:"platform_subsidy"`
	Difference      float64 `json:"difference"`
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
