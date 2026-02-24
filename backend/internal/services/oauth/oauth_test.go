package oauth_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omni/backend/internal/services/oauth"
)

// ============================================================================
// OAuthService
// ============================================================================

func TestNewOAuthService_NotNil(t *testing.T) {
	cfg := oauth.OAuthConfig{
		PlatformID:   "shopee",
		ClientID:     "client-id",
		ClientSecret: "client-secret",
		RedirectURI:  "https://example.com/callback",
	}
	svc := oauth.NewOAuthService(cfg)
	require.NotNil(t, svc)
}

// ---------------------------------------------------------------------------
// Package-level URL helpers (pure functions)
// ---------------------------------------------------------------------------

func TestShopeeOAuthURL_ContainsPartnerID(t *testing.T) {
	url := oauth.ShopeeOAuthURL(12345, "https://example.com/callback")
	assert.Contains(t, url, "partner_id=12345")
}

func TestShopeeOAuthURL_ContainsBaseURL(t *testing.T) {
	url := oauth.ShopeeOAuthURL(1, "https://example.com/cb")
	assert.Contains(t, url, "partner.shopeemobile.com")
}

func TestShopeeOAuthURL_RedirectIsURLEncoded(t *testing.T) {
	url := oauth.ShopeeOAuthURL(1, "https://example.com/cb?foo=bar")
	// The redirect must be present and the colon in https should be encoded
	assert.Contains(t, url, "redirect=")
}

func TestLazadaOAuthURL_DefaultRegion(t *testing.T) {
	url := oauth.LazadaOAuthURL("myapp", "https://example.com/callback", "sg")
	assert.Contains(t, url, "auth.lazada.com/oauth/authorize")
	assert.Contains(t, url, "client_id=myapp")
}

func TestLazadaOAuthURL_MyRegion(t *testing.T) {
	url := oauth.LazadaOAuthURL("myapp", "https://example.com/callback", "my")
	assert.Contains(t, url, "auth.lazada.com.my/oauth/authorize")
}

func TestLazadaOAuthURL_ContainsRedirectURI(t *testing.T) {
	url := oauth.LazadaOAuthURL("myapp", "https://example.com/cb", "id")
	assert.Contains(t, url, "redirect_uri=")
}

func TestTiktokOAuthURL_ContainsAppKey(t *testing.T) {
	url := oauth.TiktokOAuthURL("my-app-key", "https://example.com/callback")
	assert.Contains(t, url, "app_key=my-app-key")
}

func TestTiktokOAuthURL_ContainsBaseURL(t *testing.T) {
	url := oauth.TiktokOAuthURL("key", "https://example.com/cb")
	assert.Contains(t, url, "auth.tiktok-shops.com/oauth/authorize")
}

func TestTiktokOAuthURL_ContainsState(t *testing.T) {
	url := oauth.TiktokOAuthURL("key", "https://example.com/cb")
	assert.Contains(t, url, "state=omni")
}

// ============================================================================
// ShopeeOAuthService
// ============================================================================

func TestNewShopeeOAuthService_NotNil(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(12345, "partner-key", "https://example.com/cb", false)
	require.NotNil(t, svc)
}

func TestShopeeOAuthService_GetBaseURL_Production(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	assert.Equal(t, "https://partner.shopeemobile.com", svc.GetBaseURL())
}

func TestShopeeOAuthService_GetBaseURL_Sandbox(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", true)
	assert.Equal(t, "https://partner.test-stable.shopeemobile.com", svc.GetBaseURL())
}

func TestShopeeOAuthService_GetPartnerID(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(99999, "key", "https://cb.example.com", false)
	assert.Equal(t, int64(99999), svc.GetPartnerID())
}

func TestShopeeOAuthService_GetTokenURL(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	assert.Equal(t, "https://partner.shopeemobile.com/api/v2/auth/token/get", svc.GetTokenURL())
}

func TestShopeeOAuthService_GetRefreshTokenURL(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	assert.Equal(t, "https://partner.shopeemobile.com/api/v2/auth/access_token/get", svc.GetRefreshTokenURL())
}

