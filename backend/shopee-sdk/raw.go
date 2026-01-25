package shopeesdk

import (
	"context"
	"encoding/json"
)

// CallRaw executes any Shopee API path with proper signing.
// Useful for endpoints not yet modeled explicitly (ads, discounts, vouchers, bundles, etc.).
func (c *Client) CallRaw(ctx context.Context, call RawCall, out interface{}) error {
	if err := apiReady(ctx); err != nil {
		return err
	}
	method := call.Method
	if method == "" {
		method = "GET"
	}

	var bodyBytes []byte
	if call.Body != nil {
		data, err := json.Marshal(call.Body)
		if err != nil {
			return err
		}
		bodyBytes = data
	}

	if method == "GET" {
		return c.api.RawGet(ctx, call.Path, call.Query, out)
	}
	return c.api.RawRequest(ctx, method, call.Path, call.Query, bodyBytes, out)
}

// CallRawMap is a convenience to decode into map[string]interface{}.
func (c *Client) CallRawMap(ctx context.Context, call RawCall) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.CallRaw(ctx, call, &result); err != nil {
		return nil, err
	}
	return result, nil
}
