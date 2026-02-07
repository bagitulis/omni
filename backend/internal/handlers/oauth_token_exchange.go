package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
)

// exchangeShopeeToken exchanges Shopee auth code for tokens
func (h *OAuthHandler) exchangeShopeeToken(_ *gin.Context, _, _, _ string) error {
	// Token exchange implementation will be done with HTTP client
	// This is a placeholder - actual implementation requires HTTP calls
	return nil
}

// exchangeLazadaToken exchanges Lazada auth code for tokens
func (h *OAuthHandler) exchangeLazadaToken(c *gin.Context, tenantID, code string) error {
	ctx := c.Request.Context()

	// Get Lazada credentials
	creds, err := h.configRepo.GetLazadaCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Lazada credentials: %w", err)
	}

	callbackURL := fmt.Sprintf("%s/api/platform-auth/callback/lazada", h.getBackendURL(c))
	lazadaService := oauth.NewLazadaOAuthService(creds.AppKey, creds.AppSecret, callbackURL, false)

	// Build token request
	tokenURL, params := lazadaService.BuildTokenRequest(code)

	// Make HTTP request and parse response
	tokenResp, err := h.doLazadaTokenRequest(tokenURL, params)
	if err != nil {
		return err
	}

	// Save tokens to tenant database
	return h.saveLazadaTokens(ctx, tenantID, tokenResp)
}

// doLazadaTokenRequest performs the HTTP request to Lazada token endpoint
func (h *OAuthHandler) doLazadaTokenRequest(tokenURL string, params map[string]string) (*LazadaTokenResponse, error) {
	// Build form data
	formData := url.Values{}
	for k, v := range params {
		formData.Set(k, v)
	}

	log.Printf("[Lazada OAuth] Exchanging code for token at: %s", tokenURL)

	// Make HTTP request
	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[Lazada OAuth] Token response status: %d, body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp LazadaTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response: %s", string(body))
	}

	return &tokenResp, nil
}

// saveLazadaTokens saves Lazada tokens to the tenant database
func (h *OAuthHandler) saveLazadaTokens(ctx context.Context, tenantID string, tokenResp *LazadaTokenResponse) error {
	// Get tenant-specific database
	tenantDB, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	// Use TenantPlatformConfigRepository for key-value storage
	tenantRepo := repositories.NewTenantPlatformConfigRepository(tenantDB)

	// Update tokens using the correct key-value pattern
	if err := tenantRepo.UpdateTokens(ctx, models.PlatformLazada, tokenResp.AccessToken, tokenResp.RefreshToken, tokenResp.ExpiresIn, tokenResp.RefreshExpiresIn); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	// Save country/region
	if tokenResp.Country != "" {
		if err := tenantRepo.SetConfig(ctx, models.PlatformLazada, "country", tokenResp.Country, false); err != nil {
			log.Printf("[Lazada OAuth] Warning: failed to save country: %v", err)
		}
	}

	// Save last_refresh timestamp
	if err := tenantRepo.SetConfig(ctx, models.PlatformLazada, "last_refresh", time.Now().Format(time.RFC3339), false); err != nil {
		log.Printf("[Lazada OAuth] Warning: failed to save last_refresh: %v", err)
	}

	log.Printf("[Lazada OAuth] Successfully saved tokens for tenant %s", tenantID)
	return nil
}

// exchangeTiktokToken exchanges TikTok auth code for tokens
func (h *OAuthHandler) exchangeTiktokToken(c *gin.Context, tenantID, code string) error {
	ctx := c.Request.Context()

	// Get TikTok credentials
	creds, err := h.configRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to get TikTok credentials: %w", err)
	}

	tiktokService := oauth.NewTiktokOAuthService(creds.AppKey, creds.AppSecret, "", false)

	// Build token request
	tokenURL, params := tiktokService.BuildTokenRequest(code)

	// Make HTTP request and parse response
	tokenResp, err := h.doTiktokTokenRequest(tokenURL, params)
	if err != nil {
		return err
	}

	// Save tokens to tenant database
	return h.saveTiktokTokens(ctx, tenantID, tokenResp)
}