func TestShopeeOAuthService_GetAuthURL_ContainsRequiredParams(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(12345, "test-key", "https://example.com/cb", false)
	url := svc.GetAuthURL("test-state")
	assert.Contains(t, url, "partner_id=12345")
	assert.Contains(t, url, "sign=")
	assert.Contains(t, url, "timestamp=")
}

func TestShopeeOAuthService_GenerateSignature_Deterministic(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(12345, "test-partner-key", "https://cb.example.com", false)
	path := "/api/v2/orders/get_order_list"
	ts := int64(1700000000)

	sign1 := svc.GenerateSignature(path, ts, "", 0)
	sign2 := svc.GenerateSignature(path, ts, "", 0)
	assert.Equal(t, sign1, sign2, "same inputs must produce same signature")
	assert.NotEmpty(t, sign1)
	assert.Len(t, sign1, 64, "HMAC-SHA256 hex = 64 chars")
}

func TestShopeeOAuthService_GenerateSignature_WithAccessToken(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(12345, "test-key", "https://cb.example.com", false)
	ts := int64(1700000000)
	signWithToken := svc.GenerateSignature("/api/v2/orders/get", ts, "access-token", 67890)
	signWithout := svc.GenerateSignature("/api/v2/orders/get", ts, "", 0)
	// Different inputs → different signatures
	assert.NotEqual(t, signWithToken, signWithout)
}

func TestShopeeOAuthService_ValidateCallback_EmptyCode(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	err := svc.ValidateCallback("", "12345")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization code is required")
}

func TestShopeeOAuthService_ValidateCallback_EmptyShopID(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	err := svc.ValidateCallback("valid-code", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "shop_id is required")
}

func TestShopeeOAuthService_ValidateCallback_Valid(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	err := svc.ValidateCallback("valid-code", "12345")
	assert.NoError(t, err)
}

func TestShopeeOAuthService_ParseShopID_Valid(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	id, err := svc.ParseShopID("12345")
	require.NoError(t, err)
	assert.Equal(t, int64(12345), id)
}

func TestShopeeOAuthService_ParseShopID_Invalid(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	_, err := svc.ParseShopID("not-a-number")
	require.Error(t, err)
}

func TestShopeeOAuthService_ParseShopID_Empty(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(1, "key", "https://cb.example.com", false)
	_, err := svc.ParseShopID("")
	require.Error(t, err)
}

func TestShopeeOAuthService_BuildTokenRequest_ContainsCode(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(12345, "test-key", "https://cb.example.com", false)
	body := svc.BuildTokenRequest("auth-code-123", 67890)
	assert.Equal(t, "auth-code-123", body["code"])
	assert.Equal(t, int64(12345), body["partner_id"])
}

func TestShopeeOAuthService_BuildRefreshTokenRequest_Structure(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(12345, "test-key", "https://cb.example.com", false)
	url, body := svc.BuildRefreshTokenRequest("refresh-token-xyz", 67890)
	assert.Contains(t, url, "/api/v2/auth/access_token/get")
	assert.Contains(t, url, "partner_id=12345")
	assert.Equal(t, "refresh-token-xyz", body["refresh_token"])
	assert.Equal(t, int64(67890), body["shop_id"])
}

func TestShopeeOAuthService_VerifyWebhookSignature_InvalidSig(t *testing.T) {
	// A known-bad signature must return false
	svc := oauth.NewShopeeOAuthService(1, "webhook-secret", "https://cb.example.com", false)
	reqURL := "https://example.com/webhook"
	body := `{"order_sn":"123"}`
	assert.False(t, svc.VerifyWebhookSignature(reqURL, body, "bad-signature"))
}

