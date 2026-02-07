package handlers

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
)

// OAuthHandler handles OAuth endpoints for all platforms
type OAuthHandler struct {
	oauthRepo    *repositories.OAuthRepository
	configRepo   *repositories.GlobalConfigRepository
	platformRepo *repositories.PlatformConfigRepository
	frontendURL  string
	basePath     string
}

// NewOAuthHandler creates a new OAuth handler
func NewOAuthHandler(
	oauthRepo *repositories.OAuthRepository,
	configRepo *repositories.GlobalConfigRepository,
	platformRepo *repositories.PlatformConfigRepository,
	frontendURL string,
) *OAuthHandler {
	basePath := os.Getenv("DB_BASE_PATH")
	if basePath == "" {
		basePath = "./data"
	}
	return &OAuthHandler{
		oauthRepo:    oauthRepo,
		configRepo:   configRepo,
		platformRepo: platformRepo,
		frontendURL:  frontendURL,
		basePath:     basePath,
	}
}

// InitiateAuth initiates OAuth flow for a platform
func (h *OAuthHandler) InitiateAuth(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantID - authentication required",
		})
		return
	}

	platform := c.Param("platform")
	redirectURL := c.Query("redirect_url")
	if redirectURL == "" {
		redirectURL = h.frontendURL + "/settings"
	}

	var authURL string
	var err error

	switch platform {
	case models.PlatformShopee:
		authURL, err = h.initiateShopeeAuth(c, tenantID, redirectURL)
	case models.PlatformLazada:
		authURL, err = h.initiateLazadaAuth(c, tenantID, redirectURL)
	case models.PlatformTiktok:
		authURL, err = h.initiateTiktokAuth(c, tenantID, redirectURL)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid platform",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"auth_url": authURL,
		},
	})
}

// HandleCallback handles OAuth callback for a platform
func (h *OAuthHandler) HandleCallback(c *gin.Context) {
	platform := c.Param("platform")
	state := c.Query("state")
	code := c.Query("code")

	// Validate state
	oauthState, err := h.oauthRepo.FindStateByState(c.Request.Context(), state)
	if err != nil || oauthState == nil {
		c.Redirect(http.StatusFound, h.frontendURL+"/settings?error=invalid_state")
		return
	}

	// Delete state after validation
	h.oauthRepo.DeleteState(c.Request.Context(), oauthState.ID)

	// Log OAuth attempt
	h.oauthRepo.CreateLog(c.Request.Context(), &models.OAuthLog{
		TenantID:  oauthState.TenantID,
		Platform:  platform,
		EventType: models.OAuthEventCallback,
		Status:    models.OAuthStatusReceived,
	})

	var exchangeErr error

	switch platform {
	case models.PlatformShopee:
		shopID := c.Query("shop_id")
		exchangeErr = h.exchangeShopeeToken(c, oauthState.TenantID, code, shopID)
	case models.PlatformLazada:
		exchangeErr = h.exchangeLazadaToken(c, oauthState.TenantID, code)
	case models.PlatformTiktok:
		exchangeErr = h.exchangeTiktokToken(c, oauthState.TenantID, code)
	default:
		c.Redirect(http.StatusFound, h.frontendURL+"/settings?error=invalid_platform")
		return
	}

	if exchangeErr != nil {
		h.oauthRepo.UpdateLogStatus(c.Request.Context(), oauthState.ID, models.OAuthStatusFailed, exchangeErr.Error())
		c.Redirect(http.StatusFound, fmt.Sprintf("%s/settings?error=%s", h.frontendURL, exchangeErr.Error()))
		return
	}

	c.Redirect(http.StatusFound, oauthState.RedirectURL+"?success=true&platform="+platform)
}

// initiateShopeeAuth initiates Shopee OAuth
func (h *OAuthHandler) initiateShopeeAuth(c *gin.Context, tenantID, redirectURL string) (string, error) {
	creds, err := h.configRepo.GetShopeeCredentials(c.Request.Context())
	if err != nil {
		return "", err
	}

	state, err := h.oauthRepo.CreateState(c.Request.Context(), tenantID, models.PlatformShopee, redirectURL)
	if err != nil {
		return "", err
	}

	callbackURL := fmt.Sprintf("%s/api/platform-auth/callback/shopee", h.getBackendURL(c))
	shopeeService := oauth.NewShopeeOAuthService(creds.PartnerID, creds.PartnerKey, callbackURL, false)

	return shopeeService.GetAuthURL(state.State), nil
}

// initiateLazadaAuth initiates Lazada OAuth
func (h *OAuthHandler) initiateLazadaAuth(c *gin.Context, tenantID, redirectURL string) (string, error) {
	creds, err := h.configRepo.GetLazadaCredentials(c.Request.Context())
	if err != nil {
		return "", err
	}

	state, err := h.oauthRepo.CreateState(c.Request.Context(), tenantID, models.PlatformLazada, redirectURL)
	if err != nil {
		return "", err
	}

	callbackURL := fmt.Sprintf("%s/api/platform-auth/callback/lazada", h.getBackendURL(c))
	lazadaService := oauth.NewLazadaOAuthService(creds.AppKey, creds.AppSecret, callbackURL, false)

	return lazadaService.GetAuthURL(state.State), nil
}

