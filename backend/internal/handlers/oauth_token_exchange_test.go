package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestLazadaTokenResponse_Structure tests the Lazada token response structure
func TestLazadaTokenResponse_Structure(t *testing.T) {
	jsonData := `{
		"access_token": "test-access-token",
		"refresh_token": "test-refresh-token",
		"expires_in": 3600,
		"refresh_expires_in": 86400,
		"country": "ID"
	}`

	var resp LazadaTokenResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "test-access-token", resp.AccessToken)
	assert.Equal(t, "test-refresh-token", resp.RefreshToken)
	assert.Equal(t, int64(3600), resp.ExpiresIn)
	assert.Equal(t, int64(86400), resp.RefreshExpiresIn)
	assert.Equal(t, "ID", resp.Country)
}

// TestTiktokTokenResponse_Structure tests the TikTok token response structure
func TestTiktokTokenResponse_Structure(t *testing.T) {
	jsonData := `{
		"code": 0,
		"message": "Success",
		"data": {
			"access_token": "test-access-token",
			"refresh_token": "test-refresh-token",
			"access_token_expire_in": 3600,
			"refresh_token_expire_in": 86400,
			"open_id": "test-open-id",
			"seller_name": "Test Shop"
		}
	}`

	var resp TiktokTokenResponse
	err := json.Unmarshal([]byte(jsonData), &resp)
	assert.NoError(t, err)
	assert.Equal(t, 0, resp.Code)
	assert.Equal(t, "Success", resp.Message)
	assert.Equal(t, "test-access-token", resp.Data.AccessToken)
	assert.Equal(t, "test-refresh-token", resp.Data.RefreshToken)
	assert.Equal(t, int64(3600), resp.Data.AccessTokenExpireIn)
	assert.Equal(t, int64(86400), resp.Data.RefreshTokenExpireIn)
	assert.Equal(t, "test-open-id", resp.Data.OpenID)
	assert.Equal(t, "Test Shop", resp.Data.SellerName)
}

// TestExchangeShopeeToken_NoConfigRepo tests that exchangeShopeeToken fails properly
// when credentials repository is not configured
func TestExchangeShopeeToken_NoConfigRepo(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewOAuthHandler(nil, nil, nil, "http://localhost:3000")

	// Create a mock gin context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	// exchangeShopeeToken should fail without configRepo — no false-positive success
	err := handler.exchangeShopeeToken(c, "tenant-id", "code", "12345")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get Shopee credentials")
}

// TestOAuthHandler_DoLazadaTokenRequest_InvalidURL tests invalid URL handling
func TestOAuthHandler_DoLazadaTokenRequest_InvalidURL(t *testing.T) {
	handler := NewOAuthHandler(nil, nil, nil, "http://localhost:3000")

	// Call with invalid URL - should fail
	_, err := handler.doLazadaTokenRequest("not-a-valid-url", map[string]string{})
	assert.Error(t, err)
}

// TestOAuthHandler_DoTiktokTokenRequest_InvalidURL tests invalid URL handling
func TestOAuthHandler_DoTiktokTokenRequest_InvalidURL(t *testing.T) {
	handler := NewOAuthHandler(nil, nil, nil, "http://localhost:3000")

	// Call with invalid URL - should fail
	_, err := handler.doTiktokTokenRequest("not-a-valid-url", map[string]string{})
	assert.Error(t, err)
}
