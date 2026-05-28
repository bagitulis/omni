package shopee

import (
	"context"
	"fmt"
	"strconv"
)

// --- Request Structs ---

// GetBookingListRequest represents request for booking list
type GetBookingListRequest struct {
	TimeRangeField string `json:"time_range_field"`
	TimeFrom       int64  `json:"time_from"`
	TimeTo         int64  `json:"time_to"`
	PageSize       int    `json:"page_size"`
	Cursor         string `json:"cursor,omitempty"`
	BookingStatus  string `json:"booking_status,omitempty"`
}

// GetBookingDetailRequest represents request for booking detail
type GetBookingDetailRequest struct {
	BookingSNList string `json:"booking_sn_list"` // comma separated, max 50
}

// --- Response Structs ---

// GetBookingListResponse represents response for booking list
type GetBookingListResponse struct {
	Response struct {
		BookingList []BookingBasic `json:"booking_list"`
		More        bool           `json:"more"`
		NextCursor  string         `json:"next_cursor"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// GetBookingDetailResponse represents response for booking detail
type GetBookingDetailResponse struct {
	Response struct {
		BookingList []BookingDetail `json:"booking_list"`
	} `json:"response"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

// --- Data Structs ---

// BookingBasic represents basic booking info
type BookingBasic struct {
	BookingSn     string `json:"booking_sn"`
	OrderSn       string `json:"order_sn"`
	BookingStatus string `json:"booking_status"`
}

// BookingDetail represents detailed booking info
type BookingDetail struct {
	BookingSn        string              `json:"booking_sn"`
	OrderSn          string              `json:"order_sn"`
	Region           string              `json:"region"`
	BookingStatus    string              `json:"booking_status"`
	MatchStatus      string              `json:"match_status"`
	ShippingCarrier  string              `json:"shipping_carrier"`
	CreateTime       int64               `json:"create_time"`
	UpdateTime       int64               `json:"update_time"`
	RecipientAddress *RecipientAddress   `json:"recipient_address"`
	ItemList         []BookingItemDetail `json:"item_list"`
	Dropshipper      string              `json:"dropshipper"`
	DropshipperPhone string              `json:"dropshipper_phone"`
	CancelBy         string              `json:"cancel_by"`
	CancelReason     string              `json:"cancel_reason"`
	FulfillmentFlag  string              `json:"fulfillment_flag"`
	PickupDoneTime   int64               `json:"pickup_done_time"`
}

// RecipientAddress represents recipient address info
type RecipientAddress struct {
	Name        string `json:"name"`
	Phone       string `json:"phone"`
	Town        string `json:"town"`
	District    string `json:"district"`
	City        string `json:"city"`
	State       string `json:"state"`
	Region      string `json:"region"`
	Zipcode     string `json:"zipcode"`
	FullAddress string `json:"full_address"`
}

// BookingItemDetail represents an item in a booking
type BookingItemDetail struct {
	ItemID            int64         `json:"item_id"`
	ModelID           int64         `json:"model_id"`
	ItemName          string        `json:"item_name"`
	ItemSku           string        `json:"item_sku"`
	ModelName         string        `json:"model_name"`
	ModelSku          string        `json:"model_sku"`
	ModelQuantity     int           `json:"model_quantity_purchased"`
	Weight            float64       `json:"weight"`
	ProductLocationId string        `json:"product_location_id"`
	ImageInfo         ItemImageInfo `json:"image_info"`
}

// GetBookingList gets list of bookings with pagination support
func (c *Client) GetBookingList(ctx context.Context, req *GetBookingListRequest) (*GetBookingListResponse, error) {
	path := "/api/v2/order/get_booking_list"
	params := map[string]string{
		"time_range_field": req.TimeRangeField,
		"time_from":        strconv.FormatInt(req.TimeFrom, 10),
		"time_to":          strconv.FormatInt(req.TimeTo, 10),
		"page_size":        strconv.Itoa(req.PageSize),
	}

	if req.Cursor != "" {
		params["cursor"] = req.Cursor
	}
	if req.BookingStatus != "" {
		params["booking_status"] = req.BookingStatus
	}

	var result GetBookingListResponse
	err := c.doRequest(ctx, "GET", path, params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetBookingDetail gets booking details by booking SN list
// booking_sn_list is a comma-separated string, max 50
func (c *Client) GetBookingDetail(ctx context.Context, req *GetBookingDetailRequest) (*GetBookingDetailResponse, error) {
	path := "/api/v2/order/get_booking_detail"
	params := map[string]string{
		"booking_sn_list": req.BookingSNList,
	}

	var result GetBookingDetailResponse
	err := c.doRequest(ctx, "GET", path, params, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
