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

// generateSign creates signature for TikTok API
// Mirrors Node.js implementation: tiktok_sdk/utils/generate-sign.ts
// Signature format: HMAC-SHA256(appSecret, appSecret + path + sortedParams + bodyJSON + appSecret)
func (c *Client) generateSign(path string, params map[string]string, body interface{}) string {
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

	// Build sign string: pathname + paramString + bodyJSON (if present)
	var signString strings.Builder
	signString.WriteString(path)
	signString.WriteString(paramString.String())

	// Append body JSON if not nil/empty
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err == nil && len(bodyBytes) > 2 { // > 2 means not empty "{}"
			signString.Write(bodyBytes)
		}
	}

	// Wrap with appSecret: appSecret + signString + appSecret
	var sb strings.Builder
	sb.WriteString(c.appSecret)
	sb.WriteString(signString.String())
	sb.WriteString(c.appSecret)

	// HMAC-SHA256
	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(sb.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// doRequest executes HTTP request (GET without body)
func (c *Client) doRequest(method, apiPath string, params map[string]string, result interface{}) error {
	timestamp := time.Now().Unix()

	// Add common params
	params["app_key"] = c.appKey
	params["timestamp"] = fmt.Sprintf("%d", timestamp)
	if c.accessToken != "" {
		params["access_token"] = c.accessToken
	}
	if c.shopCipher != "" {
		params["shop_cipher"] = c.shopCipher
	}

	// Debug: log params before signing
	fmt.Printf("[TikTok API Debug] %s %s - shop_cipher=%s, token_prefix=%s\n", 
		method, apiPath, c.shopCipher, func() string {
			if len(c.accessToken) > 20 { return c.accessToken[:20] }
			return c.accessToken
		}())

	// Generate signature (no body for GET requests)
	params["sign"] = c.generateSign(apiPath, params, nil)

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
	// TikTok requires access_token in header as x-tts-access-token
	if c.accessToken != "" {
		req.Header.Set("x-tts-access-token", c.accessToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(body, result)
}

// doRequestWithBody executes HTTP request with JSON body
func (c *Client) doRequestWithBody(method, apiPath string, params map[string]string, body interface{}, result interface{}) error {
	timestamp := time.Now().Unix()

	// Add common params
	params["app_key"] = c.appKey
	params["timestamp"] = fmt.Sprintf("%d", timestamp)
	if c.accessToken != "" {
		params["access_token"] = c.accessToken
	}
	if c.shopCipher != "" {
		params["shop_cipher"] = c.shopCipher
	}

	// Generate signature WITH body included (critical for TikTok POST requests)
	params["sign"] = c.generateSign(apiPath, params, body)

	// Build URL
	u, _ := url.Parse(BaseURL + apiPath)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Marshal body
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, u.String(), strings.NewReader(string(bodyBytes)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	// TikTok requires access_token in header as x-tts-access-token
	if c.accessToken != "" {
		req.Header.Set("x-tts-access-token", c.accessToken)
	}

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
