package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/rs/zerolog/log"
)

// shopeeTokenResponse represents the Shopee token exchange response.
type shopeeTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	ShopID       int64  `json:"shop_id"`
	Error        string `json:"error"`
}

// exchangeShopeeToken performs the HTTP POST to exchange an auth code for tokens.
func exchangeShopeeToken(ctx context.Context, shopeeService *oauth.ShopeeOAuthService, code string, shopID int64) (*shopeeTokenResponse, error) {
	tokenURL := shopeeService.GetTokenURL()
	reqBody := shopeeService.BuildTokenRequest(code, shopID)
	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal token request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("exchangeShopeeToken: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp shopeeTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("shopee token error: %s", tokenResp.Error)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token in response")
	}

	return &tokenResp, nil
}

// handleShopeeCallback handles the Shopee-specific OAuth callback logic.
// It builds the callback URL, creates the Shopee OAuth service, parses the shop ID,
// exchanges the auth code for tokens, and returns a CredentialConnection struct.
// The caller is responsible for persisting the connection.
func handleShopeeCallback(ctx context.Context, appConfig *models.CredentialAppConfig, code, shopIDStr string, claims oauth.StateClaims) (*models.CredentialConnection, error) {
	isSandbox := os.Getenv("SHOPEE_ENV") != "live"
	callbackBaseURL := os.Getenv("APP_URL")
	if callbackBaseURL == "" {
		callbackBaseURL = "https://yndigital.my.id"
		log.Warn().Msg("APP_URL not set, using fallback for Shopee callback URL")
	}
	callbackURL := callbackBaseURL + "/api/credentials/callback/shopee"

	shopeeService := oauth.NewShopeeOAuthService(appConfig.PartnerID, appConfig.PartnerKey, callbackURL, isSandbox)

	shopIDInt, err := shopeeService.ParseShopID(shopIDStr)
	if err != nil {
		return nil, fmt.Errorf("parse shop_id: %w", err)
	}

	tokenResp, err := exchangeShopeeToken(ctx, shopeeService, code, shopIDInt)
	if err != nil {
		return nil, fmt.Errorf("token exchange: %w", err)
	}

	storeIdentifier := fmt.Sprintf("%d", shopIDInt)
	tokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).UnixMilli()
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour).UnixMilli()

	conn := &models.CredentialConnection{
		TenantID:        claims.TenantID,
		Platform:        claims.Platform,
		StoreIdentifier: storeIdentifier,
		StoreName:       fmt.Sprintf("Shopee Shop %d", shopIDInt),
		Status:          "connected",
		AccessToken:     tokenResp.AccessToken,
		RefreshToken:    tokenResp.RefreshToken,
		TokenExpiry:     tokenExpiry,
		RefreshExpiry:   refreshExpiry,
		CreatedBy:       claims.UserID,
		UpdatedBy:       claims.UserID,
	}

	return conn, nil
}
