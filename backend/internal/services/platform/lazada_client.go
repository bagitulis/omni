package platform

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
	"sort"
	"strings"
	"time"

	"github.com/omni/backend/internal/config"
)

// LazadaConfigManager handles Lazada-specific configuration
type LazadaConfigManager struct {
	*BaseConfigManager
	AppKey    string
	AppSecret string
	Region    string
}

// NewLazadaConfigManager creates a Lazada config manager
func NewLazadaConfigManager(tenantID string) *LazadaConfigManager {
	return &LazadaConfigManager{
		BaseConfigManager: NewBaseConfigManager(PlatformLazada, tenantID),
		Region:            "ID", // Default to Indonesia
	}
}

// LoadConfig loads Lazada config from database
func (m *LazadaConfigManager) LoadConfig(ctx context.Context) error {
	// 1. Load global credentials from system.db
	globalConfig := config.GetGlobalConfigService()
	if globalConfig != nil {
		creds, err := globalConfig.GetLazadaCredentials()
		if err == nil {
			m.AppKey = creds.AppKey
			m.AppSecret = creds.AppSecret
		}
	}

	// 2. Load tenant-specific config from tenant db
	if err := m.LoadConfigFromDB(ctx); err != nil {
		m.logger.WithTenantID(m.tenantID).Warn("Failed to load config from DB: " + err.Error())
	}

	// Parse region
	if region, ok := m.GetConfig("region"); ok && region != "" {
		m.Region = region
	}

	// Log status
	if m.AppKey != "" && m.AppSecret != "" {
		m.logger.WithTenantID(m.tenantID).Info("Lazada config loaded successfully")
	} else {
		m.logger.WithTenantID(m.tenantID).Warn("Lazada credentials missing")
	}

	return nil
}

// IsConfigured returns whether Lazada is properly configured
func (m *LazadaConfigManager) IsConfigured() bool {
	return m.AppKey != "" && m.AppSecret != "" && m.GetAccessToken() != ""
}

// LazadaAPIClient implements APIClient for Lazada
type LazadaAPIClient struct {
	config      *LazadaConfigManager
	initialized bool
}

// NewLazadaAPIClient creates a Lazada API client
func NewLazadaAPIClient(config *LazadaConfigManager) *LazadaAPIClient {
	c := &LazadaAPIClient{
		config: config,
	}

	if config.IsConfigured() {
		c.initialized = true
	}

	return c
}

// IsInitialized returns whether client is ready
func (c *LazadaAPIClient) IsInitialized() bool {
	return c.initialized && c.config.IsConfigured()
}

// getBaseURL returns API base URL based on region
func (c *LazadaAPIClient) getBaseURL() string {
	switch c.config.Region {
	case "SG":
		return "https://api.lazada.sg/rest"
	case "MY":
		return "https://api.lazada.com.my/rest"
	case "VN":
		return "https://api.lazada.vn/rest"
	case "TH":
		return "https://api.lazada.co.th/rest"
	case "PH":
		return "https://api.lazada.com.ph/rest"
	case "ID":
		return "https://api.lazada.co.id/rest"
	default:
		return "https://api.lazada.co.id/rest"
	}
}

// generateSignature generates HMAC-SHA256 signature for Lazada API
func (c *LazadaAPIClient) generateSignature(apiPath string, params map[string]string) string {
	// Sort params by key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build string to sign: path + sorted params
	var signStr strings.Builder
	signStr.WriteString(apiPath)
	for _, k := range keys {
		signStr.WriteString(k)
		signStr.WriteString(params[k])
	}

	// HMAC-SHA256
	h := hmac.New(sha256.New, []byte(c.config.AppSecret))
	h.Write([]byte(signStr.String()))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// request makes HTTP request to Lazada API
func (c *LazadaAPIClient) request(apiPath string, apiParams map[string]string) (map[string]interface{}, error) {
	// System params
	params := map[string]string{
		"app_key":      c.config.AppKey,
		"timestamp":    fmt.Sprintf("%d000", time.Now().Unix()),
		"sign_method":  "sha256",
		"access_token": c.config.GetAccessToken(),
	}

	// Add API params
	for k, v := range apiParams {
		params[k] = v
	}

	// Generate signature
	params["sign"] = c.generateSignature(apiPath, params)

	// Build URL
	baseURL := c.getBaseURL()
	u, _ := url.Parse(baseURL + apiPath)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for error
	if code, ok := result["code"].(string); ok && code != "0" {
		msg := ""
		if m, ok := result["message"].(string); ok {
			msg = m
		}
		return nil, fmt.Errorf("lazada API error: code=%s, message=%s", code, msg)
	}

	return result, nil
}

// Note: API methods (GetOrderList, GetOrderDetails, GetProductList)
// are defined in lazada_client_api.go
