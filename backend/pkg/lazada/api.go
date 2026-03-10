package lazada

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// FlexibleString handles JSON that can be either string or number
// Lazada API returns item_id as number, but we need it as string
type FlexibleString string

// UnmarshalJSON implements json.Unmarshaler for FlexibleString
func (f *FlexibleString) UnmarshalJSON(data []byte) error {
	// Try to unmarshal as string first
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*f = FlexibleString(s)
		return nil
	}

	// Try to unmarshal as number
	var n json.Number
	if err := json.Unmarshal(data, &n); err == nil {
		*f = FlexibleString(n.String())
		return nil
	}

	// Try to unmarshal as int64 directly
	var i int64
	if err := json.Unmarshal(data, &i); err == nil {
		*f = FlexibleString(strconv.FormatInt(i, 10))
		return nil
	}

	return fmt.Errorf("FlexibleString: cannot unmarshal %s", string(data))
}

// String returns the string value
func (f FlexibleString) String() string {
	return string(f)
}

// OrderListResponse represents Lazada order list response
type OrderListResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Count  int `json:"count"`
		Orders []struct {
			OrderID          string  `json:"order_id"`
			OrderNumber      string  `json:"order_number"`
			Status           string  `json:"status"`
			Price            float64 `json:"price"`
			CustomerName     string  `json:"customer_first_name"`
			PromisedShipDate string  `json:"promised_shipping_times"` // Shipping deadline (ISO date)
			ShippingType     string  `json:"delivery_info"`           // Shipping carrier/type
			CreatedAt        string  `json:"created_at"`              // Order creation time
			UpdatedAt        string  `json:"updated_at"`              // Order update time
		} `json:"orders"`
	} `json:"data"`
}

// GetOrders fetches orders from Lazada API
func (c *Client) GetOrders(status string, offset, limit int) (*OrderListResponse, error) {
	params := map[string]string{
		"status": status,
		"offset": fmt.Sprintf("%d", offset),
		"limit":  fmt.Sprintf("%d", limit),
	}

	var result OrderListResponse
	err := c.doRequest("GET", "/orders/get", params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Code != "0" && result.Code != "" {
		return nil, fmt.Errorf("lazada API error (code %s): %s", result.Code, result.Message)
	}

	return &result, nil
}
