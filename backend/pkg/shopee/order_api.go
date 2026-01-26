package shopee

import (
	"fmt"
	"strconv"
	"strings"
)

// GetOrderList gets list of orders with pagination support
func (c *Client) GetOrderList(timeFrom, timeTo int64, timeRangeField string, orderStatus string, cursor string) (*GetOrderListResponse, error) {
	path := "/api/v2/order/get_order_list"
	params := map[string]string{
		"time_from":        strconv.FormatInt(timeFrom, 10),
		"time_to":          strconv.FormatInt(timeTo, 10),
		"time_range_field": timeRangeField,
		"page_size":        "100",
	}

	// Add cursor for pagination if provided
	if cursor != "" {
		params["cursor"] = cursor
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
