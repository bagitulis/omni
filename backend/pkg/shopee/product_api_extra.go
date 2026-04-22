package shopee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// GetWalletTransactionRequest represents wallet transaction request
type GetWalletTransactionRequest struct {
	TransactionTypes   []string `json:"-"`
	TransactionTabType string   `json:"-"`
	MoneyFlow          string   `json:"-"`
	PageNo             int      `json:"-"`
	PageSize           int      `json:"-"`
	StartDate          int64    `json:"-"`
	EndDate            int64    `json:"-"`
}

// WalletTransactionResponse represents wallet transaction response
type WalletTransactionResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		TransactionList []WalletTransaction `json:"transaction_list"`
		More            bool                `json:"more"`
	} `json:"response"`
}

// WalletTransaction represents a wallet transaction
type WalletTransaction struct {
	TransactionID   int64   `json:"transaction_id"`
	Status          string  `json:"status"`
	TransactionType string  `json:"transaction_type"`
	Amount          float64 `json:"amount"`
	CreateTime      int64   `json:"create_time"`
	OrderSN         string  `json:"order_sn,omitempty"`
	RefundSN        string  `json:"refund_sn,omitempty"`
	Description     string  `json:"description,omitempty"`
	BuyerUsername   string  `json:"buyer_user_name,omitempty"`
}

// GetWalletTransactions gets wallet transactions using GET request with query params
func (c *Client) GetWalletTransactions(req GetWalletTransactionRequest) (*WalletTransactionResponse, error) {
	params := map[string]string{
		"page_no":   strconv.Itoa(req.PageNo),
		"page_size": strconv.Itoa(req.PageSize),
	}

	if req.StartDate > 0 {
		params["create_time_from"] = strconv.FormatInt(req.StartDate, 10)
	}
	if req.EndDate > 0 {
		params["create_time_to"] = strconv.FormatInt(req.EndDate, 10)
	}
	if req.TransactionTabType != "" {
		params["transaction_tab_type"] = req.TransactionTabType
	}
	if req.MoneyFlow != "" {
		params["money_flow"] = req.MoneyFlow
	}

	reqURL, urlErr := c.buildURL("/api/v2/payment/get_wallet_transaction_list", params)
	if urlErr != nil {
		return nil, urlErr
	}
	httpReq, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	var result WalletTransactionResponse
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func createPostRequest(url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

// CategoryRecommendResponse represents category recommendation response
type CategoryRecommendResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Warning  string `json:"warning"`
	Response struct {
		CategoryID []int64 `json:"category_id"`
	} `json:"response"`
}

// GetCategoryRecommend gets category recommendation by product name
func (c *Client) GetCategoryRecommend(itemName string, imageID string) (*CategoryRecommendResponse, error) {
	params := map[string]string{
		"item_name": itemName,
	}
	if imageID != "" {
		params["product_cover_image"] = imageID
	}

	var result CategoryRecommendResponse
	if err := c.doRequest("GET", "/api/v2/product/category_recommend", params, &result); err != nil {
		return nil, fmt.Errorf("category recommend request failed: %w", err)
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// RecommendAttributeResponse represents recommend attribute response
type RecommendAttributeResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Warning  string `json:"warning"`
	Response struct {
		AttributeList []RecommendedAttribute `json:"attribute_list"`
	} `json:"response"`
}

// RecommendedAttribute represents a recommended attribute
type RecommendedAttribute struct {
	AttributeID      int64                  `json:"attribute_id"`
	OriginalAttrName string                 `json:"original_attribute_name"`
	DisplayAttrName  string                 `json:"display_attribute_name"`
	IsMandatory      bool                   `json:"is_mandatory"`
	InputType        string                 `json:"input_type"`
	FormatType       string                 `json:"format_type"`
	AttributeUnit    []string               `json:"attribute_unit"`
	AttributeValues  []RecommendedAttrValue `json:"attribute_value_list"`
}

// RecommendedAttrValue represents attribute value
type RecommendedAttrValue struct {
	ValueID         int64  `json:"value_id"`
	OriginalValName string `json:"original_value_name"`
	DisplayValName  string `json:"display_value_name"`
	ValueUnit       string `json:"value_unit,omitempty"`
}

// GetRecommendAttribute gets recommended attributes for a category
func (c *Client) GetRecommendAttribute(categoryID int64, itemName string) (*RecommendAttributeResponse, error) {
	params := map[string]string{
		"category_id": fmt.Sprintf("%d", categoryID),
		"item_name":   itemName,
	}

	var result RecommendAttributeResponse
	if err := c.doRequest("GET", "/api/v2/product/get_recommend_attribute", params, &result); err != nil {
		return nil, fmt.Errorf("get recommend attribute request failed: %w", err)
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}

// GetLogisticChannelsResponse represents logistics channels response
type GetLogisticChannelsResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		LogisticsList []LogisticChannel `json:"logistics_channel_list"`
	} `json:"response"`
}

// LogisticChannel represents a logistics channel
type LogisticChannel struct {
	LogisticID   int64  `json:"logistics_channel_id"`
	LogisticName string `json:"logistics_channel_name"`
	Enabled      bool   `json:"enabled"`
	Preferred    bool   `json:"preferred"`
}

// GetLogisticChannels gets available logistics channels for shop
func (c *Client) GetLogisticChannels() (*GetLogisticChannelsResponse, error) {
	var result GetLogisticChannelsResponse
	if err := c.doRequest("GET", "/api/v2/logistics/get_channel_list", nil, &result); err != nil {
		return nil, fmt.Errorf("get logistics channels failed: %w", err)
	}

	if result.Error != "" {
		return &result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return &result, nil
}
