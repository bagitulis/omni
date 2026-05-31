package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/rs/zerolog/log"
)

// redirectWithError sends a redirect to the frontend with an error query parameter.
func redirectWithError(c *gin.Context, frontendURL, redirectPath, errorKey string) {
	target := frontendURL
	if redirectPath != "" {
		target += redirectPath
	}
	c.Redirect(http.StatusTemporaryRedirect, target+"?error="+errorKey)
}

// HandleCredentialCallback processes the Shopee OAuth callback.
// It parses the signed state, validates the attempt, exchanges the auth code for tokens,
// persists the connection, and redirects the user back to the frontend.
func (s *CredentialApiService) HandleCredentialCallback(c *gin.Context) {
	ctx := c.Request.Context()
	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:5173"
	}

	// Step 1: Parse and validate state
	stateStr := c.Query("state")
	if stateStr == "" {
		redirectWithError(c, frontendURL, "", "missing_state")
		return
	}

	claims, err := oauth.ParseSignedState(stateStr)
	if err != nil {
		if strings.Contains(err.Error(), "expired_state") {
			redirectWithError(c, frontendURL, "", "expired_state")
			return
		}
		redirectWithError(c, frontendURL, "", "invalid_state")
		return
	}

	redirectPath := claims.RedirectPath

	// Step 2: Check for Shopee error/cancel
	if errParam := c.Query("error"); errParam == "access_denied" {
		redirectWithError(c, frontendURL, redirectPath, "access_denied")
		return
	}

	code := c.Query("code")
	if code == "" {
		redirectWithError(c, frontendURL, redirectPath, "user_cancelled")
		return
	}

	shopIDStr := c.Query("shop_id")
	if shopIDStr == "" {
		redirectWithError(c, frontendURL, redirectPath, "missing_shop_id")
		return
	}

	// Step 3: Validate attempt
	repo := repositories.NewCredentialRepository(s.db)
	attempt, err := repo.GetAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID)
	if err != nil || attempt == nil || attempt.Status != "pending" {
		redirectWithError(c, frontendURL, redirectPath, "invalid_attempt")
		return
	}

	// Step 4: Exchange auth code for tokens
	appConfig, err := repo.GetAppConfig(ctx, claims.TenantID, claims.Platform)
	if err != nil || appConfig == nil {
		redirectWithError(c, frontendURL, redirectPath, "app_not_configured")
		return
	}

	isSandbox := os.Getenv("SHOPEE_ENV") != "live"
	callbackBaseURL := os.Getenv("APP_URL")
	if callbackBaseURL == "" {
		callbackBaseURL = "https://yndigital.my.id"
		log.Warn().Msg("APP_URL not set, using fallback for callback URL")
	}
	callbackURL := callbackBaseURL + "/api/credentials/callback/shopee"
	shopeeService := oauth.NewShopeeOAuthService(appConfig.PartnerID, appConfig.PartnerKey, callbackURL, isSandbox)

	shopIDInt, err := shopeeService.ParseShopID(shopIDStr)
	if err != nil {
		log.Error().Err(err).Str("shop_id", shopIDStr).Msg("Failed to parse shop_id")
		repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
		redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
		return
	}

	tokenResp, err := exchangeShopeeToken(ctx, shopeeService, code, shopIDInt)
	if err != nil {
		log.Error().Err(err).Msg("Token exchange failed")
		repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
		redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
		return
	}

	// Step 5: Persist connection
	storeIdentifier := fmt.Sprintf("%d", shopIDInt)
	tokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).Unix()
	refreshExpiry := time.Now().Add(7 * 24 * time.Hour).Unix()

	conn := &models.CredentialConnection{
		TenantID:        claims.TenantID,
		Platform:        claims.Platform,
		StoreIdentifier: storeIdentifier,
		Status:          "connected",
		AccessToken:     tokenResp.AccessToken,
		RefreshToken:    tokenResp.RefreshToken,
		TokenExpiry:     tokenExpiry,
		RefreshExpiry:   refreshExpiry,
		CreatedBy:       claims.UserID,
		UpdatedBy:       claims.UserID,
	}

	existing, _ := repo.GetConnection(ctx, claims.TenantID, claims.Platform, storeIdentifier)
	if existing != nil {
		conn.Region = existing.Region
		if err := repo.ReconnectWithToken(ctx, conn); err != nil {
			log.Error().Err(err).Msg("Failed to reconnect existing connection")
			repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
			redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
			return
		}
	} else {
		if err := repo.CreateConnection(ctx, conn); err != nil {
			log.Error().Err(err).Msg("Failed to create connection")
			repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
			redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
			return
		}
	}

	// Step 6: Complete attempt + audit event
	repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "completed")

	auditEvent := &models.CredentialAuditEvent{
		TenantID:        claims.TenantID,
		Platform:        claims.Platform,
		StoreIdentifier: storeIdentifier,
		EventType:       "oauth_callback",
		Status:          "success",
		Actor:           claims.UserID,
		Metadata:        models.JSONMap{"intent": claims.Intent, "attempt_id": claims.AttemptID},
	}
	if err := repo.CreateAuditEvent(ctx, auditEvent); err != nil {
		log.Error().Err(err).Msg("Failed to create audit event")
	}

	// Redirect to frontend with success
	target := frontendURL
	if redirectPath != "" {
		target += redirectPath
	}
	c.Redirect(http.StatusTemporaryRedirect, target+"?success=true")
}

// shopeeTokenResponse represents the Shopee token exchange response.
type shopeeTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
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
