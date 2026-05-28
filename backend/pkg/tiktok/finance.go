// Package tiktok provides Finance API types and methods for TikTok Shop
package tiktok

import (
	"context"
	"fmt"
)

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
		SettlementAmount      string                 `json:"settlement_amount"`   // Total settlement amount for the order
		ShippingCostAmount    string                 `json:"shipping_cost_amount"` // Total shipping cost (Actual Fee)
		SkuTransactions       []SkuTransaction       `json:"sku_transactions"`
		OrderLevelCharges     []OrderLevelCharge     `json:"order_level_charges"`
		StatementTransactions []StatementTransaction `json:"statement_transactions,omitempty"` // v202309
	} `json:"data"`
}

// SkuTransaction represents SKU level transaction from TikTok Finance API v202501
// Fields based on: /finance/202501/orders/{order_id}/statement_transactions
// Note: Quantity is string per TikTok official SDK (not int)
type SkuTransaction struct {
	SkuID                 string `json:"sku_id"`
	ProductName           string `json:"product_name"`
	SkuName               string `json:"sku_name"`
	Quantity              string `json:"quantity"`
	RevenueAmount         string      `json:"revenue_amount"`    // Total revenue for this SKU (sale_price equivalent)
	SettlementAmount      string      `json:"settlement_amount"` // SKU-level settlement amount
	SkuSubtotalBeforeDisc string      `json:"sku_subtotal_before_discount"`
	SkuPlatformDiscount   string      `json:"sku_platform_discount"`
	SkuSellerDiscount     string      `json:"sku_seller_discount"`
	SkuExtPlatformDisc    string      `json:"sku_ext_platform_discount"`
	SkuExtSellerDisc      string      `json:"sku_ext_seller_discount"`
	SkuSubtotalAfterDisc  string      `json:"sku_subtotal_after_discount"`
	RetailDeliveryFee     string      `json:"sku_retail_delivery_fee"`
	SkuEstimatedPkg       string      `json:"sku_estimated_package_on_buyer"`
	TransactionFee        string      `json:"transaction_fee"`
	ReferralFee           string      `json:"referral_fee"`
	AffiliateCommission   string      `json:"affiliate_commission"`
	AffiliatePartnerComm  string      `json:"affiliate_partner_commission"`
	SkuNetSales           string      `json:"sku_net_sales"`
	SkuNetPayout          string      `json:"sku_net_payout"`
}

// OrderLevelCharge represents order level charges
type OrderLevelCharge struct {
	ChargeType   string `json:"charge_type"`
	ChargeAmount string `json:"charge_amount"`
}

// StatementTransaction represents statement transaction (v202309)
type StatementTransaction struct {
	StatementID                     string `json:"statement_id"`
	StatementTime                   int64  `json:"statement_time"`
	TransactionType                 string `json:"transaction_type"`
	Amount                          string `json:"amount"`
	Currency                        string `json:"currency"`
	SettlementAmount                string `json:"settlement_amount"`                     // Total settlement for this transaction
	ActualShippingFeeAmount         string `json:"actual_shipping_fee_amount"`           // Actual shipping cost
	CustomerPaidShippingFeeAmount   string `json:"customer_paid_shipping_fee_amount"`     // Buyer Paid
	PlatformShippingFeeDiscountAmount string `json:"platform_shipping_fee_discount_amount"` // Platform Subsidy
}

// GetOrderTransactions fetches transaction details for an order (v202501 API)
func (c *Client) GetOrderTransactions(ctx context.Context, orderID string) (*OrderTransactionResponse, error) {
	params := map[string]string{}

	var result OrderTransactionResponse
	// Use v202501 API for full SKU transaction details
	err := c.doRequest(ctx, "GET", "/finance/202501/orders/"+orderID+"/statement_transactions", params, &result)
	return &result, err
}

// GetOrderTransactionsV202309 fetches transactions using older API version
func (c *Client) GetOrderTransactionsV202309(ctx context.Context, orderID string) (*OrderTransactionResponse, error) {
	params := map[string]string{}

	var result OrderTransactionResponse
	err := c.doRequest(ctx, "GET", "/finance/202309/orders/"+orderID+"/statement_transactions", params, &result)
	return &result, err
}

// =============================================================================
// Finance API - Statement-Based (Statement-First Sync)
// =============================================================================

// StatementListResponse is the response from GET /finance/202309/statements
type StatementListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Statements    []StatementItem `json:"statements"`
		NextPageToken string          `json:"next_page_token"`
	} `json:"data"`
}

// StatementItem represents one statement entry (a daily settlement summary)
type StatementItem struct {
	StatementID string `json:"statement_id"`
}

// StatementTxListResponse is the response from GET /finance/202501/statements/{id}/statement_transactions
// Uses v202501 which is applicable for all regions including SEA (ID, TH, MY, VN, PH)
type StatementTxListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Transactions  []StatementTxItem `json:"transactions"`
		NextPageToken string            `json:"next_page_token"`
	} `json:"data"`
}

// StatementTxItem represents one order-level transaction within a statement
type StatementTxItem struct {
	OrderID          string `json:"order_id"`
	Type             string `json:"type"`             // "ORDER", "ADJUSTMENT", "RESERVE", etc.
	SettlementAmount string `json:"settlement_amount"` // Net seller payout
	RevenueAmount    string `json:"revenue_amount"`    // Gross sales amount
}

// GetStatements fetches statement IDs for a settlement time window.
// Uses /finance/202309/statements which works for all regions.
// TikTok generates one statement per day at 00:00 UTC.
func (c *Client) GetStatements(ctx context.Context, startTime, endTime int64, pageToken string) (*StatementListResponse, error) {
	params := map[string]string{
		"sort_field":        "statement_time",
		"sort_order":        "ASC",
		"statement_time_ge": fmt.Sprintf("%d", startTime),
		"statement_time_lt": fmt.Sprintf("%d", endTime),
		"page_size":         "100",
	}
	if pageToken != "" {
		params["page_token"] = pageToken
	}
	var result StatementListResponse
	err := c.doRequest(ctx, "GET", "/finance/202309/statements", params, &result)
	return &result, err
}

// GetStatementTransactions fetches order transactions for a specific statement.
// Uses /finance/202501 (all-region support including SEA/Indonesia).
func (c *Client) GetStatementTransactions(ctx context.Context, statementID, pageToken string) (*StatementTxListResponse, error) {
	params := map[string]string{
		"sort_field": "order_create_time",
		"sort_order": "ASC",
		"page_size":  "100",
	}
	if pageToken != "" {
		params["page_token"] = pageToken
	}
	var result StatementTxListResponse
	err := c.doRequest(ctx, "GET", "/finance/202501/statements/"+statementID+"/statement_transactions", params, &result)
	return &result, err
}

