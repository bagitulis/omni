package lazadasdk

import "context"

// GetShipmentProviders wraps GET /shipment/providers/get.
func (c *Client) GetShipmentProviders(ctx context.Context) (*RawResponse, error) {
	var resp RawResponse
	if err := c.callRaw(ctx, "GET", "/shipment/providers/get", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
