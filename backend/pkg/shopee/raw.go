package shopee

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// RawRequest executes arbitrary Shopee API calls with proper signing.
// Path should start with "/api/v2/...".
func (c *Client) RawRequest(ctx context.Context, method, path string, params map[string]string, body []byte, result interface{}) error {
	if method == "" {
		method = http.MethodGet
	}

	reqURL := c.buildURL(path, params)
	var reader *bytes.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	} else {
		reader = bytes.NewReader(nil)
	}

	req, err := http.NewRequestWithContext(ctx, method, reqURL, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if result == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(result)
}

// RawGet executes a GET request and decodes into result.
func (c *Client) RawGet(ctx context.Context, path string, params map[string]string, result interface{}) error {
	return c.RawRequest(ctx, http.MethodGet, path, params, nil, result)
}

// RawPost executes a POST request with JSON body.
func (c *Client) RawPost(ctx context.Context, path string, params map[string]string, payload interface{}, result interface{}) error {
	var body []byte
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}
		body = data
	}
	return c.RawRequest(ctx, http.MethodPost, path, params, body, result)
}