// doTiktokTokenRequest performs the HTTP request to TikTok token endpoint
func (h *OAuthHandler) doTiktokTokenRequest(tokenURL string, params map[string]string) (*TiktokTokenResponse, error) {
	// Build query string
	queryParams := url.Values{}
	for k, v := range params {
		queryParams.Set(k, v)
	}

	fullURL := tokenURL + "?" + queryParams.Encode()
	log.Printf("[TikTok OAuth] Exchanging code for token at: %s", fullURL)

	// Make HTTP GET request (TikTok uses GET for token)
	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[TikTok OAuth] Token response status: %d, body: %s", resp.StatusCode, string(body))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp TiktokTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Code != 0 {
		return nil, fmt.Errorf("TikTok API error: code=%d, message=%s", tokenResp.Code, tokenResp.Message)
	}

	if tokenResp.Data.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response: %s", string(body))
	}

	return &tokenResp, nil
}

// saveTiktokTokens saves TikTok tokens to the tenant database
func (h *OAuthHandler) saveTiktokTokens(ctx context.Context, tenantID string, tokenResp *TiktokTokenResponse) error {
	// Get tenant-specific database
	tenantDB, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	// Use TenantPlatformConfigRepository for key-value storage
	tenantRepo := repositories.NewTenantPlatformConfigRepository(tenantDB)

	// TikTok API returns expiry as Unix timestamps OR relative seconds
	// We need to normalize to relative seconds for UpdateTokens()
	nowSeconds := time.Now().Unix()

	// Handle access token expiry
	accessTokenExpireValue := tokenResp.Data.AccessTokenExpireIn
	var expiresInSeconds int64
	if accessTokenExpireValue > nowSeconds {
		// It's a Unix timestamp - convert to relative seconds
		expiresInSeconds = accessTokenExpireValue - nowSeconds
		log.Printf("[TikTok OAuth] access_token_expire_in is Unix timestamp: %d (relative: %d seconds)", accessTokenExpireValue, expiresInSeconds)
	} else if accessTokenExpireValue > 0 {
		// It's relative seconds - use as-is
		expiresInSeconds = accessTokenExpireValue
		log.Printf("[TikTok OAuth] access_token_expire_in is relative: %d seconds", expiresInSeconds)
	} else {
		// Default to 7 days
		expiresInSeconds = 7 * 24 * 60 * 60
		log.Printf("[TikTok OAuth] access_token_expire_in not provided, using default: %d seconds", expiresInSeconds)
	}

	// Handle refresh token expiry
	refreshTokenExpireValue := tokenResp.Data.RefreshTokenExpireIn
	var refreshExpiresInSeconds int64
	if refreshTokenExpireValue > nowSeconds {
		// It's a Unix timestamp - convert to relative seconds
		refreshExpiresInSeconds = refreshTokenExpireValue - nowSeconds
		log.Printf("[TikTok OAuth] refresh_token_expire_in is Unix timestamp: %d (relative: %d seconds = %d days)", refreshTokenExpireValue, refreshExpiresInSeconds, refreshExpiresInSeconds/86400)
	} else if refreshTokenExpireValue > 0 {
		// It's relative seconds - use as-is
		refreshExpiresInSeconds = refreshTokenExpireValue
		log.Printf("[TikTok OAuth] refresh_token_expire_in is relative: %d seconds = %d days", refreshExpiresInSeconds, refreshExpiresInSeconds/86400)
	} else {
		// Default to 90 days
		refreshExpiresInSeconds = 90 * 24 * 60 * 60
		log.Printf("[TikTok OAuth] refresh_token_expire_in not provided, using default: %d seconds (90 days)", refreshExpiresInSeconds)
	}

	// Update tokens with normalized expiry values
	if err := tenantRepo.UpdateTokens(ctx, models.PlatformTiktok, tokenResp.Data.AccessToken, tokenResp.Data.RefreshToken, expiresInSeconds, refreshExpiresInSeconds); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	// Save openId and sellerName
	if tokenResp.Data.OpenID != "" {
		if err := tenantRepo.SetConfig(ctx, models.PlatformTiktok, "openId", tokenResp.Data.OpenID, false); err != nil {
			log.Printf("[TikTok OAuth] Warning: failed to save openId: %v", err)
		}
	}
	if tokenResp.Data.SellerName != "" {
		if err := tenantRepo.SetConfig(ctx, models.PlatformTiktok, "sellerName", tokenResp.Data.SellerName, false); err != nil {
			log.Printf("[TikTok OAuth] Warning: failed to save sellerName: %v", err)
		}
	}

	// Save last_refresh timestamp
	if err := tenantRepo.SetConfig(ctx, models.PlatformTiktok, "last_refresh", time.Now().Format(time.RFC3339), false); err != nil {
		log.Printf("[TikTok OAuth] Warning: failed to save last_refresh: %v", err)
	}

	log.Printf("[TikTok OAuth] Successfully saved tokens for tenant %s", tenantID)
	return nil
}
