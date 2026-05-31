package services

import (
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

// HandleCredentialCallback processes OAuth callbacks from Shopee, Lazada, and TikTok.
// It parses the signed state, validates the attempt, dispatches to the platform-specific
// token exchange, persists the connection, dual-writes to platform_configs, and redirects.
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

	// Step 2: Check for error/cancel
	if errParam := c.Query("error"); errParam == "access_denied" {
		redirectWithError(c, frontendURL, redirectPath, "access_denied")
		return
	}

	code := c.Query("code")
	if code == "" {
		redirectWithError(c, frontendURL, redirectPath, "user_cancelled")
		return
	}

	// Step 2.5: Platform validation — prevent URL path manipulation
	urlPlatform := c.Param("platform")
	if claims.Platform != urlPlatform {
		log.Warn().
			Str("claims_platform", claims.Platform).
			Str("url_platform", urlPlatform).
			Msg("Platform mismatch between state claims and URL param")
		redirectWithError(c, frontendURL, redirectPath, "platform_mismatch")
		return
	}

	// Step 3: Validate attempt
	repo := repositories.NewCredentialRepository(s.db)
	attempt, err := repo.GetAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID)
	if err != nil || attempt == nil || attempt.Status != "pending" {
		redirectWithError(c, frontendURL, redirectPath, "invalid_attempt")
		return
	}

	// Step 4: Fetch app config
	appConfig, err := repo.GetAppConfig(ctx, claims.TenantID, claims.Platform)
	if err != nil || appConfig == nil {
		redirectWithError(c, frontendURL, redirectPath, "app_not_configured")
		return
	}

	// Step 5: Build callback URL base
	callbackBaseURL := os.Getenv("APP_URL")
	if callbackBaseURL == "" {
		callbackBaseURL = "https://yndigital.my.id"
		log.Warn().Msg("APP_URL not set, using fallback for callback URL")
	}

	// Step 6: Platform dispatch — each case builds a CredentialConnection struct
	var conn *models.CredentialConnection

	switch claims.Platform {
	case "shopee":
		shopIDStr := c.Query("shop_id")
		if shopIDStr == "" {
			redirectWithError(c, frontendURL, redirectPath, "missing_shop_id")
			return
		}
		conn, err = handleShopeeCallback(ctx, appConfig, code, shopIDStr, claims)
		if err != nil {
			log.Error().Err(err).Msg("Shopee callback failed")
			repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
			redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
			return
		}

	case "lazada":
		callbackURL := callbackBaseURL + "/api/credentials/callback/lazada"
		lazadaService := oauth.NewLazadaOAuthService(appConfig.AppKey, appConfig.AppSecret, callbackURL, false)

		tokenResp, tokenErr := exchangeLazadaToken(ctx, lazadaService, code)
		if tokenErr != nil {
			log.Error().Err(tokenErr).Msg("Lazada token exchange failed")
			repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
			redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
			return
		}

		storeIdentifier, storeName, fetchErr := fetchLazadaStoreIdentifier(ctx, lazadaService, tokenResp.AccessToken)
		if fetchErr != nil {
			log.Error().Err(fetchErr).Msg("Lazada store identifier fetch failed")
			repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
			redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
			return
		}

		tokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).UnixMilli()
		refreshExpiry := time.Now().Add(time.Duration(tokenResp.RefreshExpiresIn) * time.Second).UnixMilli()

		conn = &models.CredentialConnection{
			TenantID:        claims.TenantID,
			Platform:        claims.Platform,
			StoreIdentifier: storeIdentifier,
			StoreName:       storeName,
			Status:          "connected",
			AccessToken:     tokenResp.AccessToken,
			RefreshToken:    tokenResp.RefreshToken,
			TokenExpiry:     tokenExpiry,
			RefreshExpiry:   refreshExpiry,
			CreatedBy:       claims.UserID,
			UpdatedBy:       claims.UserID,
		}

	case "tiktok":
		callbackURL := callbackBaseURL + "/api/credentials/callback/tiktok"
		tiktokService := oauth.NewTiktokOAuthService(appConfig.AppKey, appConfig.AppSecret, callbackURL, false)

		tokenResp, tokenErr := exchangeTiktokToken(ctx, tiktokService, code)
		if tokenErr != nil {
			log.Error().Err(tokenErr).Msg("TikTok token exchange failed")
			repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
			redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
			return
		}

		storeIdentifier, shopCipher, storeName, fetchErr := fetchTiktokStoreIdentifier(ctx, tiktokService, tokenResp.AccessToken)
		if fetchErr != nil {
			log.Error().Err(fetchErr).Msg("TikTok store identifier fetch failed")
			repo.CompleteAttempt(ctx, claims.TenantID, claims.Platform, claims.AttemptID, "failed")
			redirectWithError(c, frontendURL, redirectPath, "token_exchange_failed")
			return
		}

		tokenExpiry := time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second).UnixMilli()
		refreshExpiry := time.Now().Add(time.Duration(tokenResp.RefreshExpiresIn) * time.Second).UnixMilli()

		conn = &models.CredentialConnection{
			TenantID:        claims.TenantID,
			Platform:        claims.Platform,
			StoreIdentifier: storeIdentifier,
			StoreName:       storeName,
			ShopCipher:      shopCipher,
			Status:          "connected",
			AccessToken:     tokenResp.AccessToken,
			RefreshToken:    tokenResp.RefreshToken,
			TokenExpiry:     tokenExpiry,
			RefreshExpiry:   refreshExpiry,
			CreatedBy:       claims.UserID,
			UpdatedBy:       claims.UserID,
		}

	default:
		redirectWithError(c, frontendURL, redirectPath, "unsupported_platform")
		return
	}

	// Step 7: Persist connection
	storeIdentifier := conn.StoreIdentifier
	existing, getErr := repo.GetConnection(ctx, claims.TenantID, claims.Platform, storeIdentifier)
	if getErr != nil {
		log.Error().Err(getErr).Str("tenant_id", claims.TenantID).Str("platform", claims.Platform).Msg("failed to check existing connection")
		redirectWithError(c, frontendURL, redirectPath, "connection_check_failed")
		return
	}
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

	// Step 8: Dual-write to platform_configs (best-effort with retry)
	platformConfigRepo := repositories.NewTenantPlatformConfigRepository(s.db)
	expiresInSeconds := (conn.TokenExpiry - time.Now().UnixMilli()) / 1000
	refreshExpiresInSeconds := (conn.RefreshExpiry - time.Now().UnixMilli()) / 1000
	if dwErr := retryDualWrite(3, func() error {
		return platformConfigRepo.UpdateTokens(ctx, conn.Platform, conn.AccessToken, conn.RefreshToken, expiresInSeconds, refreshExpiresInSeconds)
	}); dwErr != nil {
		log.Error().Err(dwErr).
			Str("platform", conn.Platform).
			Str("store", storeIdentifier).
			Msg("dual-write to platform_configs FAILED after retries — manual reconciliation may be needed")
	}

	// Step 9: Complete attempt + audit event
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
