package lazadasdk

import "context"

// GetSeller wraps GET /seller/get.
func (c *Client) GetSeller(ctx context.Context) (*RawResponse, error) {
	var resp RawResponse
	if err := c.callRaw(ctx, "GET", "/seller/get", nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
