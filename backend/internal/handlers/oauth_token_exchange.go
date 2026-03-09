package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
)

// exchangeShopeeToken exchanges Shopee auth code for tokens
func (h *OAuthHandler) exchangeShopeeToken(c *gin.Context, tenantID, code, shopIDStr string) error {
	ctx := c.Request.Context()

	// Validate callback params
	if code == "" {
		return fmt.Errorf("authorization code is required")
	}
	if shopIDStr == "" {
		return fmt.Errorf("shop_id is required")
	}

	// Get Shopee credentials
	if h.configRepo == nil {
		return fmt.Errorf("failed to get Shopee credentials: config repository not configured")
	}
	creds, err := h.configRepo.GetShopeeCredentials(ctx)
	if err != nil {
		return fmt.Errorf("failed to get Shopee credentials: %w", err)
	}

	callbackURL := fmt.Sprintf("%s/api/platform-auth/callback/shopee", h.getBackendURL(c))
	shopeeService := oauth.NewShopeeOAuthService(creds.PartnerID, creds.PartnerKey, callbackURL, false)

	// Parse shop ID
	shopID, err := shopeeService.ParseShopID(shopIDStr)
	if err != nil {
		return fmt.Errorf("invalid shop_id: %w", err)
	}

	// Build token request
	tokenURL := shopeeService.GetTokenURL()
	reqBody := shopeeService.BuildTokenRequest(code, shopID)

	// Make HTTP request and parse response
	tokenResp, err := h.doShopeeTokenRequest(tokenURL, reqBody)
	if err != nil {
		return err
	}

	// Save tokens to tenant database
	return h.saveShopeeTokens(ctx, tenantID, shopID, tokenResp)
}

// doShopeeTokenRequest performs the HTTP request to Shopee token endpoint
func (h *OAuthHandler) doShopeeTokenRequest(tokenURL string, body map[string]interface{}) (*ShopeeTokenResponse, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	log.Printf("[Shopee OAuth] Exchanging code for token at: %s", tokenURL)

	resp, err := http.Post(tokenURL, "application/json", strings.NewReader(string(bodyJSON)))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	log.Printf("[Shopee OAuth] Token response status: %d, body_size: %d bytes", resp.StatusCode, len(respBody))

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status %d", resp.StatusCode)
	}

	var tokenResp ShopeeTokenResponse
	if err := json.Unmarshal(respBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("shopee token error: %s - %s", tokenResp.Error, tokenResp.Message)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("no access token in response")
	}

	return &tokenResp, nil
}

// saveShopeeTokens saves Shopee tokens to the tenant database
func (h *OAuthHandler) saveShopeeTokens(ctx context.Context, tenantID string, shopID int64, tokenResp *ShopeeTokenResponse) error {
	tenantDB, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return fmt.Errorf("failed to get tenant database: %w", err)
	}

	tenantRepo := repositories.NewTenantPlatformConfigRepository(tenantDB)

	// Shopee expire_in is relative seconds (typically 14400 = 4 hours)
	expiresIn := tokenResp.ExpireIn
	if expiresIn <= 0 {
		expiresIn = 14400 // Default 4 hours
		log.Printf("[Shopee OAuth] expire_in not provided, using default: %d seconds", expiresIn)
	}
	// Shopee refresh token valid for 7 days per API documentation
	refreshExpiresIn := int64(7 * 24 * 60 * 60)

	log.Printf("[Shopee OAuth] expire_in: %d seconds (%d hours)", expiresIn, expiresIn/3600)

	if err := tenantRepo.UpdateTokens(ctx, models.PlatformShopee, tokenResp.AccessToken, tokenResp.RefreshToken, expiresIn, refreshExpiresIn); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	// Save shop_id
	if err := tenantRepo.SetConfig(ctx, models.PlatformShopee, "shopId", fmt.Sprintf("%d", shopID), false); err != nil {
		log.Printf("[Shopee OAuth] Warning: failed to save shopId: %v", err)
	}

	// Save last_refresh timestamp
	if err := tenantRepo.SetConfig(ctx, models.PlatformShopee, "last_refresh", time.Now().Format(time.RFC3339), false); err != nil {
		log.Printf("[Shopee OAuth] Warning: failed to save last_refresh: %v", err)
	}

	log.Printf("[Shopee OAuth] Successfully saved tokens for tenant %s, shop %d", tenantID, shopID)
	return nil
}