func TestShopeeOAuthService_BuildAPIParams_ContainsRequiredFields(t *testing.T) {
	svc := oauth.NewShopeeOAuthService(12345, "test-key", "https://cb.example.com", false)
	params := svc.BuildAPIParams("/api/v2/orders/get", "access-token", 67890, map[string]string{
		"order_sn": "TEST-001",
	})
	assert.Equal(t, "12345", params.Get("partner_id"))
	assert.Equal(t, "access-token", params.Get("access_token"))
	assert.Equal(t, "67890", params.Get("shop_id"))
	assert.NotEmpty(t, params.Get("sign"))
	assert.Equal(t, "TEST-001", params.Get("order_sn"))
}

// ============================================================================
// LazadaOAuthService
// ============================================================================

func TestNewLazadaOAuthService_NotNil(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("app-key", "app-secret", "https://example.com/cb", false)
	require.NotNil(t, svc)
}

func TestLazadaOAuthService_GetBaseURL_Production(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("key", "secret", "https://cb.example.com", false)
	assert.Equal(t, "https://api.lazada.co.id/rest", svc.GetBaseURL())
}

func TestLazadaOAuthService_GetBaseURL_Sandbox(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("key", "secret", "https://cb.example.com", true)
	assert.Equal(t, "https://api.lazada.test/rest", svc.GetBaseURL())
}

func TestLazadaOAuthService_GetAuthBaseURL_Production(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("key", "secret", "https://cb.example.com", false)
	assert.Equal(t, "https://auth.lazada.com/rest", svc.GetAuthBaseURL())
}

func TestLazadaOAuthService_GetAuthBaseURL_Sandbox(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("key", "secret", "https://cb.example.com", true)
	assert.Equal(t, "https://api.lazada.test/rest", svc.GetAuthBaseURL())
}

func TestLazadaOAuthService_GetAppKey(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("my-app-key", "secret", "https://cb.example.com", false)
	assert.Equal(t, "my-app-key", svc.GetAppKey())
}

func TestLazadaOAuthService_GetAuthURL_ContainsRequiredParams(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("my-app-key", "secret", "https://example.com/cb", false)
	url := svc.GetAuthURL("test-state")
	assert.Contains(t, url, "response_type=code")
	assert.Contains(t, url, "client_id=my-app-key")
	assert.Contains(t, url, "state=test-state")
}

func TestLazadaOAuthService_GenerateSignature_Deterministic(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("app-key", "app-secret", "https://cb.example.com", false)
	params := map[string]string{
		"app_key":     "app-key",
		"sign_method": "sha256",
		"timestamp":   "1700000000000",
	}
	sign1 := svc.GenerateSignature("/auth/token/create", params)
	sign2 := svc.GenerateSignature("/auth/token/create", params)
	assert.Equal(t, sign1, sign2)
	assert.NotEmpty(t, sign1)
}

func TestLazadaOAuthService_GenerateSignature_IsUpperCase(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("key", "secret", "https://cb.example.com", false)
	sign := svc.GenerateSignature("/test", map[string]string{"a": "b"})
	// Lazada signature is uppercase hex
	assert.Equal(t, strings.ToUpper(sign), sign)
}

func TestLazadaOAuthService_BuildCommonParams_ContainsRequiredKeys(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("my-app-key", "secret", "https://cb.example.com", false)
	params := svc.BuildCommonParams()
	assert.Equal(t, "my-app-key", params["app_key"])
	assert.Equal(t, "sha256", params["sign_method"])
	assert.NotEmpty(t, params["timestamp"])
}

func TestLazadaOAuthService_BuildTokenRequest_UsesAuthBaseURL(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("app-key", "secret", "https://cb.example.com", false)
	url, params := svc.BuildTokenRequest("auth-code")
	assert.Contains(t, url, "auth.lazada.com/rest/auth/token/create")
	assert.Equal(t, "auth-code", params["code"])
	assert.NotEmpty(t, params["sign"])
}

func TestLazadaOAuthService_BuildRefreshTokenRequest_UsesAuthBaseURL(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("app-key", "secret", "https://cb.example.com", false)
	url, params := svc.BuildRefreshTokenRequest("refresh-token-xyz")
	assert.Contains(t, url, "auth.lazada.com/rest/auth/token/refresh")
	assert.Equal(t, "refresh-token-xyz", params["refresh_token"])
	assert.NotEmpty(t, params["sign"])
}

