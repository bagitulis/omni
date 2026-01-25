package oauth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// ShopeeOAuthConfig holds Shopee OAuth configuration
type ShopeeOAuthConfig struct {
	PartnerID   int64
	PartnerKey  string
	RedirectURL string
	IsSandbox   bool
}

// ShopeeOAuthService handles Shopee OAuth operations
type ShopeeOAuthService struct {
	config *ShopeeOAuthConfig
}

// NewShopeeOAuthService creates a new Shopee OAuth service
func NewShopeeOAuthService(partnerID int64, partnerKey, redirectURL string, isSandbox bool) *ShopeeOAuthService {
	return &ShopeeOAuthService{
		config: &ShopeeOAuthConfig{
			PartnerID:   partnerID,
			PartnerKey:  partnerKey,
			RedirectURL: redirectURL,
			IsSandbox:   isSandbox,
		},
	}
}

// GetBaseURL returns the API base URL
func (s *ShopeeOAuthService) GetBaseURL() string {
	if s.config.IsSandbox {
		return "https://partner.test-stable.shopeemobile.com"
	}
	return "https://partner.shopeemobile.com"
}

// GenerateSignature generates Shopee API signature
func (s *ShopeeOAuthService) GenerateSignature(path string, timestamp int64, accessToken string, shopID int64) string {
	var baseString string
	if accessToken != "" && shopID > 0 {
		baseString = fmt.Sprintf("%d%s%d%s%d", s.config.PartnerID, path, timestamp, accessToken, shopID)
	} else {
		baseString = fmt.Sprintf("%d%s%d", s.config.PartnerID, path, timestamp)
	}
	h := hmac.New(sha256.New, []byte(s.config.PartnerKey))
	h.Write([]byte(baseString))
	return hex.EncodeToString(h.Sum(nil))
}

// GetAuthURL generates the OAuth authorization URL
func (s *ShopeeOAuthService) GetAuthURL(state string) string {
	timestamp := time.Now().Unix()
	path := "/api/v2/shop/auth_partner"
	sign := s.GenerateSignature(path, timestamp, "", 0)

	params := url.Values{}
	params.Add("partner_id", strconv.FormatInt(s.config.PartnerID, 10))
	params.Add("timestamp", strconv.FormatInt(timestamp, 10))
	params.Add("sign", sign)
	params.Add("redirect", s.config.RedirectURL)

	return fmt.Sprintf("%s%s?%s", s.GetBaseURL(), path, params.Encode())
}

// GetTokenURL returns the token exchange URL
func (s *ShopeeOAuthService) GetTokenURL() string {
	return fmt.Sprintf("%s/api/v2/auth/token/get", s.GetBaseURL())
}

// GetRefreshTokenURL returns the refresh token URL
func (s *ShopeeOAuthService) GetRefreshTokenURL() string {
	return fmt.Sprintf("%s/api/v2/auth/access_token/get", s.GetBaseURL())
}

// BuildTokenRequest builds the token exchange request body
func (s *ShopeeOAuthService) BuildTokenRequest(code string, shopID int64) map[string]interface{} {
	timestamp := time.Now().Unix()
	path := "/api/v2/auth/token/get"
	sign := s.GenerateSignature(path, timestamp, "", 0)

	return map[string]interface{}{
		"code":       code,
		"shop_id":    shopID,
		"partner_id": s.config.PartnerID,
		"timestamp":  timestamp,
		"sign":       sign,
	}
}

// BuildRefreshTokenRequest builds the refresh token request URL with query params and body
// Shopee API requires partner_id, timestamp, sign in query params
// and refresh_token, shop_id, partner_id in POST body
func (s *ShopeeOAuthService) BuildRefreshTokenRequest(refreshToken string, shopID int64) (string, map[string]interface{}) {
	timestamp := time.Now().Unix()
	path := "/api/v2/auth/access_token/get"
	sign := s.GenerateSignature(path, timestamp, "", 0)

	// Build URL with query params
	params := url.Values{}
	params.Add("partner_id", strconv.FormatInt(s.config.PartnerID, 10))
	params.Add("timestamp", strconv.FormatInt(timestamp, 10))
	params.Add("sign", sign)
	
	fullURL := fmt.Sprintf("%s%s?%s", s.GetBaseURL(), path, params.Encode())

	// Build POST body
	body := map[string]interface{}{
		"refresh_token": refreshToken,
		"shop_id":       shopID,
		"partner_id":    s.config.PartnerID,
	}

	return fullURL, body
}

// BuildAPIParams builds authenticated API request parameters
func (s *ShopeeOAuthService) BuildAPIParams(path, accessToken string, shopID int64, params map[string]string) url.Values {
	timestamp := time.Now().Unix()
	sign := s.GenerateSignature(path, timestamp, accessToken, shopID)

	result := url.Values{}
	result.Add("partner_id", strconv.FormatInt(s.config.PartnerID, 10))
	result.Add("timestamp", strconv.FormatInt(timestamp, 10))
	result.Add("access_token", accessToken)
	result.Add("shop_id", strconv.FormatInt(shopID, 10))
	result.Add("sign", sign)

	for k, v := range params {
		result.Add(k, v)
	}
	return result
}

// VerifyWebhookSignature verifies Shopee webhook signature
func (s *ShopeeOAuthService) VerifyWebhookSignature(requestURL, body, signature string) bool {
	baseString := requestURL + "|" + body
	h := hmac.New(sha256.New, []byte(s.config.PartnerKey))
	h.Write([]byte(baseString))
	expectedSign := hex.EncodeToString(h.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSign))
}

// GetPartnerID returns partner ID
func (s *ShopeeOAuthService) GetPartnerID() int64 {
	return s.config.PartnerID
}

// ValidateCallback validates OAuth callback parameters
func (s *ShopeeOAuthService) ValidateCallback(code, shopID string) error {
	if code == "" {
		return fmt.Errorf("authorization code is required")
	}
	if shopID == "" {
		return fmt.Errorf("shop_id is required")
	}
	return nil
}

// ParseShopID parses shop ID from string
func (s *ShopeeOAuthService) ParseShopID(shopIDStr string) (int64, error) {
	return strconv.ParseInt(shopIDStr, 10, 64)
}
