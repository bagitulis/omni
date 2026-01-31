package shopee

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	BaseURLProduction = "https://partner.shopeemobile.com"
	BaseURLTest       = "https://partner.test-stable.shopeemobile.com"
)

// Client is the Shopee API client
type Client struct {
	partnerID   int64
	partnerKey  string
	shopID      int64
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

// NewClient creates a new Shopee API client
func NewClient(partnerID int64, partnerKey string, isProduction bool) *Client {
	baseURL := BaseURLTest
	if isProduction {
		baseURL = BaseURLProduction
	}

	return &Client{
		partnerID:  partnerID,
		partnerKey: partnerKey,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetShopCredentials sets shop-level credentials
func (c *Client) SetShopCredentials(shopID int64, accessToken string) {
	c.shopID = shopID
	c.accessToken = accessToken
}

// generateSign creates HMAC-SHA256 signature for API calls
func (c *Client) generateSign(path string, timestamp int64) string {
	baseString := fmt.Sprintf("%d%s%d%s%d",
		c.partnerID, path, timestamp, c.accessToken, c.shopID)

	h := hmac.New(sha256.New, []byte(c.partnerKey))
	h.Write([]byte(baseString))
	return hex.EncodeToString(h.Sum(nil))
}

// buildURL constructs the full API URL with required params
func (c *Client) buildURL(path string, params map[string]string) string {
	timestamp := time.Now().Unix()
	sign := c.generateSign(path, timestamp)

	u, _ := url.Parse(c.baseURL + path)
	q := u.Query()
	q.Set("partner_id", strconv.FormatInt(c.partnerID, 10))
	q.Set("timestamp", strconv.FormatInt(timestamp, 10))
	q.Set("sign", sign)
	q.Set("shop_id", strconv.FormatInt(c.shopID, 10))
	q.Set("access_token", c.accessToken)

	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// doRequest executes HTTP request and parses response
func (c *Client) doRequest(method, path string, params map[string]string, result interface{}) error {
	reqURL := c.buildURL(path, params)

	req, err := http.NewRequest(method, reqURL, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Debug: log raw response for item-related APIs
	if strings.Contains(path, "get_item_extra_info") || strings.Contains(path, "get_item_base_info") {
		log.Printf("[Shopee API Debug] %s response (truncated): %s", path, truncateString(string(body), 500))
	}

	return json.Unmarshal(body, result)
}

// GetShipmentInfo gets shipment info for an order (stub - implementing interface)
func (c *Client) GetShipmentInfo(ctx context.Context, orderSN string) (map[string]interface{}, error) {
	path := "/api/v2/logistics/get_shipment_info"
	params := map[string]string{
		"order_sn": orderSN,
	}

	var result map[string]interface{}
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetShippingOptions gets shipping parameters for an order
func (c *Client) GetShippingOptions(ctx context.Context, orderSn string) (map[string]interface{}, error) {
	path := "/api/v2/logistics/get_shipping_parameter"
	params := map[string]string{
		"order_sn": orderSn,
	}

	var result map[string]interface{}
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetTrackingInfo gets tracking number for an order
func (c *Client) GetTrackingInfo(ctx context.Context, orderSn string) (map[string]interface{}, error) {
	path := "/api/v2/logistics/get_tracking_number"
	params := map[string]string{
		"order_sn": orderSn,
	}

	var result map[string]interface{}
	err := c.doRequest("GET", path, params, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// truncateString truncates string to maxLen
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
