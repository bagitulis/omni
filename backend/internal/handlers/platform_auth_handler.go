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
	"github.com/omni/backend/internal/services/oauth"
)

// PlatformAuthHandler handles platform auth status endpoints
type PlatformAuthHandler struct {
	configRepo  *repositories.GlobalConfigRepository
	frontendURL string
	basePath    string
}

// NewPlatformAuthHandler creates a new platform auth handler
func NewPlatformAuthHandler(
	configRepo *repositories.GlobalConfigRepository,
	frontendURL string,
	basePath string,
) *PlatformAuthHandler {
	return &PlatformAuthHandler{
		configRepo:  configRepo,
		frontendURL: frontendURL,
		basePath:    basePath,
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewPlatformConfigRepository(db)
	cfg, err := repo.FindByTenantAndPlatform(c.Request.Context(), tenantID, platform)
	if err != nil || cfg == nil {
		c.JSON(http.StatusNotFound, response.Error("Platform not connected"))
		return
	}

	if err := repo.Delete(c.Request.Context(), tenantID, platform); err != nil {
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
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	repo := repositories.NewPlatformConfigRepository(db)
	configs, err := repo.FindByTenant(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get connections"))
		return
	}

	now := time.Now()
	statuses := make(map[string]ConnectionStatus)
	platforms := []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}

	// Initialize all platforms as disconnected
	for _, p := range platforms {
		statuses[p] = ConnectionStatus{
			Platform:    p,
			Connected:   false,
			LastChecked: now,
		}
	}

	// Update with actual status
	for _, cfg := range configs {
		expiresSoon := cfg.ExpiresAt > 0 && cfg.ExpiresAt < now.Add(24*time.Hour).Unix()
		expired := cfg.ExpiresAt > 0 && cfg.ExpiresAt < now.Unix()

		statuses[cfg.Platform] = ConnectionStatus{
			Platform:    cfg.Platform,
			Connected:   cfg.IsActive && cfg.AccessToken != "",
			ShopID:      cfg.ShopID,
			ShopName:    cfg.ShopName,
			ExpiresAt:   cfg.ExpiresAt,
			ExpiresSoon: expiresSoon,
			Expired:     expired,
			LastChecked: now,
		}
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"connections": statuses,
		"checkedAt":   now,
	}))
}