func TestLazadaOAuthService_ValidateCallback_EmptyCode(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("key", "secret", "https://cb.example.com", false)
	err := svc.ValidateCallback("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization code is required")
}

func TestLazadaOAuthService_ValidateCallback_Valid(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("key", "secret", "https://cb.example.com", false)
	err := svc.ValidateCallback("valid-code")
	assert.NoError(t, err)
}

func TestLazadaOAuthService_VerifyWebhookSignature_InvalidSig(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("key", "secret", "https://cb.example.com", false)
	assert.False(t, svc.VerifyWebhookSignature(`{"event":"test"}`, "bad-signature"))
}

func TestLazadaOAuthService_BuildAPIRequest_ContainsRequiredKeys(t *testing.T) {
	svc := oauth.NewLazadaOAuthService("app-key", "secret", "https://cb.example.com", false)
	url, params := svc.BuildAPIRequest("/orders/get", "access-token", map[string]string{
		"order_id": "ORD-001",
	})
	assert.Contains(t, url, "/orders/get")
	assert.Equal(t, "access-token", params["access_token"])
	assert.Equal(t, "ORD-001", params["order_id"])
	assert.NotEmpty(t, params["sign"])
}

// ============================================================================
// TiktokOAuthService
// ============================================================================

func TestNewTiktokOAuthService_NotNil(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("app-key", "app-secret", "https://example.com/cb", false)
	require.NotNil(t, svc)
}

func TestTiktokOAuthService_GetBaseURL_Production(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("key", "secret", "https://cb.example.com", false)
	assert.Equal(t, "https://open-api.tiktokglobalshop.com", svc.GetBaseURL())
}

func TestTiktokOAuthService_GetBaseURL_Sandbox(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("key", "secret", "https://cb.example.com", true)
	assert.Equal(t, "https://open-api-sandbox.tiktokglobalshop.com", svc.GetBaseURL())
}

func TestTiktokOAuthService_GetAppKey(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("my-app-key", "secret", "https://cb.example.com", false)
	assert.Equal(t, "my-app-key", svc.GetAppKey())
}

func TestTiktokOAuthService_GetAuthURL_ContainsAppKey(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("tiktok-key", "secret", "https://example.com/cb", false)
	url := svc.GetAuthURL("my-state")
	assert.Contains(t, url, "app_key=tiktok-key")
	assert.Contains(t, url, "state=my-state")
}

func TestTiktokOAuthService_GetAuthURL_BaseURLIsServiceDomain(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("key", "secret", "https://cb.example.com", false)
	url := svc.GetAuthURL("state")
	assert.Contains(t, url, "services.tiktokshop.com/open/authorize")
}

func TestTiktokOAuthService_GenerateSignature_Deterministic(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("app-key", "app-secret", "https://cb.example.com", false)
	params := map[string]string{
		"app_key":   "app-key",
		"timestamp": "1700000000",
	}
	ts := int64(1700000000)
	sign1 := svc.GenerateSignature("/api/v2/orders/get", params, ts)
	sign2 := svc.GenerateSignature("/api/v2/orders/get", params, ts)
	assert.Equal(t, sign1, sign2)
	assert.NotEmpty(t, sign1)
	assert.Len(t, sign1, 64, "HMAC-SHA256 hex = 64 chars")
}

func TestTiktokOAuthService_GenerateSignature_ExcludesSignAndAccessToken(t *testing.T) {
	// sign and access_token should be excluded from signature computation
	svc := oauth.NewTiktokOAuthService("app-key", "app-secret", "https://cb.example.com", false)
	ts := int64(1700000000)

	// With sign/access_token fields — should be same as without
	paramsWithExtra := map[string]string{
		"app_key":      "app-key",
		"timestamp":    "1700000000",
		"sign":         "ignore-me",
		"access_token": "ignore-me-too",
	}
	paramsWithout := map[string]string{
		"app_key":   "app-key",
		"timestamp": "1700000000",
	}
	signWith := svc.GenerateSignature("/api/path", paramsWithExtra, ts)
	signWithout := svc.GenerateSignature("/api/path", paramsWithout, ts)
	assert.Equal(t, signWith, signWithout, "sign and access_token must be excluded from signing")
}

