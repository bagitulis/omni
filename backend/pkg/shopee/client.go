package shopee

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	BaseURLProduction = "https://partner.shopeemobile.com"
	BaseURLTest       = "https://partner.test-stable.shopeemobile.com"
)

// Client is the Shopee API client
type Client struct {
	partnerID   int64
	partnerKey  string
	shopID      int64
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

// GetEscrowDetailsRequest represents request for escrow details
type GetEscrowDetailsRequest struct {
	OrderSNList []string `json:"order_sn_list"`
}

// GetEscrowDetailsResponse represents response for escrow details batch
type GetEscrowDetailsResponse struct {
	Response []EscrowDetailWrapper `json:"response"`
	Error    string                `json:"error"`
	Message  string                `json:"message"`
}

// EscrowDetailWrapper wraps escrow_detail in batch response
type EscrowDetailWrapper struct {
	EscrowDetail *EscrowOrder `json:"escrow_detail"`
}

// GetOrderList returns flattened order list from batch response
func (r *GetEscrowDetailsResponse) GetOrderList() []EscrowOrder {
	var orders []EscrowOrder
	for _, wrapper := range r.Response {
		if wrapper.EscrowDetail != nil {
			orders = append(orders, *wrapper.EscrowDetail)
		}
	}
	return orders
}

// EscrowOrder represents an order in escrow response
type EscrowOrder struct {
	OrderSN          string          `json:"order_sn"`
	BuyerUsername    string          `json:"buyer_user_name"`
	PayTime          int64           `json:"pay_time"`
	OrderStatus      string          `json:"order_status"`
	TotalAmount      float64         `json:"escrow_amount"`
	OrderIncome      EscrowOrderData `json:"order_income"`
	BuyerPaymentInfo map[string]any  `json:"buyer_payment_info"`
}

// EscrowOrderData represents the order_income object from Shopee API
type EscrowOrderData struct {
	EscrowAmount               float64          `json:"escrow_amount"`
	BuyerTotalAmount           float64          `json:"buyer_total_amount"`
	OriginalPrice              float64          `json:"original_price"`
	SellerDiscount             float64          `json:"seller_discount"`
	ShopeeDiscount             float64          `json:"shopee_discount"`
	VoucherFromSeller          float64          `json:"voucher_from_seller"`
	VoucherFromShopee          float64          `json:"voucher_from_shopee"`
	Coins                      float64          `json:"coins"`
	BuyerPaidShippingFee       float64          `json:"buyer_paid_shipping_fee"`
	BuyerTransactionFee        float64          `json:"buyer_transaction_fee"`
	CrossBorderTax             float64          `json:"cross_border_tax"`
	PaymentPromotion           float64          `json:"payment_promotion"`
	CommissionFee              float64          `json:"commission_fee"`
	ServiceFee                 float64          `json:"service_fee"`
	SellerTransactionFee       float64          `json:"seller_transaction_fee"`
	SellerLostCompensation     float64          `json:"seller_lost_compensation"`
	SellerCoinCashBack         float64          `json:"seller_coin_cash_back"`
	EscrowTax                  float64          `json:"escrow_tax"`
	FinalShippingFee           float64          `json:"final_shipping_fee"`
	ActualShippingFee          float64          `json:"actual_shipping_fee"`
	ShopeeShippingRebate       float64          `json:"shopee_shipping_rebate"`
	ShippingFeeDiscountFrom3pl float64          `json:"shipping_fee_discount_from_3pl"`
	SellerShippingDiscount     float64          `json:"seller_shipping_discount"`
	EstimatedShippingFee       float64          `json:"estimated_shipping_fee"`
	SellerOrderProcessingFee   float64          `json:"seller_order_processing_fee"`
	DrcAdjustableRefund        float64          `json:"drc_adjustable_refund"`
	BuyerPaymentMethod         string           `json:"buyer_payment_method"`
	Items                      []EscrowItemData `json:"items"`
}

// EscrowItemData represents an item in escrow order_income
type EscrowItemData struct {
	ItemID                    int64   `json:"item_id"`
	ModelID                   int64   `json:"model_id"`
	ItemName                  string  `json:"item_name"`
	ModelName                 string  `json:"model_name"`
	ItemSKU                   string  `json:"item_sku"`
	ModelSKU                  string  `json:"model_sku"`
	QuantityPurchased         int     `json:"quantity_purchased"`
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

// EscrowOrderIncome represents income component (legacy, kept for compatibility)
type EscrowOrderIncome struct {
	EscrowAccountType string  `json:"escrow_account_type"`
	Amount            float64 `json:"amount"`
	Description       string  `json:"description"`
}

// EscrowOrderItem represents item in escrow (legacy, kept for compatibility)
type EscrowOrderItem struct {
	ItemID          int64   `json:"item_id"`
	ItemName        string  `json:"item_name"`
	ModelSKU        string  `json:"model_sku"`
	Quantity        int     `json:"quantity"`
	OriginalPrice   float64 `json:"original_price"`
	DiscountedPrice float64 `json:"discounted_price"`
}

// GetOrderListRequest represents request for order list
type GetOrderListRequest struct {
	TimeFrom       int64  `json:"time_from"`
	TimeTo         int64  `json:"time_to"`
	TimeRangeField string `json:"time_range_field"`
	PageSize       int    `json:"page_size"`
	Cursor         string `json:"cursor,omitempty"`
}

// GetOrderListResponse represents response for order list
type GetOrderListResponse struct {
	Response struct {
		OrderList  []OrderBasic `json:"order_list"`
		NextCursor string       `json:"next_cursor"`
		More       bool         `json:"more"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// OrderBasic represents basic order info
type OrderBasic struct {
	OrderSN string `json:"order_sn"`
}

// NewClient creates a new Shopee API client
func NewClient(partnerID int64, partnerKey string, isProduction bool) *Client {
	baseURL := BaseURLTest
	if isProduction {
		baseURL = BaseURLProduction
	}

	return &Client{
		partnerID:  partnerID,
		partnerKey: partnerKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetShopCredentials sets shop-level credentials
func (c *Client) SetShopCredentials(shopID int64, accessToken string) {
	c.shopID = shopID
	c.accessToken = accessToken
}

// generateSign creates HMAC-SHA256 signature for API calls
func (c *Client) generateSign(path string, timestamp int64) string {
	baseString := fmt.Sprintf("%d%s%d%s%d",
		c.partnerID, path, timestamp, c.accessToken, c.shopID)

	h := hmac.New(sha256.New, []byte(c.partnerKey))
	h.Write([]byte(baseString))
	return hex.EncodeToString(h.Sum(nil))
}

// buildURL constructs the full API URL with required params
func (c *Client) buildURL(path string, params map[string]string) string {
	timestamp := time.Now().Unix()
	sign := c.generateSign(path, timestamp)

	u, _ := url.Parse(c.baseURL + path)
	q := u.Query()
	q.Set("partner_id", strconv.FormatInt(c.partnerID, 10))
	q.Set("timestamp", strconv.FormatInt(timestamp, 10))
	q.Set("sign", sign)
	q.Set("shop_id", strconv.FormatInt(c.shopID, 10))
	q.Set("access_token", c.accessToken)

	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// doRequest executes HTTP request and parses response
func (c *Client) doRequest(method, path string, params map[string]string, result interface{}) error {
	reqURL := c.buildURL(path, params)

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Debug: log raw response for image-related APIs
	if strings.Contains(path, "get_item_extra_info") || strings.Contains(path, "get_item_base_info") {
		log.Printf("[Shopee API Debug] %s response (truncated): %s", path, truncateString(string(body), 500))
	}

	return json.Unmarshal(body, result)
}

// truncateString truncates string to maxLen
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// GetEscrowDetails gets escrow details for orders using batch endpoint
// Uses POST request with order_sn_list in JSON body
func (c *Client) GetEscrowDetails(req GetEscrowDetailsRequest) (*GetEscrowDetailsResponse, error) {
	path := "/api/v2/payment/get_escrow_detail_batch"

	// Build request body
	bodyData, _ := json.Marshal(map[string][]string{
		"order_sn_list": req.OrderSNList,
	})

	var result GetEscrowDetailsResponse
	err := c.doPostRequestWithBody(path, bodyData, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// doPostRequestWithBody executes HTTP POST request with body (no query params)
func (c *Client) doPostRequestWithBody(path string, body []byte, result interface{}) error {
	reqURL := c.buildURL(path, nil)

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(respBody, result)
}

// GetOrderList gets list of orders
func (c *Client) GetOrderList(timeFrom, timeTo int64, timeRangeField string, orderStatus string) (*GetOrderListResponse, error) {
	path := "/api/v2/order/get_order_list"
	params := map[string]string{
		"time_from":        strconv.FormatInt(timeFrom, 10),
		"time_to":          strconv.FormatInt(timeTo, 10),
		"time_range_field": timeRangeField,
		"page_size":        "100",
	}

	// Add order_status filter if specified
	if orderStatus != "" {
		params["order_status"] = orderStatus
	}

	var result GetOrderListResponse
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetShipmentInfo gets shipment info for an order (stub - implementing interface)
func (c *Client) GetShipmentInfo(ctx context.Context, orderSN string) (map[string]interface{}, error) {
	path := "/api/v2/logistics/get_shipment_info"
	params := map[string]string{
		"order_sn": orderSN,
	}

	var result map[string]interface{}
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetOrderDetail gets order details by order SNs
// Includes item_list via response_optional_fields
func (c *Client) GetOrderDetail(orderSNList []string) (*GetOrderDetailResponse, error) {
	path := "/api/v2/order/get_order_detail"
	params := map[string]string{
		"order_sn_list":            strings.Join(orderSNList, ","),
		"response_optional_fields": "item_list",
	}

	var result GetOrderDetailResponse
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetOrderDetailResponse represents response for order details
type GetOrderDetailResponse struct {
	Response struct {
		OrderList []OrderDetailItem `json:"order_list"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// OrderDetailItem represents an order detail item
type OrderDetailItem struct {
	OrderSN       string            `json:"order_sn"`
	OrderStatus   string            `json:"order_status"`
	TotalAmount   float64           `json:"total_amount"`
	Currency      string            `json:"currency"`
	CreateTime    int64             `json:"create_time"`
	UpdateTime    int64             `json:"update_time"`
	PaymentMethod string            `json:"payment_method"`
	BuyerUsername string            `json:"buyer_username"`
	ItemList      []OrderItemDetail `json:"item_list"` // Order items
}

// OrderItemDetail represents an item in an order
type OrderItemDetail struct {
	ItemID                 int64   `json:"item_id"`
	ModelID                int64   `json:"model_id"`
	ItemName               string  `json:"item_name"`
	ModelName              string  `json:"model_name"`
	ItemSKU                string  `json:"item_sku"`
	ModelSKU               string  `json:"model_sku"`
	ModelQuantityPurchased int     `json:"model_quantity_purchased"`
	ModelOriginalPrice     float64 `json:"model_original_price"`
	ModelDiscountedPrice   float64 `json:"model_discounted_price"`
}

// GetWalletBalance gets wallet balance (implements APIClient interface)
func (c *Client) GetWalletBalance(ctx context.Context) (map[string]interface{}, error) {
	path := "/api/v2/payment/get_wallet_transaction_list"
	params := map[string]string{}

	var result map[string]interface{}
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetShippingOptions gets shipping options for an order (implements APIClient interface)
func (c *Client) GetShippingOptions(ctx context.Context, orderSn string) (map[string]interface{}, error) {
	path := "/api/v2/logistics/get_shipping_parameter"
	params := map[string]string{
		"order_sn": orderSn,
	}

	var result map[string]interface{}
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTrackingInfo gets tracking info for an order (implements APIClient interface)
func (c *Client) GetTrackingInfo(ctx context.Context, orderSn string) (map[string]interface{}, error) {
	path := "/api/v2/logistics/get_tracking_number"
	params := map[string]string{
		"order_sn": orderSn,
	}

	var result map[string]interface{}
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}
