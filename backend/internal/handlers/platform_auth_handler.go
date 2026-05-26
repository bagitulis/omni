package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	credentialsvc "github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/services/oauth"
)

// PlatformAuthHandler handles platform auth status endpoints
type PlatformAuthHandler struct {
	configRepo        *repositories.GlobalConfigRepository
	frontendURL       string
	basePath          string
	credentialService *credentialsvc.CredentialApiService
}

// NewPlatformAuthHandler creates a new platform auth handler
func NewPlatformAuthHandler(
	configRepo *repositories.GlobalConfigRepository,
	frontendURL string,
	basePath string,
	credentialService *credentialsvc.CredentialApiService,
) *PlatformAuthHandler {
	return &PlatformAuthHandler{
		configRepo:        configRepo,
		frontendURL:       frontendURL,
		basePath:          basePath,
		credentialService: credentialService,
	}
}

// OAuthURLInfo represents OAuth URL for a platform
type OAuthURLInfo struct {
	Platform string `json:"platform"`
	AuthURL  string `json:"auth_url"`
	Status   string `json:"status"`
}

// GetOAuthURLs handles GET /api/platform-auth/urls
func (h *PlatformAuthHandler) GetOAuthURLs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	ctx := c.Request.Context()
	backendURL := h.getBackendURL(c)

	urls := make([]OAuthURLInfo, 0, 3)

	// Shopee OAuth URL
	shopeeCreds, err := h.configRepo.GetShopeeCredentials(ctx)
	if err == nil && shopeeCreds != nil {
		callbackURL := fmt.Sprintf("%s/api/platform-auth/callback/shopee", backendURL)
		shopeeService := oauth.NewShopeeOAuthService(shopeeCreds.PartnerID, shopeeCreds.PartnerKey, callbackURL, false)
		urls = append(urls, OAuthURLInfo{
			Platform: models.PlatformShopee,
			AuthURL:  shopeeService.GetAuthURL(""),
			Status:   "ready",
		})
	}

	// Lazada OAuth URL
	lazadaCreds, err := h.configRepo.GetLazadaCredentials(ctx)
	if err == nil && lazadaCreds != nil {
		callbackURL := fmt.Sprintf("%s/api/platform-auth/callback/lazada", backendURL)
		lazadaService := oauth.NewLazadaOAuthService(lazadaCreds.AppKey, lazadaCreds.AppSecret, callbackURL, false)
		urls = append(urls, OAuthURLInfo{
			Platform: models.PlatformLazada,
			AuthURL:  lazadaService.GetAuthURL(""),
			Status:   "ready",
		})
	}

	// TikTok OAuth URL
	tiktokCreds, err := h.configRepo.GetTiktokCredentials(ctx)
	if err == nil && tiktokCreds != nil {
		tiktokService := oauth.NewTiktokOAuthService(tiktokCreds.AppKey, tiktokCreds.AppSecret, "", false)
		urls = append(urls, OAuthURLInfo{
			Platform: models.PlatformTiktok,
			AuthURL:  tiktokService.GetAuthURL(""),
			Status:   "ready",
		})
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"urls": urls}))
}

// DisconnectShopee handles GET /api/platform-auth/shopee/disconnect
func (h *PlatformAuthHandler) DisconnectShopee(c *gin.Context) {
	h.disconnectPlatform(c, models.PlatformShopee)
}

// DisconnectLazada handles GET /api/platform-auth/lazada/disconnect
func (h *PlatformAuthHandler) DisconnectLazada(c *gin.Context) {
	h.disconnectPlatform(c, models.PlatformLazada)
}

// disconnectPlatform disconnects a platform for a tenant
func (h *PlatformAuthHandler) disconnectPlatform(c *gin.Context, platform string) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewPlatformConfigAdapter(db)
	cfg, err := repo.FindByTenantAndPlatform(c.Request.Context(), tenantID, platform)
	if err != nil || cfg == nil {
		c.JSON(http.StatusNotFound, response.Error("Platform not connected"))
		return
	}

	if err := repo.DeleteByPlatform(c.Request.Context(), platform); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to disconnect"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message":  fmt.Sprintf("%s disconnected successfully", platform),
		"platform": platform,
	}))
}

// ConnectionStatus represents platform connection status
type ConnectionStatus struct {
	Platform              string    `json:"platform"`
	Connected             bool      `json:"connected"`
	ShopID                string    `json:"shop_id,omitempty"`
	ShopName              string    `json:"shop_name,omitempty"`
	ExpiresAt             int64     `json:"expires_at,omitempty"`
	RefreshTokenExpiresAt int64     `json:"refresh_token_expires_at,omitempty"`
	ExpiresSoon           bool      `json:"expires_soon,omitempty"`
	Expired               bool      `json:"expired,omitempty"`
	LastChecked           time.Time `json:"last_checked"`
}

