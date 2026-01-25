package lazadasdk

import (
	"context"
	"strings"
)

// CallRaw executes any Lazada endpoint and decodes into out.
func (c *Client) CallRaw(ctx context.Context, call RawCall, out interface{}) error {
	method := strings.ToUpper(call.Method)
	if method == "" {
		method = "GET"
	}
	return c.callRaw(ctx, method, call.Path, call.Params, out)
}

// CallRawMap is a convenience helper to decode into map[string]interface{}.
func (c *Client) CallRawMap(ctx context.Context, call RawCall) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := c.CallRaw(ctx, call, &result); err != nil {
		return nil, err
	}
	return result, nil
}
