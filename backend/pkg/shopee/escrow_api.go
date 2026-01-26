package shopee

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// GetEscrowDetails gets escrow details for orders using batch endpoint
// Uses POST request with order_sn_list in JSON body
func (c *Client) GetEscrowDetails(req GetEscrowDetailsRequest) (*GetEscrowDetailsResponse, error) {
	path := "/api/v2/payment/get_escrow_detail_batch"

	// Build request body
	bodyData, _ := json.Marshal(map[string][]string{
		"order_sn_list": req.OrderSNList,
	})

	var result GetEscrowDetailsResponse
	err := c.doPostRequestWithBody(path, bodyData, &result)
	if err != nil {
		return nil, err
	}

	// Check for API error in response
	if result.Error != "" {
		return nil, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetOrderList returns flattened order list from batch response
func (r *GetEscrowDetailsResponse) GetOrderList() []EscrowOrder {
	var orders []EscrowOrder
	for _, wrapper := range r.Response {
		if wrapper.EscrowDetail != nil {
			orders = append(orders, *wrapper.EscrowDetail)
		}
	}
	return orders
}

// GetWalletBalance gets wallet balance (implements APIClient interface)
func (c *Client) GetWalletBalance(ctx context.Context) (map[string]interface{}, error) {
	path := "/api/v2/payment/get_wallet_transaction_list"
	params := map[string]string{}

	var result map[string]interface{}
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// doPostRequestWithBody executes HTTP POST request with body (no query params)
func (c *Client) doPostRequestWithBody(path string, body []byte, result interface{}) error {
	reqURL := c.buildURL(path, nil)

	req, err := http.NewRequest("POST", reqURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(respBody, result)
}