// initiateTiktokAuth initiates TikTok OAuth
func (h *OAuthHandler) initiateTiktokAuth(c *gin.Context, tenantID, redirectURL string) (string, error) {
	creds, err := h.configRepo.GetTiktokCredentials(c.Request.Context())
	if err != nil {
		return "", err
	}

	state, err := h.oauthRepo.CreateState(c.Request.Context(), tenantID, models.PlatformTiktok, redirectURL)
	if err != nil {
		return "", err
	}

	tiktokService := oauth.NewTiktokOAuthService(creds.AppKey, creds.AppSecret, "", false)

	return tiktokService.GetAuthURL(state.State), nil
}

// getBackendURL gets backend URL from request
func (h *OAuthHandler) getBackendURL(c *gin.Context) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}

// GetOAuthLogs gets OAuth logs for a tenant
func (h *OAuthHandler) GetOAuthLogs(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantID - authentication required",
		})
		return
	}

	platform := c.Query("platform")
	limit := 50

	var logs []models.OAuthLog
	var err error

	if platform != "" {
		logs, err = h.oauthRepo.FindLogsByPlatform(c.Request.Context(), tenantID, platform, limit)
	} else {
		logs, err = h.oauthRepo.FindLogsByTenant(c.Request.Context(), tenantID, limit)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    logs,
	})
}

// GetTokenStatus returns token status for all platforms
// GET /api/oauth/status
func (h *OAuthHandler) GetTokenStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantID - authentication required",
		})
		return
	}

	// Get tenant database connection
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database connection failed",
		})
		return
	}

	// Use TenantPlatformConfigRepository which uses key-value pattern
	repo := repositories.NewTenantPlatformConfigRepository(db)

	nowMs := time.Now().UnixMilli()
	result := make(map[string]gin.H)
	platforms := []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}

	// Get config for each platform using key-value pattern
	for _, platform := range platforms {
		configMap, err := repo.GetAllConfigByPlatform(c.Request.Context(), platform)
		if err != nil {
			result[platform] = gin.H{
				"isExpired":             true,
				"expiresAt":             nil,
				"refreshTokenExpiresAt": nil,
				"status":                "not_configured",
				"valid":                 false,
			}
			continue
		}

		// Check if connected (has access token)
		accessToken := configMap["accessToken"]
		connected := accessToken != ""

		if !connected {
			result[platform] = gin.H{
				"isExpired":             true,
				"expiresAt":             nil,
				"refreshTokenExpiresAt": nil,
				"status":                "not_configured",
				"valid":                 false,
			}
			continue
		}

		// Parse token expiry timestamp (stored in milliseconds)
		var tokenExpiryMs int64
		if expStr, ok := configMap["tokenExpiry"]; ok && expStr != "" {
			fmt.Sscanf(expStr, "%d", &tokenExpiryMs)
		}

		// Parse refresh token expiry timestamp (stored in milliseconds)
		var refreshTokenExpiryMs int64
		if expStr, ok := configMap["refreshTokenExpiry"]; ok && expStr != "" {
			fmt.Sscanf(expStr, "%d", &refreshTokenExpiryMs)
		}

		// Calculate expiry status
		isExpired := tokenExpiryMs > 0 && tokenExpiryMs < nowMs
		expiresSoon := tokenExpiryMs > 0 && tokenExpiryMs < nowMs+(24*60*60*1000)

		status := "valid"
		if isExpired {
			status = "expired"
		} else if expiresSoon {
			status = "expiring"
		}

		// Convert timestamps to RFC3339 format for frontend
		var expiresAtStr, refreshTokenExpiresAtStr string
		if tokenExpiryMs > 0 {
			expiresAtStr = time.UnixMilli(tokenExpiryMs).Format(time.RFC3339)
		}
		if refreshTokenExpiryMs > 0 {
			refreshTokenExpiresAtStr = time.UnixMilli(refreshTokenExpiryMs).Format(time.RFC3339)
		}

		// Get shop info
		shopID := configMap["shopId"]
		shopName := configMap["shopName"]

		result[platform] = gin.H{
			"isExpired":             isExpired,
			"expiresAt":             expiresAtStr,
			"refreshTokenExpiresAt": refreshTokenExpiresAtStr,
			"status":                status,
			"valid":                 !isExpired,
			"shopId":                shopID,
			"shopName":              shopName,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// RefreshAllTokens refreshes tokens for all platforms
// POST /api/oauth/refresh-all
func (h *OAuthHandler) RefreshAllTokens(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantID - authentication required",
		})
		return
	}

	configs, err := h.platformRepo.FindByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	results := make(map[string]gin.H)
	for _, config := range configs {
		// TODO: Implement actual token refresh for each platform
		// For now, return status based on current expiry
		now := time.Now().Unix()
		isExpired := config.ExpiresAt < now

		if isExpired {
			results[config.Platform] = gin.H{
				"success": false,
				"error":   "Token expired - manual re-authorization required",
			}
		} else {
			results[config.Platform] = gin.H{
				"success": true,
				"message": "Token is still valid",
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
	})
}
