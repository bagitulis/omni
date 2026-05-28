package shopeesdk

import (
	"context"
	"errors"
)

// ShipOrder triggers shipment with pickup/dropoff params.
func (c *Client) ShipOrder(ctx context.Context, req ShipOrderRequest) (*ShipOrderResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if req.OrderSN == "" {
		return nil, errors.New("order_sn is required")
	}
	return c.api.ShipOrder(req)
}

// CancelOrder cancels an order.
func (c *Client) CancelOrder(ctx context.Context, req CancelOrderRequest) (*CancelOrderResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if req.OrderSN == "" {
		return nil, errors.New("order_sn is required")
	}
	return c.api.CancelOrder(req)
}

// GetShippingParameter lists shipping options for an order.
func (c *Client) GetShippingParameter(ctx context.Context, orderSN string) (*GetShippingParameterResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if orderSN == "" {
		return nil, errors.New("order_sn is required")
	}
	return c.api.GetShippingParameter(ctx, orderSN)
}

// GetTrackingNumber fetches tracking number for an order.
func (c *Client) GetTrackingNumber(ctx context.Context, orderSN string) (*GetTrackingNumberResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if orderSN == "" {
		return nil, errors.New("order_sn is required")
	}
	return c.api.GetTrackingNumber(ctx, orderSN)
}
