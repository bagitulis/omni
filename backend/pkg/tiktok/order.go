// Package tiktok provides Order API types and methods for TikTok Shop
package tiktok

import (
	"context"
	"fmt"
)

// =============================================================================
// Order Search API
// =============================================================================

// OrderSearchRequest represents order search request
type OrderSearchRequest struct {
	OrderStatus  string `json:"order_status,omitempty"`
	CreateTimeGe int64  `json:"create_time_ge,omitempty"` // Unix timestamp
	CreateTimeLt int64  `json:"create_time_lt,omitempty"` // Unix timestamp
}

// OrderSearchResponse represents order search response
type OrderSearchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Orders        []TiktokOrder `json:"orders"`
		NextPageToken string        `json:"next_page_token"`
		TotalCount    int           `json:"total_count"`
	} `json:"data"`
}

// TiktokOrder represents a TikTok order
type TiktokOrder struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	CreateTime   int64  `json:"create_time"`
	UpdateTime   int64  `json:"update_time"`
	BuyerEmail   string `json:"buyer_email,omitempty"`
	BuyerMessage string `json:"buyer_message,omitempty"`
	PaymentInfo  struct {
		Currency            string `json:"currency"`
		OriginalTotalAmount string `json:"original_total_product_price"`
		TotalAmount         string `json:"total_amount"`
		SubTotal            string `json:"sub_total"`
		ShippingFee         string `json:"shipping_fee"`
		PlatformDiscount    string `json:"platform_discount"`
		SellerDiscount      string `json:"seller_discount"`
		ShippingFeeDiscount string `json:"shipping_fee_seller_discount"`
		ShippingFeePlatform string `json:"shipping_fee_platform_discount"`
	} `json:"payment_info"`
	LineItems        []TiktokOrderItem `json:"line_items"`
	RecipientAddress RecipientAddress  `json:"recipient_address,omitempty"`
}

// RecipientAddress represents the delivery address
type RecipientAddress struct {
	Name        string `json:"name"`
	Phone       string `json:"phone_number"`
	AddressLine string `json:"address_line1"`
	City        string `json:"city"`
	State       string `json:"state"`
	PostalCode  string `json:"postal_code"`
	Country     string `json:"region_code"`
}

// TiktokOrderItem represents an order item
type TiktokOrderItem struct {
	ID               string `json:"id"`
	SkuID            string `json:"sku_id"`
	SkuName          string `json:"sku_name"`
	ProductID        string `json:"product_id"`
	ProductName      string `json:"product_name"`
	SellerSku        string `json:"seller_sku"`
	Quantity         int    `json:"quantity"`
	OriginalPrice    string `json:"original_price"`
	SalePrice        string `json:"sale_price"`
	PlatformDiscount string `json:"platform_discount"`
	SellerDiscount   string `json:"seller_discount"`
}

// SearchOrders searches orders from TikTok API
func (c *Client) SearchOrders(ctx context.Context, req OrderSearchRequest, pageSize int, pageToken string) (*OrderSearchResponse, error) {
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", pageSize),
	}
	if pageToken != "" {
		params["page_token"] = pageToken
	}

	// Build request body
	body := make(map[string]interface{})
	if req.OrderStatus != "" {
		body["order_status"] = req.OrderStatus
	}
	if req.CreateTimeGe > 0 {
		body["create_time_ge"] = req.CreateTimeGe
	}
	if req.CreateTimeLt > 0 {
		body["create_time_lt"] = req.CreateTimeLt
	}

	var result OrderSearchResponse
	err := c.doRequestWithBody(ctx, "POST", "/order/202309/orders/search", params, body, &result)
	return &result, err
}
