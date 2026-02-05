// Package tiktok provides Shipping/Fulfillment API types and methods for TikTok Shop
package tiktok

import "fmt"

// =============================================================================
// Order Detail API (includes packages)
// =============================================================================

// OrderDetailResponse represents order detail response with packages
type OrderDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Orders []OrderDetailData `json:"orders"`
	} `json:"data"`
}

// OrderDetailData represents detailed order data
type OrderDetailData struct {
	ID               string            `json:"id"`
	Status           string            `json:"status"`
	CreateTime       int64             `json:"create_time"`
	UpdateTime       int64             `json:"update_time"`
	RtsSlaTime       int64             `json:"rts_sla_time"`       // Ready-to-ship deadline
	ShippingDueTime  int64             `json:"shipping_due_time"`  // Ship by deadline
	Packages         []PackageInfo     `json:"packages,omitempty"` // Package information
	LineItems        []OrderLineItem   `json:"line_items,omitempty"`
	RecipientAddress *RecipientAddress `json:"recipient_address,omitempty"`
	PaymentInfo      *PaymentInfo      `json:"payment_info,omitempty"`
}

// OrderLineItem represents a line item in the order
type OrderLineItem struct {
	ID            string `json:"id"`
	SkuID         string `json:"sku_id"`
	SkuName       string `json:"sku_name"`
	ProductID     string `json:"product_id"`
	ProductName   string `json:"product_name"`
	SellerSku     string `json:"seller_sku"`
	Quantity      int    `json:"quantity"`
	SalePrice     string `json:"sale_price"`
	OriginalPrice string `json:"original_price"`
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

// PaymentInfo represents payment details
type PaymentInfo struct {
	Currency    string `json:"currency"`
	TotalAmount string `json:"total_amount"`
	SubTotal    string `json:"sub_total"`
	ShippingFee string `json:"shipping_fee"`
}

// GetOrderDetail fetches detailed order info including packages
// TikTok API: GET /order/202309/orders
func (c *Client) GetOrderDetail(orderIDs []string) (*OrderDetailResponse, error) {
	apiPath := "/order/202309/orders"
	params := map[string]string{}

	// TikTok accepts comma-separated order IDs
	if len(orderIDs) > 0 {
		ids := ""
		for i, id := range orderIDs {
			if i > 0 {
				ids += ","
			}
			ids += id
		}
		params["ids"] = ids
	}

	var result OrderDetailResponse
	if err := c.doRequest("GET", apiPath, params, &result); err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("TikTok API error: code=%d, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

// =============================================================================
// Handover Time Slots API
// =============================================================================

// HandoverTimeSlotsResponse represents available pickup/drop-off time slots
type HandoverTimeSlotsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		TimeSlots []HandoverTimeSlot `json:"time_slots"`
	} `json:"data"`
}

// HandoverTimeSlot represents a single time slot
type HandoverTimeSlot struct {
	StartTime int64  `json:"start_time"`     // Unix timestamp
	EndTime   int64  `json:"end_time"`       // Unix timestamp
	Type      string `json:"type,omitempty"` // PICKUP, DROP_OFF, etc.
}

// GetHandoverTimeSlots retrieves available pickup/drop-off time slots for a package
// TikTok API: GET /fulfillment/202309/packages/{package_id}/handover_time_slots
func (c *Client) GetHandoverTimeSlots(packageID string) (*HandoverTimeSlotsResponse, error) {
	apiPath := fmt.Sprintf("/fulfillment/202309/packages/%s/handover_time_slots", packageID)
	params := map[string]string{}

	var result HandoverTimeSlotsResponse
	if err := c.doRequest("GET", apiPath, params, &result); err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("TikTok API error: code=%d, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

// GetHandoverTimeSlotsForOrder retrieves available time slots using order ID
// This first resolves the order to package ID, then fetches time slots
// TikTok API: GET /fulfillment/202309/orders/{order_id}/handover_time_slots
func (c *Client) GetHandoverTimeSlotsForOrder(orderID string, lineItemIDs []string) (*HandoverTimeSlotsResponse, error) {
	apiPath := fmt.Sprintf("/fulfillment/202309/orders/%s/handover_time_slots", orderID)
	params := map[string]string{}

	// Add line item IDs if provided
	if len(lineItemIDs) > 0 {
		for i, id := range lineItemIDs {
			params[fmt.Sprintf("order_line_item_ids[%d]", i)] = id
		}
	}

	var result HandoverTimeSlotsResponse
	if err := c.doRequest("GET", apiPath, params, &result); err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("TikTok API error: code=%d, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

// =============================================================================
// Package Detail API
// =============================================================================

// PackageDetailResponse represents package detail response
type PackageDetailResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		ID               string `json:"id"`
		Status           string `json:"status"`
		TrackingNumber   string `json:"tracking_number"`
		ShippingProvider struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"shipping_provider"`
		HandoverTimeSlot *HandoverTimeSlot `json:"handover_time_slot,omitempty"`
		OrderLineItems   []string          `json:"order_line_item_ids,omitempty"`
	} `json:"data"`
}

// GetPackageDetail gets detailed package info
// TikTok API: GET /fulfillment/202309/packages/{package_id}
func (c *Client) GetPackageDetail(packageID string) (*PackageDetailResponse, error) {
	apiPath := fmt.Sprintf("/fulfillment/202309/packages/%s", packageID)
	params := map[string]string{}

	var result PackageDetailResponse
	if err := c.doRequest("GET", apiPath, params, &result); err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("TikTok API error: code=%d, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

// =============================================================================
// Helper Functions
// =============================================================================

// ResolveOrderToPackageID resolves an order ID to its first package ID
// Returns packageID and package status, or error if not found
func (c *Client) ResolveOrderToPackageID(orderID string) (string, string, error) {
	orderDetail, err := c.GetOrderDetail([]string{orderID})
	if err != nil {
		return "", "", fmt.Errorf("failed to get order detail: %w", err)
	}

	if len(orderDetail.Data.Orders) == 0 {
		return "", "", fmt.Errorf("order not found: %s", orderID)
	}

	order := orderDetail.Data.Orders[0]
	if len(order.Packages) == 0 {
		return "", "", fmt.Errorf("no packages found for order: %s", orderID)
	}

	// Return first package (most orders have single package)
	pkg := order.Packages[0]
	return pkg.ID, pkg.Status, nil
}
