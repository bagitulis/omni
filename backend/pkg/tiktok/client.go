package tiktok

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	BaseURL = "https://open-api.tiktokglobalshop.com"
)

// Client is the TikTok Shop API client
type Client struct {
	appKey      string
	appSecret   string
	accessToken string
	shopCipher  string
	httpClient  *http.Client
}

// NewClient creates a new TikTok Shop API client
func NewClient(appKey, appSecret string) *Client {
	return &Client{
		appKey:     appKey,
		appSecret:  appSecret,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetCredentials sets access token and shop cipher
func (c *Client) SetCredentials(accessToken, shopCipher string) {
	c.accessToken = accessToken
	c.shopCipher = shopCipher
}

// generateSign creates signature for TikTok API (without body)
// Format: HMAC-SHA256(appSecret, appSecret + path + sortedParams + appSecret)
func (c *Client) generateSign(path string, params map[string]string) string {
	return c.generateSignWithBody(path, params, nil)
}

// generateSignWithBody creates signature for TikTok API with raw body bytes
// This matches the official TikTok SDK signature algorithm
// Format: HMAC-SHA256(appSecret, appSecret + path + sortedParams + bodyBytes + appSecret)
func (c *Client) generateSignWithBody(path string, params map[string]string, bodyBytes []byte) string {
	// Sort keys, excluding 'sign' and 'access_token'
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" && k != "access_token" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build param string: key1value1key2value2...
	var paramString strings.Builder
	for _, k := range keys {
		paramString.WriteString(k)
		paramString.WriteString(params[k])
	}

	// Build sign string: path + sortedParams + body (matches official SDK)
	var signString strings.Builder
	signString.WriteString(path)
	signString.WriteString(paramString.String())

	// Append raw body bytes if present (critical: use exact bytes, not re-marshaled)
	if len(bodyBytes) > 0 {
		signString.Write(bodyBytes)
	}

	// Wrap with appSecret: appSecret + signString + appSecret
	finalInput := c.appSecret + signString.String() + c.appSecret

	// HMAC-SHA256
	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(finalInput))
	return hex.EncodeToString(h.Sum(nil))
}

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}

// doRequest executes HTTP request (GET without body) with retry logic
func (c *Client) doRequest(method, apiPath string, params map[string]string, result interface{}) error {
	const maxRetries = 3

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
		}

		timestamp := time.Now().Unix()

		// Add common params (refresh timestamp on each retry)
		params["app_key"] = c.appKey
		params["timestamp"] = fmt.Sprintf("%d", timestamp)
		if c.accessToken != "" {
			params["access_token"] = c.accessToken
		}
		if c.shopCipher != "" {
			params["shop_cipher"] = c.shopCipher
		}

		// Generate signature (no body for GET requests)
		params["sign"] = c.generateSign(apiPath, params)

		// Build URL
		u, _ := url.Parse(BaseURL + apiPath)
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(method, u.String(), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if c.accessToken != "" {
			req.Header.Set("x-tts-access-token", c.accessToken)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue // Network error — retry
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Retry on 429 (rate limit) or 5xx (server error)
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			log.Warn().Str("method", method).Str("api_path", apiPath).Int("status_code", resp.StatusCode).Msg("TikTok API rate limit or server error, retrying")
			lastErr = fmt.Errorf("TikTok API error: status %d", resp.StatusCode)
			continue
		}

		// Parse request_id from response
		var baseResp struct {
			RequestID string `json:"request_id"`
		}
		json.Unmarshal(body, &baseResp)

		// Unmarshal into result
		if err := json.Unmarshal(body, result); err != nil {
			log.Warn().Str("method", method).Str("api_path", apiPath).Str("request_id", baseResp.RequestID).Str("raw_body", truncateString(string(body), 2000)).Msg("TikTok API unmarshal failed")
			return err
		}

		log.Debug().Str("method", method).Str("api_path", apiPath).Str("request_id", baseResp.RequestID).Msg("TikTok API request successful")
		return nil
}

