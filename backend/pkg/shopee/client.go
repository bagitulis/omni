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

// buildURL constructs the full API URL with required params.
// Generates a fresh timestamp and signature per call.
func (c *Client) buildURL(path string, params map[string]string) (string, error) {
	timestamp := time.Now().Unix()
	sign := c.generateSign(path, timestamp)

	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL %s%s: %w", c.baseURL, path, err)
	}
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
	return u.String(), nil
}

// shopeeBaseResponse captures Shopee business-level error fields returned in 200 OK responses.
type shopeeBaseResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// doRequest executes HTTP request and parses response with centralized retry logic.
// Checks both HTTP-level and Shopee business-level errors.
func (c *Client) doRequest(method, path string, params map[string]string, result interface{}) error {
	log.Info().
		Str("method", method).
		Str("path", path).
		Interface("params", params).
		Msg("[Shopee API] Request")

	var resp *http.Response
	var err error
	var lastErr error
	lastStatusCode := 0
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			log.Info().Int("attempt", attempt+1).Dur("backoff", backoff).Msg("[Shopee API] Retrying request")
			time.Sleep(backoff)
		}

		// Rebuild URL each attempt so timestamp + signature are fresh
		reqURL, urlErr := c.buildURL(path, params)
		if urlErr != nil {
			return urlErr
		}

		req, reqErr := http.NewRequest(method, reqURL, nil)
		if reqErr != nil {
			log.Error().Err(reqErr).Msg("[Shopee API] Failed to create request")
			return reqErr
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err = c.httpClient.Do(req)
		if err != nil {
			log.Error().Err(err).Msg("[Shopee API] HTTP request failed")
			lastErr = err
			if attempt < maxRetries {
				continue
			}
			return err
		}
		lastStatusCode = resp.StatusCode

		// Retry on rate limit (429) or server errors (5xx)
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			resp.Body.Close()
			if attempt < maxRetries {
				log.Warn().Int("status", resp.StatusCode).Msg("[Shopee API] Rate limit or server error, retrying")
				continue
			}
		}

		break
	}

	if resp == nil {
		if lastErr != nil {
			return fmt.Errorf("shopee API request failed after %d retries: %w", maxRetries, lastErr)
		}
		return fmt.Errorf("shopee API request failed after %d retries [http_status=%d]", maxRetries, lastStatusCode)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Error().Err(err).Msg("[Shopee API] Failed to read response body")
		return err
	}

	// HTTP-level error
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("shopee API error [http_status=%d]: %s", resp.StatusCode, truncateString(string(body), 2000))
	}

	// Log response for key API paths
	if strings.Contains(path, "logistics") || strings.Contains(path, "shipping") ||
		strings.Contains(path, "get_item_extra_info") || strings.Contains(path, "get_item_base_info") ||
		strings.Contains(path, "get_model_list") {
		log.Info().
			Str("path", path).
			Int("status_code", resp.StatusCode).
			Str("response", truncateString(string(body), 2000)).
			Msg("[Shopee API] Response")
	}

	// Check Shopee business-level error (returned with HTTP 200)
	var baseResp shopeeBaseResponse
	if err := json.Unmarshal(body, &baseResp); err == nil && baseResp.Error != "" {
		log.Warn().
			Str("path", path).
			Str("error", baseResp.Error).
			Str("message", baseResp.Message).
			Msg("[Shopee API] Business error in 200 response")
		return fmt.Errorf("shopee API error: %s — %s", baseResp.Error, baseResp.Message)
	}

	return json.Unmarshal(body, result)
}

// GetItemList fetches product items via Shopee API v2
// API: GET /api/v2/product/get_item_list
func (c *Client) GetItemList(offset, pageSize int, itemStatus string) (map[string]interface{}, error) {
	params := map[string]string{
		"offset":      strconv.Itoa(offset),
		"page_size":   strconv.Itoa(pageSize),
		"item_status": itemStatus,
	}

	var result map[string]interface{}
	if err := c.doRequest("GET", "/api/v2/product/get_item_list", params, &result); err != nil {
		return nil, err
	}
	return result, nil
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
