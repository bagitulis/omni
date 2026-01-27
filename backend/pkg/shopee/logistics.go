package shopee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// ShipOrderRequest represents request to ship an order
type ShipOrderRequest struct {
	OrderSN       string             `json:"order_sn"`
	PackageNumber string             `json:"package_number,omitempty"`
	Pickup        *PickupInfo        `json:"pickup,omitempty"`
	Dropoff       *DropoffInfo       `json:"dropoff,omitempty"`
	NonIntegrated *NonIntegratedInfo `json:"non_integrated,omitempty"`
}

// PickupInfo represents pickup details
type PickupInfo struct {
	AddressID    int64  `json:"address_id"`
	PickupTimeID string `json:"pickup_time_id,omitempty"`
}

// DropoffInfo represents dropoff details
type DropoffInfo struct {
	BranchID       int64  `json:"branch_id,omitempty"`
	SenderRealName string `json:"sender_real_name,omitempty"`
	TrackingNo     string `json:"tracking_no,omitempty"`
	Slug           string `json:"slug,omitempty"`
}

// NonIntegratedInfo for non-integrated logistics
type NonIntegratedInfo struct {
	TrackingNumber string `json:"tracking_number"`
}

// ShipOrderResponse represents ship order API response
type ShipOrderResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		OrderSN string `json:"order_sn"`
	} `json:"response"`
}

// CancelOrderRequest represents request to cancel order
type CancelOrderRequest struct {
	OrderSN      string       `json:"order_sn"`
	CancelReason string       `json:"cancel_reason"`
	ItemList     []CancelItem `json:"item_list,omitempty"`
}

// CancelItem represents item to cancel
type CancelItem struct {
	ItemID  int64 `json:"item_id"`
	ModelID int64 `json:"model_id"`
}

// CancelOrderResponse represents cancel order API response
type CancelOrderResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		OrderSN    string `json:"order_sn"`
		UpdateTime int64  `json:"update_time"`
	} `json:"response"`
}

// GetShippingParameterResponse represents shipping parameter response
type GetShippingParameterResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		InfoNeeded struct {
			Pickup  []PickupAddressInfo `json:"pickup,omitempty"`
			Dropoff []BranchInfo        `json:"dropoff,omitempty"`
		} `json:"info_needed"`
	} `json:"response"`
}

// PickupAddressInfo represents pickup address
type PickupAddressInfo struct {
	AddressID    int64      `json:"address_id"`
	Address      string     `json:"address"`
	TimeSlotList []TimeSlot `json:"time_slot_list"`
}

// TimeSlot represents time slot for pickup
type TimeSlot struct {
	PickupTimeID string `json:"pickup_time_id"`
	Date         string `json:"date"`
	TimeText     string `json:"time_text"`
}

// BranchInfo represents dropoff branch
type BranchInfo struct {
	BranchID int64  `json:"branch_id"`
	Address  string `json:"address"`
	City     string `json:"city"`
	State    string `json:"state"`
}

// GetTrackingNumberResponse represents tracking number response
type GetTrackingNumberResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		TrackingNumber string `json:"tracking_number"`
		Hint           string `json:"hint,omitempty"`
	} `json:"response"`
}

// ShipOrder ships an order via Shopee API
func (c *Client) ShipOrder(req ShipOrderRequest) (*ShipOrderResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result ShipOrderResponse
	if err := c.doPostRequest("/api/v2/logistics/ship_order", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// CancelOrder cancels an order via Shopee API
func (c *Client) CancelOrder(req CancelOrderRequest) (*CancelOrderResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	var result CancelOrderResponse
	if err := c.doPostRequest("/api/v2/order/cancel_order", nil, body, &result); err != nil {
		return nil, err
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetShippingParameter gets shipping parameter for an order
func (c *Client) GetShippingParameter(orderSN string) (*GetShippingParameterResponse, error) {
	params := map[string]string{
		"order_sn": orderSN,
	}

	var result GetShippingParameterResponse
	if err := c.doRequest("GET", "/api/v2/logistics/get_shipping_parameter", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// GetTrackingNumber gets tracking number for an order
func (c *Client) GetTrackingNumber(orderSN string) (*GetTrackingNumberResponse, error) {
	params := map[string]string{
		"order_sn": orderSN,
	}

	var result GetTrackingNumberResponse
	if err := c.doRequest("GET", "/api/v2/logistics/get_tracking_number", params, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

// doPostRequest executes POST request with body
func (c *Client) doPostRequest(path string, params map[string]string, body []byte, result interface{}) error {
	reqURL := c.buildURL(path, params)

	// 🔍 LOG REQUEST
	log.Printf("[Shopee API] 📤 POST %s", path)
	log.Printf("[Shopee API] 📦 Request Body: %s", string(body))

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))
	if err != nil {
		log.Printf("[Shopee API] ❌ Failed to create request: %v", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Shopee API] ❌ HTTP request failed: %v", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body for logging
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("[Shopee API] ❌ Failed to read response: %v", err)
		return err
	}

	// 🔍 LOG RESPONSE
	log.Printf("[Shopee API] 📥 Response Status: %s", resp.Status)
	log.Printf("[Shopee API] 📥 Response Body: %s", string(respBody))

	// Decode response
	if err := json.Unmarshal(respBody, result); err != nil {
		log.Printf("[Shopee API] ❌ Failed to parse response: %v", err)
		return err
	}

	log.Printf("[Shopee API] ✅ Request completed successfully")
	return nil
}
