package shopeesdk

import (
	"context"
	"errors"
	"time"

	shopee "github.com/omni/backend/pkg/shopee"
)

// GetOrders fetches order list within a time window (<=15 days).
func (c *Client) GetOrders(ctx context.Context, params OrderListParams) (*OrderListResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}

	end := params.To
	if end.IsZero() {
		end = time.Now()
	}
	start := params.From
	if start.IsZero() {
		start = end.AddDate(0, 0, -15)
	}
	timeRangeField := params.TimeRangeField
	if timeRangeField == "" {
		timeRangeField = "create_time"
	}

	return c.api.GetOrderList(start.Unix(), end.Unix(), timeRangeField, params.Status)
}

// GetOrderDetails fetches order details for up to 50 order_sn values.
func (c *Client) GetOrderDetails(ctx context.Context, params OrderDetailParams) (*OrderDetailResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if len(params.OrderSNs) == 0 {
		return nil, errors.New("order_sns is required")
	}
	return c.api.GetOrderDetail(params.OrderSNs)
}

// GetEscrowDetails returns escrow breakdown for orders.
func (c *Client) GetEscrowDetails(ctx context.Context, orderSNs []string) (*GetEscrowDetailsResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if len(orderSNs) == 0 {
		return nil, errors.New("order_sns is required")
	}
	return c.api.GetEscrowDetails(shopee.GetEscrowDetailsRequest{OrderSNList: orderSNs})
}
