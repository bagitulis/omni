package shopee

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

// GetOrderList gets list of orders with pagination support
func (c *Client) GetOrderList(ctx context.Context, timeFrom, timeTo int64, timeRangeField string, orderStatus string, cursor string) (*GetOrderListResponse, error) {
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

// GetOrderDetail gets order details by order SNs
// Includes buyer info, payment, shipping via response_optional_fields
func (c *Client) GetOrderDetail(ctx context.Context, orderSNList []string) (*GetOrderDetailResponse, error) {
	path := "/api/v2/order/get_order_detail"
	// Request all important optional fields from Shopee API
	optionalFields := strings.Join([]string{
		"buyer_user_id",
		"buyer_username",
		"recipient_address",
		"actual_shipping_fee",
		"goods_to_declare",
		"note",
		"note_update_time",
		"item_list",
		"pay_time",
		"dropshipper",
		"dropshipper_phone",
		"split_up",
		"buyer_cancel_reason",
		"cancel_by",
		"cancel_reason",
		"actual_shipping_fee_confirmed",
		"buyer_cpf_id",
		"fulfillment_flag",
		"pickup_done_time",
		"package_list",
		"shipping_carrier",
		"payment_method",
		"total_amount",
		"ship_by_date",
		"message_to_seller",
		"invoice_data",
		"checkout_shipping_carrier",
		"reverse_shipping_fee",
		"order_chargeable_weight_gram",
		"edt",
		"prescription_images",
		"prescription_check_status",
	}, ",")
	params := map[string]string{
		"order_sn_list":            strings.Join(orderSNList, ","),
		"response_optional_fields": optionalFields,
	}

	var result GetOrderDetailResponse
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
