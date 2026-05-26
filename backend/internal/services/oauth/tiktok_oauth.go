package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TiktokOAuthConfig holds TikTok OAuth configuration
type TiktokOAuthConfig struct {
	AppKey      string
	AppSecret   string
	RedirectURL string
	IsSandbox   bool
}

// TiktokOAuthService handles TikTok OAuth operations
type TiktokOAuthService struct {
	config *TiktokOAuthConfig
}

// NewTiktokOAuthService creates a new TikTok OAuth service
func NewTiktokOAuthService(appKey, appSecret, redirectURL string, isSandbox bool) *TiktokOAuthService {
	return &TiktokOAuthService{
		config: &TiktokOAuthConfig{
			AppKey:      appKey,
			AppSecret:   appSecret,
			RedirectURL: redirectURL,
			IsSandbox:   isSandbox,
		},
	}
}

// GetBaseURL returns the API base URL
func (s *TiktokOAuthService) GetBaseURL() string {
	if s.config.IsSandbox {
		return "https://open-api-sandbox.tiktokglobalshop.com"
	}
	return "https://open-api.tiktokglobalshop.com"
}

// GetAuthURL generates the OAuth authorization URL
func (s *TiktokOAuthService) GetAuthURL(state string) string {
	baseURL := "https://services.tiktokshop.com/open/authorize"

	params := url.Values{}
	params.Add("app_key", s.config.AppKey)
	params.Add("redirect_uri", s.config.RedirectURL)
	params.Add("state", state)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// GenerateSignature generates TikTok API signature
func (s *TiktokOAuthService) GenerateSignature(path string, params map[string]string, timestamp int64) string {
	// Exclude sign, access_token from signature
	signParams := make(map[string]string)
	for k, v := range params {
		if k != "sign" && k != "access_token" {
			signParams[k] = v
		}
	}

	// Sort by key
	keys := make([]string, 0, len(signParams))
	for k := range signParams {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build base string: app_secret + path + sorted_params + app_secret
	var builder strings.Builder
	builder.WriteString(s.config.AppSecret)
	builder.WriteString(path)
	for _, k := range keys {
		builder.WriteString(k)
		builder.WriteString(signParams[k])
	}
	builder.WriteString(s.config.AppSecret)

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(s.config.AppSecret))
	h.Write([]byte(builder.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// BuildCommonParams builds common API parameters
func (s *TiktokOAuthService) BuildCommonParams() map[string]string {
	timestamp := time.Now().Unix()
	return map[string]string{
		"app_key":   s.config.AppKey,
		"timestamp": strconv.FormatInt(timestamp, 10),
	}
}

// BuildTokenRequest builds the token exchange request
func (s *TiktokOAuthService) BuildTokenRequest(authCode string) (string, map[string]string) {
	apiPath := "/api/v2/token/get"
	params := s.BuildCommonParams()
	params["auth_code"] = authCode
	params["grant_type"] = "authorized_code"

	timestamp, _ := strconv.ParseInt(params["timestamp"], 10, 64)
	sign := s.GenerateSignature(apiPath, params, timestamp)
	params["sign"] = sign

	return fmt.Sprintf("%s%s", s.GetBaseURL(), apiPath), params
}

// BuildRefreshTokenRequest builds the refresh token request
// TikTok uses a DIFFERENT auth endpoint for token refresh
// Endpoint: https://auth.tiktok-shops.com/api/v2/token/refresh
// No signature needed - uses app_secret directly
func (s *TiktokOAuthService) BuildRefreshTokenRequest(refreshToken string) (string, map[string]string) {
	// TikTok token refresh uses auth.tiktok-shops.com, NOT open-api.tiktokglobalshop.com
	authURL := "https://auth.tiktok-shops.com/api/v2/token/refresh"

	// Parameters for refresh - no signature needed
	params := map[string]string{
		"app_key":       s.config.AppKey,
		"app_secret":    s.config.AppSecret, // Include app_secret directly
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}

	return authURL, params
}

// BuildAPIRequest builds an authenticated API request
func (s *TiktokOAuthService) BuildAPIRequest(apiPath, accessToken, shopCipher string, apiParams map[string]string) (string, map[string]string) {
	params := s.BuildCommonParams()
	params["access_token"] = accessToken
	if shopCipher != "" {
		params["shop_cipher"] = shopCipher
	}

	// Merge API-specific parameters
	for k, v := range apiParams {
		params[k] = v
	}

	timestamp, _ := strconv.ParseInt(params["timestamp"], 10, 64)
	sign := s.GenerateSignature(apiPath, params, timestamp)
	params["sign"] = sign

	return fmt.Sprintf("%s%s", s.GetBaseURL(), apiPath), params
}

// GetAppKey returns app key
func (s *TiktokOAuthService) GetAppKey() string {
	return s.config.AppKey
}

// ValidateCallback validates OAuth callback parameters
func (s *TiktokOAuthService) ValidateCallback(code string) error {
	if code == "" {
		return fmt.Errorf("authorization code is required")
	}
	return nil
}

// VerifyWebhookSignature verifies TikTok webhook signature
func (s *TiktokOAuthService) VerifyWebhookSignature(body, timestamp, signature string) bool {
	// TikTok signature: HMAC-SHA256(app_secret, timestamp + body)
	h := hmac.New(sha256.New, []byte(s.config.AppSecret))
	h.Write([]byte(timestamp + body))
	expectedSign := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSign))
}

// GetAuthorizedShops returns authorized shop list URL
func (s *TiktokOAuthService) GetAuthorizedShops(accessToken string) (string, map[string]string) {
	return s.BuildAPIRequest("/api/v2/seller/global/active_shops", accessToken, "", nil)
}
