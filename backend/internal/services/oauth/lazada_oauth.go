package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"time"
)

// LazadaOAuthConfig holds Lazada OAuth configuration
type LazadaOAuthConfig struct {
	AppKey      string
	AppSecret   string
	RedirectURL string
	IsSandbox   bool
}

// LazadaOAuthService handles Lazada OAuth operations
type LazadaOAuthService struct {
	config *LazadaOAuthConfig
}

// NewLazadaOAuthService creates a new Lazada OAuth service
func NewLazadaOAuthService(appKey, appSecret, redirectURL string, isSandbox bool) *LazadaOAuthService {
	return &LazadaOAuthService{
		config: &LazadaOAuthConfig{
			AppKey:      appKey,
			AppSecret:   appSecret,
			RedirectURL: redirectURL,
			IsSandbox:   isSandbox,
		},
	}
}

// GetBaseURL returns the API base URL
func (s *LazadaOAuthService) GetBaseURL() string {
	if s.config.IsSandbox {
		return "https://api.lazada.test/rest"
	}
	return "https://api.lazada.co.id/rest"
}

// GetAuthBaseURL returns the OAuth base URL for token operations
// Token create/refresh MUST use auth.lazada.com, not country-specific API
func (s *LazadaOAuthService) GetAuthBaseURL() string {
	if s.config.IsSandbox {
		return "https://api.lazada.test/rest"
	}
	return "https://auth.lazada.com/rest"
}

// GetAuthURL generates the OAuth authorization URL
func (s *LazadaOAuthService) GetAuthURL(state string) string {
	baseURL := "https://auth.lazada.com/oauth/authorize"
	if s.config.IsSandbox {
		baseURL = "https://auth.lazada.com/oauth/authorize"
	}

	params := url.Values{}
	params.Add("response_type", "code")
	params.Add("force_auth", "true")
	params.Add("redirect_uri", s.config.RedirectURL)
	params.Add("client_id", s.config.AppKey)
	params.Add("state", state)

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// GenerateSignature generates Lazada API signature
func (s *LazadaOAuthService) GenerateSignature(apiPath string, params map[string]string) string {
	// Sort parameters by key
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Build concatenated string
	var concat strings.Builder
	concat.WriteString(apiPath)
	for _, k := range keys {
		concat.WriteString(k)
		concat.WriteString(params[k])
	}

	// Generate HMAC-SHA256 signature
	h := hmac.New(sha256.New, []byte(s.config.AppSecret))
	h.Write([]byte(concat.String()))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// BuildCommonParams builds common API parameters
func (s *LazadaOAuthService) BuildCommonParams() map[string]string {
	return map[string]string{
		"app_key":      s.config.AppKey,
		"sign_method":  "sha256",
		"timestamp":    fmt.Sprintf("%d000", time.Now().Unix()), // Lazada uses milliseconds
		"partner_id":   "lazop-sdk-go",
		"debug":        "false",
	}
}

// BuildTokenRequest builds the token exchange request
// Uses auth.lazada.com for OAuth token operations
func (s *LazadaOAuthService) BuildTokenRequest(code string) (string, map[string]string) {
	apiPath := "/auth/token/create"
	params := s.BuildCommonParams()
	params["code"] = code

	sign := s.GenerateSignature(apiPath, params)
	params["sign"] = sign

	return fmt.Sprintf("%s%s", s.GetAuthBaseURL(), apiPath), params
}

// BuildRefreshTokenRequest builds the refresh token request
// Uses auth.lazada.com for OAuth token operations (NOT country-specific API)
func (s *LazadaOAuthService) BuildRefreshTokenRequest(refreshToken string) (string, map[string]string) {
	apiPath := "/auth/token/refresh"
	params := s.BuildCommonParams()
	params["refresh_token"] = refreshToken

	sign := s.GenerateSignature(apiPath, params)
	params["sign"] = sign

	return fmt.Sprintf("%s%s", s.GetAuthBaseURL(), apiPath), params
}

// BuildAPIRequest builds an authenticated API request
func (s *LazadaOAuthService) BuildAPIRequest(apiPath, accessToken string, apiParams map[string]string) (string, map[string]string) {
	params := s.BuildCommonParams()
	params["access_token"] = accessToken

	// Merge API-specific parameters
	for k, v := range apiParams {
		params[k] = v
	}

	sign := s.GenerateSignature(apiPath, params)
	params["sign"] = sign

	return fmt.Sprintf("%s%s", s.GetBaseURL(), apiPath), params
}

// GetAppKey returns app key
func (s *LazadaOAuthService) GetAppKey() string {
	return s.config.AppKey
}

// ValidateCallback validates OAuth callback parameters
func (s *LazadaOAuthService) ValidateCallback(code string) error {
	if code == "" {
		return fmt.Errorf("authorization code is required")
	}
	return nil
}

// VerifyWebhookSignature verifies Lazada webhook signature
func (s *LazadaOAuthService) VerifyWebhookSignature(body, signature string) bool {
	h := hmac.New(sha256.New, []byte(s.config.AppSecret))
	h.Write([]byte(body))
	expectedSign := strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
	return hmac.Equal([]byte(signature), []byte(expectedSign))
}
