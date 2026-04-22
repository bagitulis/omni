package shopee

import (
	"encoding/json"
	"fmt"

	"github.com/rs/zerolog/log"
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

// GetShippingParameterResponse represents shipping parameter response from Shopee API
type GetShippingParameterResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		InfoNeeded struct {
			Pickup  []string `json:"pickup,omitempty"`
			Dropoff []string `json:"dropoff,omitempty"`
		} `json:"info_needed"`
		Pickup struct {
			AddressList []PickupAddressInfo `json:"address_list,omitempty"`
		} `json:"pickup,omitempty"`
		Dropoff struct {
			BranchList []BranchInfo `json:"branch_list,omitempty"`
		} `json:"dropoff,omitempty"`
	} `json:"response"`
}

// PickupAddressInfo represents pickup address from Shopee API
type PickupAddressInfo struct {
	AddressID    int64      `json:"address_id"`
	Region       string     `json:"region,omitempty"`
	State        string     `json:"state,omitempty"`
	City         string     `json:"city,omitempty"`
	District     string     `json:"district,omitempty"`
	Town         string     `json:"town,omitempty"`
	Address      string     `json:"address"`
	Zipcode      string     `json:"zipcode,omitempty"`
	AddressFlag  []string   `json:"address_flag,omitempty"`
	TimeSlotList []TimeSlot `json:"time_slot_list"`
}

// TimeSlot represents time slot for pickup
type TimeSlot struct {
	PickupTimeID string   `json:"pickup_time_id"`
	Date         int64    `json:"date"`
	TimeText     string   `json:"time_text,omitempty"`
	Flags        []string `json:"flags,omitempty"`
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

	if result.Error != "" {
		log.Warn().Str("error", result.Error).Str("message", result.Message).Msg("[Shopee API] GetShippingParameter error")
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
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
