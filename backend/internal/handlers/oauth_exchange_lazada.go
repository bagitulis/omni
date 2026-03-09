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

	log.Printf("[Lazada OAuth] Token response status: %d, body_size: %d bytes", resp.StatusCode, len(body))

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

	// Lazada API returns expiry as relative seconds from NOW
	// Example: expires_in: 604800 (7 days in seconds)
	//          refresh_expires_in: 2592000 (30 days in seconds)
	log.Printf("[Lazada OAuth] expires_in: %d seconds (%d days)", tokenResp.ExpiresIn, tokenResp.ExpiresIn/86400)
	log.Printf("[Lazada OAuth] refresh_expires_in: %d seconds (%d days)", tokenResp.RefreshExpiresIn, tokenResp.RefreshExpiresIn/86400)

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
