// Package tiktok provides Finance API types and methods for TikTok Shop
package tiktok

// =============================================================================
// Finance API - Get Transactions by Order
// =============================================================================

// OrderTransactionResponse represents order transaction response
type OrderTransactionResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		OrderID               string                 `json:"order_id"`
		Currency              string                 `json:"currency"`
		SettlementAmount      string                 `json:"settlement_amount"` // Total settlement amount for the order
		SkuTransactions       []SkuTransaction       `json:"sku_transactions"`
		OrderLevelCharges     []OrderLevelCharge     `json:"order_level_charges"`
		StatementTransactions []StatementTransaction `json:"statement_transactions,omitempty"` // v202309
	} `json:"data"`
}

// SkuTransaction represents SKU level transaction from TikTok Finance API v202501
// Fields based on: /finance/202501/orders/{order_id}/statement_transactions
type SkuTransaction struct {
	SkuID                 string `json:"sku_id"`
	ProductName           string `json:"product_name"`
	SkuName               string `json:"sku_name"`
	Quantity              int    `json:"quantity"`
	RevenueAmount         string `json:"revenue_amount"`    // Total revenue for this SKU (sale_price equivalent)
	SettlementAmount      string `json:"settlement_amount"` // SKU-level settlement amount
	SkuSubtotalBeforeDisc string `json:"sku_subtotal_before_discount"`
	SkuPlatformDiscount   string `json:"sku_platform_discount"`
	SkuSellerDiscount     string `json:"sku_seller_discount"`
	SkuExtPlatformDisc    string `json:"sku_ext_platform_discount"`
	SkuExtSellerDisc      string `json:"sku_ext_seller_discount"`
	SkuSubtotalAfterDisc  string `json:"sku_subtotal_after_discount"`
	RetailDeliveryFee     string `json:"sku_retail_delivery_fee"`
	SkuEstimatedPkg       string `json:"sku_estimated_package_on_buyer"`
	TransactionFee        string `json:"transaction_fee"`
	ReferralFee           string `json:"referral_fee"`
	AffiliateCommission   string `json:"affiliate_commission"`
	AffiliatePartnerComm  string `json:"affiliate_partner_commission"`
	SkuNetSales           string `json:"sku_net_sales"`
	SkuNetPayout          string `json:"sku_net_payout"`
}

// OrderLevelCharge represents order level charges
type OrderLevelCharge struct {
	ChargeType   string `json:"charge_type"`
	ChargeAmount string `json:"charge_amount"`
}

// StatementTransaction represents statement transaction (v202309)
type StatementTransaction struct {
	StatementID     string `json:"statement_id"`
	StatementTime   int64  `json:"statement_time"`
	TransactionType string `json:"transaction_type"`
	Amount          string `json:"amount"`
	Currency        string `json:"currency"`
}

// GetOrderTransactions fetches transaction details for an order (v202501 API)
func (c *Client) GetOrderTransactions(orderID string) (*OrderTransactionResponse, error) {
	params := map[string]string{}

	var result OrderTransactionResponse
	// Use v202501 API for full SKU transaction details
	err := c.doRequest("GET", "/finance/202501/orders/"+orderID+"/statement_transactions", params, &result)
	return &result, err
}

// GetOrderTransactionsV202309 fetches transactions using older API version
func (c *Client) GetOrderTransactionsV202309(orderID string) (*OrderTransactionResponse, error) {
	params := map[string]string{}

	var result OrderTransactionResponse
	err := c.doRequest("GET", "/finance/202309/orders/"+orderID+"/statement_transactions", params, &result)
	return &result, err
}
