// Package tiktok provides core API types and methods for TikTok Shop
package tiktok

import "fmt"

// =============================================================================
// Core Response Types
// =============================================================================

// BaseResponse is common response structure
type BaseResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// =============================================================================
// Order List API (Legacy)
// =============================================================================

// OrderListResponse represents TikTok order list response
type OrderListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		OrderList []struct {
			OrderID          string  `json:"order_id"`
			OrderStatus      string  `json:"order_status"`
			TotalAmount      float64 `json:"payment_info.total_amount"`
			CreateTime       int64   `json:"create_time"`
			UpdateTime       int64   `json:"update_time"`
			RtsSlaTime       int64   `json:"rts_sla_time"`      // Ready-to-ship deadline (Unix timestamp)
			ShippingDueTime  int64   `json:"shipping_due_time"` // Ship by deadline
			ShippingProvider string  `json:"shipping_provider"` // Shipping carrier name
			TrackingNumber   string  `json:"tracking_number"`   // Tracking number
			BuyerMessage     string  `json:"buyer_message"`     // Buyer message/note
			BuyerEmail       string  `json:"buyer_email"`       // Buyer email/username
		} `json:"order_list"`
		TotalCount int `json:"total_count"`
	} `json:"data"`
}

// GetOrders fetches orders from TikTok API
func (c *Client) GetOrders(status string, pageSize int) (*OrderListResponse, error) {
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", pageSize),
	}
	if status != "" {
		params["order_status"] = status
	}

	var result OrderListResponse
	err := c.doRequest("GET", "/order/202309/orders/search", params, &result)
	return &result, err
}

// =============================================================================
// Product List API (Legacy)
// =============================================================================

// ProductListResponse represents TikTok product list response
type ProductListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Products []struct {
			ProductID   string  `json:"id"`
			Name        string  `json:"title"`
			Description string  `json:"description"`
			Price       float64 `json:"sale_price"`
			Status      string  `json:"status"`
		} `json:"products"`
		TotalCount int `json:"total_count"`
	} `json:"data"`
}

// GetProducts fetches products from TikTok API
func (c *Client) GetProducts(pageSize int) (*ProductListResponse, error) {
	params := map[string]string{
		"page_size": fmt.Sprintf("%d", pageSize),
	}

	var result ProductListResponse
	err := c.doRequest("GET", "/product/202309/products/search", params, &result)
	return &result, err
}

// =============================================================================
// Order Operations (Shipping/Cancel)
// =============================================================================

// ShipPackageRequest represents ship package request
type ShipPackageRequest struct {
	OrderID          string `json:"order_id"`
	PackageID        string `json:"package_id"`
	ShippingProvider string `json:"shipping_provider_id"`
	TrackingNumber   string `json:"tracking_number"`
}

// ShipPackageResponse represents ship package response
type ShipPackageResponse struct {
	BaseResponse
	Data struct {
		PackageID string `json:"package_id"`
	} `json:"data"`
}

// ShipPackage marks a package as shipped
func (c *Client) ShipPackage(req ShipPackageRequest) (*ShipPackageResponse, error) {
	params := map[string]string{
		"order_id":             req.OrderID,
		"package_id":           req.PackageID,
		"shipping_provider_id": req.ShippingProvider,
		"tracking_number":      req.TrackingNumber,
	}

	var result ShipPackageResponse
	err := c.doRequest("POST", "/fulfillment/202309/packages/ship", params, &result)
	return &result, err
}

// CancelOrderRequest represents cancel order request
type CancelOrderRequest struct {
	OrderID      string `json:"order_id"`
	CancelReason string `json:"cancel_reason"`
}

// CancelOrder cancels an order
func (c *Client) CancelOrder(req CancelOrderRequest) (*BaseResponse, error) {
	params := map[string]string{
		"order_id":      req.OrderID,
		"cancel_reason": req.CancelReason,
	}

	var result BaseResponse
	err := c.doRequest("POST", "/order/202309/orders/cancel", params, &result)
	return &result, err
}

// =============================================================================
// Inventory Update API
// =============================================================================

// UpdateInventoryRequest represents TikTok inventory update request
// API: POST /product/202309/products/{product_id}/inventory/update
type UpdateInventoryRequest struct {
	ProductID string             `json:"-"` // Used in URL path, not body
	Skus      []InventorySkuInfo `json:"skus"`
}

// InventorySkuInfo represents SKU inventory info
type InventorySkuInfo struct {
	ID        string              `json:"id"`
	Inventory []InventoryQuantity `json:"inventory"`
}

// InventoryQuantity represents inventory quantity (warehouse optional per TikTok API)
type InventoryQuantity struct {
	WarehouseID string `json:"warehouse_id,omitempty"`
	Quantity    int    `json:"quantity"`
}

// UpdateInventoryResponse represents inventory update response
type UpdateInventoryResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Skus []struct {
			ID      string `json:"id"`
			Success bool   `json:"success"`
		} `json:"skus"`
	} `json:"data"`
}

// UpdateInventory updates product inventory
// TikTok v202309 API: POST /product/202309/products/{product_id}/inventory/update
func (c *Client) UpdateInventory(req UpdateInventoryRequest) (*UpdateInventoryResponse, error) {
	params := map[string]string{}

	body := struct {
		Skus []InventorySkuInfo `json:"skus"`
	}{
		Skus: req.Skus,
	}

	var result UpdateInventoryResponse
	endpoint := fmt.Sprintf("/product/202309/products/%s/inventory/update", req.ProductID)
	err := c.doRequestWithBody("POST", endpoint, params, body, &result)

	if err != nil {
		return nil, err
	}

	// TikTok returns code 0 for success
	if result.Code != 0 {
		return &result, fmt.Errorf("tiktok API error: code=%d, message=%s", result.Code, result.Message)
	}

	return &result, nil
}