// doRequestWithBody executes HTTP request with JSON body and retry logic
func (c *Client) doRequestWithBody(method, apiPath string, params map[string]string, body interface{}, result interface{}) error {
	const maxRetries = 3

	// Marshal body ONCE (same bytes for signature AND request across retries)
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal body: %w", err)
		}
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<attempt) * 100 * time.Millisecond)
		}

		timestamp := time.Now().Unix()

		// Add common params (refresh timestamp on each retry)
		params["app_key"] = c.appKey
		params["timestamp"] = fmt.Sprintf("%d", timestamp)
		if c.accessToken != "" {
			params["access_token"] = c.accessToken
		}
		if c.shopCipher != "" {
			params["shop_cipher"] = c.shopCipher
		}

		// Generate signature using raw body bytes
		params["sign"] = c.generateSignWithBody(apiPath, params, bodyBytes)

		// Build URL
		u, _ := url.Parse(BaseURL + apiPath)
		q := u.Query()
		for k, v := range params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()

		req, err := http.NewRequest(method, u.String(), strings.NewReader(string(bodyBytes)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if c.accessToken != "" {
			req.Header.Set("x-tts-access-token", c.accessToken)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue // Network error — retry
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		// Retry on 429 (rate limit) or 5xx (server error)
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			log.Warn().Str("method", method).Str("api_path", apiPath).Int("status_code", resp.StatusCode).Msg("TikTok API rate limit or server error, retrying")
			lastErr = fmt.Errorf("TikTok API error: status %d", resp.StatusCode)
			continue
		}

		// Parse request_id from response
		var baseResp struct {
			RequestID string `json:"request_id"`
		}
		json.Unmarshal(respBody, &baseResp)

		// Unmarshal into result
		if err := json.Unmarshal(respBody, result); err != nil {
			log.Warn().Str("method", method).Str("api_path", apiPath).Str("request_id", baseResp.RequestID).Str("raw_body", truncateString(string(respBody), 2000)).Msg("TikTok API unmarshal failed")
			return err
		}

		log.Debug().Str("method", method).Str("api_path", apiPath).Str("request_id", baseResp.RequestID).Msg("TikTok API request successful")
		return nil
}

// DoGet executes a GET request
func (c *Client) DoGet(apiPath string, params map[string]string, result interface{}) error {
	if params == nil {
		params = make(map[string]string)
	}
	return c.doRequest("GET", apiPath, params, result)
}

// DoPost executes a POST request with JSON body
func (c *Client) DoPost(apiPath string, params map[string]string, body interface{}, result interface{}) error {
	if params == nil {
		params = make(map[string]string)
	}
	return c.doRequestWithBody("POST", apiPath, params, body, result)
}

// DoPut executes a PUT request with JSON body
func (c *Client) DoPut(apiPath string, params map[string]string, body interface{}, result interface{}) error {
	if params == nil {
		params = make(map[string]string)
	}
	return c.doRequestWithBody("PUT", apiPath, params, body, result)
}

// ShippingDocumentResponse represents the TikTok shipping document API response
type ShippingDocumentResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DocURL         string `json:"doc_url"`
		TrackingNumber string `json:"tracking_number"`
	} `json:"data"`
}

// GetShippingDocument retrieves shipping label/document URL for a package
// documentType: "SHIPPING_LABEL", "PACKING_SLIP", etc.
func (c *Client) GetShippingDocument(packageID, documentType string) (string, error) {
	apiPath := fmt.Sprintf("/fulfillment/202309/packages/%s/shipping_documents", packageID)
	params := map[string]string{
		"document_type": documentType,
	}

	var result ShippingDocumentResponse
	if err := c.doRequest("GET", apiPath, params, &result); err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("TikTok API error: code=%d, message=%s", result.Code, result.Message)
	}

	return result.Data.DocURL, nil
}

// ArrangeShipment is a wrapper for ShipPackage to match the service interface
func (c *Client) ArrangeShipment(packageID string, req *ShipPackageRequest) (*ShipPackageResponse, error) {
	return c.ShipPackage(packageID, req)
}
