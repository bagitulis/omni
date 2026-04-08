package platform

import (
	"bytes"
	"context"
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

	"github.com/omni/backend/internal/config"
	"github.com/rs/zerolog/log"
)

// TiktokConfigManager handles TikTok-specific configuration
type TiktokConfigManager struct {
	*BaseConfigManager
	AppKey     string
	AppSecret  string
	ShopID     string
	ShopCipher string
}

// NewTiktokConfigManager creates a TikTok config manager
func NewTiktokConfigManager(tenantID string) *TiktokConfigManager {
	return &TiktokConfigManager{
		BaseConfigManager: NewBaseConfigManager(PlatformTiktok, tenantID),
	}
}

// LoadConfig loads TikTok config from database
func (m *TiktokConfigManager) LoadConfig(ctx context.Context) error {
	// 1. Load global credentials from system.db
	globalConfig := config.GetGlobalConfigService()
	if globalConfig != nil {
		creds, err := globalConfig.GetTiktokCredentials()
		if err == nil {
			m.AppKey = creds.AppKey
			m.AppSecret = creds.AppSecret
		}
	}

	// 2. Load tenant-specific config from tenant db
	if err := m.LoadConfigFromDB(ctx); err != nil {
		m.logger.WithTenantID(m.tenantID).Warn("Failed to load config from DB: " + err.Error())
	}

	// Parse ShopID
	if shopID, ok := m.GetConfig("shopId"); ok && shopID != "" {
		m.ShopID = shopID
	}

	// Parse ShopCipher - required for TikTok Shop API
	if shopCipher, ok := m.GetConfig("shopCipher"); ok && shopCipher != "" {
		m.ShopCipher = shopCipher
	}

	// Log status
	if m.AppKey != "" && m.AppSecret != "" {
		if m.ShopCipher != "" {
			m.logger.WithTenantID(m.tenantID).Info("TikTok config loaded successfully")
		} else {
			m.logger.WithTenantID(m.tenantID).Warn("TikTok config loaded but shopCipher is missing")
		}
	} else {
		m.logger.WithTenantID(m.tenantID).Warn("TikTok credentials missing")
	}

	return nil
}

// IsConfigured returns whether TikTok is properly configured
func (m *TiktokConfigManager) IsConfigured() bool {
	return m.AppKey != "" && m.AppSecret != "" && m.GetAccessToken() != "" && m.ShopCipher != ""
}

// TiktokAPIClient implements APIClient for TikTok
type TiktokAPIClient struct {
	config      *TiktokConfigManager
	initialized bool
	httpClient  *http.Client
}

// NewTiktokAPIClient creates a TikTok API client
func NewTiktokAPIClient(config *TiktokConfigManager) *TiktokAPIClient {
	c := &TiktokAPIClient{
		config:     config,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	if config.IsConfigured() {
		c.initialized = true
	}

	return c
}

// IsInitialized returns whether client is ready
func (c *TiktokAPIClient) IsInitialized() bool {
	return c.initialized && c.config.IsConfigured()
}

const tiktokBaseURL = "https://open-api.tiktokglobalshop.com"

// generateSignature generates HMAC-SHA256 signature for TikTok API
func (c *TiktokAPIClient) generateSignature(path string, params map[string]string, body []byte) string {
	// Sort params by key
	keys := make([]string, 0, len(params))
	for k := range params {
		// Exclude sign, access_token from signature calculation
		if k != "sign" && k != "access_token" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build string to sign: secret + path + params + body + secret
	var signStr strings.Builder
	signStr.WriteString(c.config.AppSecret)
	signStr.WriteString(path)
	for _, k := range keys {
		signStr.WriteString(k)
		signStr.WriteString(params[k])
	}
	if len(body) > 0 {
		signStr.Write(body)
	}
	signStr.WriteString(c.config.AppSecret)

	// HMAC-SHA256
	h := hmac.New(sha256.New, []byte(c.config.AppSecret))
	h.Write([]byte(signStr.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// request makes an HTTP request to TikTok API
func (c *TiktokAPIClient) request(method, path string, queryParams map[string]string, body interface{}) (map[string]interface{}, error) {
	// Build params with required fields
	params := make(map[string]string)
	for k, v := range queryParams {
		params[k] = v
	}
	params["app_key"] = c.config.AppKey
	params["timestamp"] = fmt.Sprintf("%d", time.Now().Unix())
	if c.config.ShopCipher != "" {
		params["shop_cipher"] = c.config.ShopCipher
	}

	// Serialize body if provided
	var bodyBytes []byte
	var err error
	if body != nil {
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
	}

	// Generate signature
	signature := c.generateSignature(path, params, bodyBytes)
	params["sign"] = signature
	params["access_token"] = c.config.GetAccessToken()

	// Build URL
	u, err := url.Parse(tiktokBaseURL + path)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL %s%s: %w", tiktokBaseURL, path, err)
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Create request - send body for ANY method (POST, PUT, PATCH) if body exists
	var req *http.Request
	if len(bodyBytes) > 0 {
		req, err = http.NewRequest(method, u.String(), bytes.NewReader(bodyBytes))
	} else {
		req, err = http.NewRequest(method, u.String(), nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tts-access-token", c.config.GetAccessToken())

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error code
	if code, ok := result["code"].(float64); ok && code != 0 {
		msg := ""
		if m, ok := result["message"].(string); ok {
			msg = m
		}
		log.Error().
			Float64("code", code).
			Str("message", msg).
			Str("path", path).
			Interface("raw_result", result).
			Msg("[TiktokAPI] TikTok API Raw Error")
		return nil, fmt.Errorf("tiktok API error: code=%.0f, message=%s", code, msg)
	}

	return result, nil
}

// Request is the public wrapper for request, allowing use as wholesale.TiktokProductAPI
func (c *TiktokAPIClient) Request(method, path string, queryParams map[string]string, body interface{}) (map[string]interface{}, error) {
	return c.request(method, path, queryParams, body)
}
