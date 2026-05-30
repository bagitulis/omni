package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	credentialsvc "github.com/omni/backend/internal/services"
)

func credentialErrorStatus(err error) int {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "forbidden") || strings.Contains(msg, "not authorized"):
		return http.StatusForbidden
	case strings.Contains(msg, "not found"):
		return http.StatusNotFound
	case strings.HasPrefix(msg, "missing") || strings.Contains(msg, "invalid") || strings.Contains(msg, "unsupported"):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

type CredentialAPIService interface {
	GetPlatformStatus(ctx context.Context, tenantID, role, userID, platform, storeIdentifier string) ([]credentialsvc.CredentialPlatformStatus, error)
	UpsertAppCredential(ctx context.Context, tenantID, role, userID string, req credentialsvc.CredentialAppUpsertRequest, rotate bool) (*credentialsvc.CredentialMutationResponse, error)
	InitiateOAuth(ctx context.Context, tenantID, role, userID, callbackBaseURL string, req credentialsvc.CredentialOAuthInitiateRequest) (*credentialsvc.CredentialOAuthAttemptResponse, error)
	ReconnectOAuth(ctx context.Context, tenantID, role, userID, callbackBaseURL string, req credentialsvc.CredentialOAuthReconnectRequest) (*credentialsvc.CredentialOAuthAttemptResponse, error)
	ChangeConnectionStatus(ctx context.Context, tenantID, role, userID string, req credentialsvc.CredentialConnectionActionRequest, action string) (*credentialsvc.CredentialMutationResponse, error)
	ListAuditEvents(ctx context.Context, tenantID, role, userID, platform, storeIdentifier, limitStr, cursor string) (*credentialsvc.CredentialAuditListResponse, error)
	ApplyManualToken(ctx context.Context, tenantID, role, userID string, req credentialsvc.CredentialManualTokenRequest) (*credentialsvc.CredentialMutationResponse, error)
	HandleCredentialCallback(c *gin.Context)
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
		c.JSON(credentialErrorStatus(err), response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(gin.H{"platforms": result}))
}

// PutCredentialApp handles PUT /api/credentials/platforms/:platform/app
func (h *PlatformAuthHandler) PutCredentialApp(c *gin.Context) { h.upsertCredentialApp(c, false) }

// RotateCredentialApp handles POST /api/credentials/platforms/:platform/app/rotate
func (h *PlatformAuthHandler) RotateCredentialApp(c *gin.Context) { h.upsertCredentialApp(c, true) }

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
	result, err := h.credentialService.InitiateOAuth(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), h.getBackendURL(c), req)
	if err != nil {
		c.JSON(credentialErrorStatus(err), response.Error(err.Error()))
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
	result, err := h.credentialService.ReconnectOAuth(c.Request.Context(), tenantID, c.GetString("role"), c.GetString("userID"), h.getBackendURL(c), req)
	if err != nil {
		c.JSON(credentialErrorStatus(err), response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// PostCredentialRefresh handles POST /api/credentials/platforms/:platform/connections/:store_identifier/refresh
func (h *PlatformAuthHandler) PostCredentialRefresh(c *gin.Context) { h.changeConnectionStatus(c, "refresh") }

// PostCredentialDisconnect handles POST /api/credentials/platforms/:platform/connections/:store_identifier/disconnect
func (h *PlatformAuthHandler) PostCredentialDisconnect(c *gin.Context) { h.changeConnectionStatus(c, "disconnect") }

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
		c.JSON(credentialErrorStatus(err), response.Error(err.Error()))
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
		c.JSON(credentialErrorStatus(err), response.Error(err.Error()))
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
		c.JSON(credentialErrorStatus(err), response.Error(err.Error()))
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
		c.JSON(credentialErrorStatus(err), response.Error(err.Error()))
		return
	}
	c.JSON(http.StatusOK, response.Success(result))
}

// HandleCredentialCallback handles GET /api/credentials/callback/:platform
func (h *PlatformAuthHandler) HandleCredentialCallback(c *gin.Context) {
	if h.credentialService == nil {
		c.JSON(http.StatusInternalServerError, response.Error("Credential service unavailable"))
		return
	}
	h.credentialService.HandleCredentialCallback(c)
}