// CheckAllConnections handles POST /api/platform-auth/check-all
func (h *PlatformAuthHandler) CheckAllConnections(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	// Use TenantPlatformConfigRepository which uses key-value pattern (NOT PlatformConfigRepository!)
	repo := repositories.NewTenantPlatformConfigRepository(db)

	now := time.Now()
	statuses := make(map[string]ConnectionStatus)
	platforms := []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}

	// Get config for each platform using key-value pattern
	for _, platform := range platforms {
		configMap, err := repo.GetAllConfigByPlatform(c.Request.Context(), platform)
		if err != nil {
			statuses[platform] = ConnectionStatus{
				Platform:    platform,
				Connected:   false,
				LastChecked: now,
			}
			continue
		}

		// Check if connected (has access token)
		accessToken := configMap["accessToken"]
		connected := accessToken != ""

		// Parse expiry timestamp
		var expiresAt int64
		if expStr, ok := configMap["tokenExpiry"]; ok && expStr != "" {
			fmt.Sscanf(expStr, "%d", &expiresAt)
		}

		// Parse refresh token expiry
		var refreshTokenExpiresAt int64
		if expStr, ok := configMap["refreshTokenExpiry"]; ok && expStr != "" {
			fmt.Sscanf(expStr, "%d", &refreshTokenExpiresAt)
		}

		// Normalize to milliseconds - if value is too small, it's in seconds
		expiresAtMs := expiresAt
		if expiresAt > 0 && expiresAt < 1_000_000_000_000 {
			expiresAtMs = expiresAt * 1000
		}

		refreshTokenExpiresAtMs := refreshTokenExpiresAt
		if refreshTokenExpiresAt > 0 && refreshTokenExpiresAt < 1_000_000_000_000 {
			refreshTokenExpiresAtMs = refreshTokenExpiresAt * 1000
		}

		// Calculate expiry status using milliseconds
		nowMs := now.UnixMilli()
		expiresSoon := expiresAtMs > 0 && expiresAtMs < nowMs+(24*60*60*1000)
		expired := expiresAtMs > 0 && expiresAtMs < nowMs

		// Get shop info
		shopID := configMap["shopId"]
		shopName := configMap["shopName"]

		statuses[platform] = ConnectionStatus{
			Platform:              platform,
			Connected:             connected,
			ShopID:                shopID,
			ShopName:              shopName,
			ExpiresAt:             expiresAtMs,
			RefreshTokenExpiresAt: refreshTokenExpiresAtMs,
			ExpiresSoon:           expiresSoon,
			Expired:               expired,
			LastChecked:           now,
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"connections": statuses,
		"checked_at":  now,
	}))
}

// GetCredentialPlatforms handles GET /api/credentials/platforms
func (h *PlatformAuthHandler) GetCredentialPlatforms(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	if h.credentialService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("Credential service unavailable"))
		return
	}

	platform := c.Query("platform")
	storeIdentifier := c.Query("store_identifier")
	result, err := h.credentialService.GetPlatformStatus(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), platform, storeIdentifier)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "Missing tenant_id" || err.Error() == "tenant_id is required" {
			status = http.StatusUnauthorized
		}
		c.JSON(status, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"platforms": result}))
}

// PutCredentialApp handles PUT /api/credentials/platforms/:platform/app
func (h *PlatformAuthHandler) PutCredentialApp(c *gin.Context) {
	h.upsertCredentialApp(c, false)
}

// RotateCredentialApp handles POST /api/credentials/platforms/:platform/app/rotate
func (h *PlatformAuthHandler) RotateCredentialApp(c *gin.Context) {
	h.upsertCredentialApp(c, true)
}

// PostCredentialOAuthInitiate handles POST /api/credentials/platforms/:platform/connections/oauth/initiate
func (h *PlatformAuthHandler) PostCredentialOAuthInitiate(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	if h.credentialService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("Credential service unavailable"))
		return
	}
	var req credentialsvc.CredentialOAuthInitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	req.Platform = c.Param("platform")
	result, err := h.credentialService.InitiateOAuth(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// PostCredentialOAuthReconnect handles POST /api/credentials/platforms/:platform/connections/:store_identifier/oauth/reconnect
func (h *PlatformAuthHandler) PostCredentialOAuthReconnect(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	if h.credentialService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("Credential service unavailable"))
		return
	}
	req := credentialsvc.CredentialOAuthReconnectRequest{Platform: c.Param("platform"), StoreIdentifier: c.Param("store_identifier")}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	req.Platform = c.Param("platform")
	req.StoreIdentifier = c.Param("store_identifier")
	result, err := h.credentialService.ReconnectOAuth(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// PostCredentialRefresh handles POST /api/credentials/platforms/:platform/connections/:store_identifier/refresh
func (h *PlatformAuthHandler) PostCredentialRefresh(c *gin.Context) {
	h.changeConnectionStatus(c, "refresh")
}

// PostCredentialDisconnect handles POST /api/credentials/platforms/:platform/connections/:store_identifier/disconnect
func (h *PlatformAuthHandler) PostCredentialDisconnect(c *gin.Context) {
	h.changeConnectionStatus(c, "disconnect")
}

// GetCredentialAudit handles GET /api/credentials/platforms/:platform/audit
func (h *PlatformAuthHandler) GetCredentialAudit(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	if h.credentialService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("Credential service unavailable"))
		return
	}
	result, err := h.credentialService.ListAuditEvents(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), c.Param("platform"), c.Query("store_identifier"), c.Query("limit"), c.Query("cursor"))
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// PostCredentialManualToken handles POST /api/credentials/platforms/:platform/connections/manual-token
func (h *PlatformAuthHandler) PostCredentialManualToken(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	if h.credentialService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("Credential service unavailable"))
		return
	}
	var req credentialsvc.CredentialManualTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	req.Platform = c.Param("platform")
	result, err := h.credentialService.ApplyManualToken(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *PlatformAuthHandler) upsertCredentialApp(c *gin.Context, rotate bool) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	if h.credentialService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("Credential service unavailable"))
		return
	}
	var req credentialsvc.CredentialAppUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	req.Platform = c.Param("platform")
	result, err := h.credentialService.UpsertAppCredential(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), req, rotate)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

func (h *PlatformAuthHandler) changeConnectionStatus(c *gin.Context, action string) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	if h.credentialService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("Credential service unavailable"))
		return
	}
	var req credentialsvc.CredentialConnectionActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	req.Platform = c.Param("platform")
	req.StoreIdentifier = c.Param("store_identifier")
	result, err := h.credentialService.ChangeConnectionStatus(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), req, action)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}
