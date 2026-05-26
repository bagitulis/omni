package oauth

import (
	"fmt"
	"net/url"
	"time"
)

// OAuthConfig holds OAuth configuration
type OAuthConfig struct {
	PlatformID   string
	ClientID     string
	ClientSecret string
	RedirectURI  string
}

// OAuthService handles OAuth operations
type OAuthService struct {
	config OAuthConfig
}

// NewOAuthService creates a new OAuth service
func NewOAuthService(config OAuthConfig) *OAuthService {
	return &OAuthService{config: config}
}

// ShopeeOAuthURL generates Shopee authorization URL
func ShopeeOAuthURL(partnerID int64, redirectURI string) string {
	baseURL := "https://partner.shopeemobile.com/api/v2/shop/auth_partner"
	return fmt.Sprintf("%s?partner_id=%d&redirect=%s",
		baseURL, partnerID, url.QueryEscape(redirectURI))
}

// LazadaOAuthURL generates Lazada authorization URL
func LazadaOAuthURL(appKey, redirectURI, region string) string {
	baseURL := "https://auth.lazada.com/oauth/authorize"
	if region == "my" {
		baseURL = "https://auth.lazada.com.my/oauth/authorize"
	}
	return fmt.Sprintf("%s?response_type=code&force_auth=true&redirect_uri=%s&client_id=%s",
		baseURL, url.QueryEscape(redirectURI), appKey)
}

// TiktokOAuthURL generates TikTok authorization URL
func TiktokOAuthURL(appKey, redirectURI string) string {
	baseURL := "https://auth.tiktok-shops.com/oauth/authorize"
	return fmt.Sprintf("%s?app_key=%s&redirect_uri=%s&state=omni",
		baseURL, appKey, url.QueryEscape(redirectURI))
}

func TiktokCallbackURL(baseURL string) string {
	if baseURL == "" {
		return ""
	}
	return fmt.Sprintf("%s/api/platform-auth/callback/tiktok", baseURL)
}

func OAuthStateExpiry() time.Time {
	return time.Now().Add(StateTTL)
}