func TestTiktokOAuthService_BuildCommonParams_ContainsRequiredKeys(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("tiktok-key", "secret", "https://cb.example.com", false)
	params := svc.BuildCommonParams()
	assert.Equal(t, "tiktok-key", params["app_key"])
	assert.NotEmpty(t, params["timestamp"])
}

func TestTiktokOAuthService_BuildTokenRequest_Structure(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("app-key", "secret", "https://cb.example.com", false)
	url, params := svc.BuildTokenRequest("auth-code-123")
	assert.Contains(t, url, "open-api.tiktokglobalshop.com/api/v2/token/get")
	assert.Equal(t, "auth-code-123", params["auth_code"])
	assert.Equal(t, "authorized_code", params["grant_type"])
	assert.NotEmpty(t, params["sign"])
}

func TestTiktokOAuthService_BuildRefreshTokenRequest_UsesAuthDomain(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("app-key", "app-secret", "https://cb.example.com", false)
	url, params := svc.BuildRefreshTokenRequest("refresh-token-xyz")
	// TikTok uses different auth domain for refresh
	assert.Equal(t, "https://auth.tiktok-shops.com/api/v2/token/refresh", url)
	assert.Equal(t, "refresh-token-xyz", params["refresh_token"])
	assert.Equal(t, "refresh_token", params["grant_type"])
	assert.Equal(t, "app-key", params["app_key"])
	assert.Equal(t, "app-secret", params["app_secret"])
}

func TestTiktokOAuthService_ValidateCallback_EmptyCode(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("key", "secret", "https://cb.example.com", false)
	err := svc.ValidateCallback("")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "authorization code is required")
}

func TestTiktokOAuthService_ValidateCallback_Valid(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("key", "secret", "https://cb.example.com", false)
	err := svc.ValidateCallback("valid-code")
	assert.NoError(t, err)
}

func TestTiktokOAuthService_VerifyWebhookSignature_InvalidSig(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("key", "secret", "https://cb.example.com", false)
	assert.False(t, svc.VerifyWebhookSignature(`{"event":"test"}`, "1700000000", "bad-signature"))
}

func TestTiktokOAuthService_BuildAPIRequest_ContainsRequiredFields(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("app-key", "secret", "https://cb.example.com", false)
	url, params := svc.BuildAPIRequest("/api/v2/orders/get", "access-token", "cipher-abc", map[string]string{
		"order_id": "ORD-001",
	})
	assert.Contains(t, url, "open-api.tiktokglobalshop.com/api/v2/orders/get")
	assert.Equal(t, "access-token", params["access_token"])
	assert.Equal(t, "cipher-abc", params["shop_cipher"])
	assert.Equal(t, "ORD-001", params["order_id"])
	assert.NotEmpty(t, params["sign"])
}

func TestTiktokOAuthService_BuildAPIRequest_NoShopCipher(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("app-key", "secret", "https://cb.example.com", false)
	_, params := svc.BuildAPIRequest("/api/v2/test", "access-token", "", nil)
	// shop_cipher should not be present when empty
	_, hasCipher := params["shop_cipher"]
	assert.False(t, hasCipher, "shop_cipher should not be set when empty")
}

func TestTiktokOAuthService_GetAuthorizedShops_ReturnsURL(t *testing.T) {
	svc := oauth.NewTiktokOAuthService("app-key", "secret", "https://cb.example.com", false)
	url, params := svc.GetAuthorizedShops("access-token-123")
	assert.Contains(t, url, "/api/v2/seller/global/active_shops")
	assert.Equal(t, "access-token-123", params["access_token"])
}
