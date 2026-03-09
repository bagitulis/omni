package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
)

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

	log.Printf("[TikTok OAuth] Token response status: %d, body_size: %d bytes", resp.StatusCode, len(body))

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

	// TikTok API returns expiry as relative seconds from NOW (not Unix timestamps)
	// Example: access_token_expire_in: 604800 (7 days in seconds)
	//          refresh_token_expire_in: 15552000 (180 days in seconds)

	// Handle access token expiry
	expiresInSeconds := tokenResp.Data.AccessTokenExpireIn
	if expiresInSeconds <= 0 {
		// Default to 7 days if not provided
		expiresInSeconds = 7 * 24 * 60 * 60
		log.Printf("[TikTok OAuth] access_token_expire_in not provided, using default: %d seconds (7 days)", expiresInSeconds)
	} else {
		log.Printf("[TikTok OAuth] access_token_expire_in: %d seconds (%d days)", expiresInSeconds, expiresInSeconds/86400)
	}

	// Handle refresh token expiry
	refreshExpiresInSeconds := tokenResp.Data.RefreshTokenExpireIn
	nowSec := time.Now().Unix()
	if refreshExpiresInSeconds > nowSec {
		// It's a Unix timestamp (e.g. TikTok sentinel 4875922303 = year 2124) — convert to relative seconds
		refreshExpiresInSeconds = refreshExpiresInSeconds - nowSec
		log.Printf("[TikTok OAuth] refresh_token_expire_in was Unix timestamp, converted to relative: %d seconds (%d days)", refreshExpiresInSeconds, refreshExpiresInSeconds/86400)
	}
	// Cap at 365 days max — guards against bogus far-future sentinel values
	const maxTiktokRefreshSeconds = int64(365 * 24 * 60 * 60)
	if refreshExpiresInSeconds <= 0 {
		// Default to 90 days if not provided
		refreshExpiresInSeconds = 90 * 24 * 60 * 60
		log.Printf("[TikTok OAuth] refresh_token_expire_in not provided, using default: %d seconds (90 days)", refreshExpiresInSeconds)
	} else if refreshExpiresInSeconds > maxTiktokRefreshSeconds {
		log.Printf("[TikTok OAuth] refresh_token_expire_in %d seconds (%d days) exceeds 365 days cap, capping", refreshExpiresInSeconds, refreshExpiresInSeconds/86400)
		refreshExpiresInSeconds = maxTiktokRefreshSeconds
	} else {
		log.Printf("[TikTok OAuth] refresh_token_expire_in: %d seconds (%d days)", refreshExpiresInSeconds, refreshExpiresInSeconds/86400)
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
