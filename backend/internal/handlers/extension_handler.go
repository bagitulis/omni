package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/extensions"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
)

// ExtensionHandler exposes the extensions API.
//
// Handlers stay thin: they extract the tenant and user from the request context,
// delegate to the service, and format the response. No business logic lives here
// and no database access happens here.
type ExtensionHandler struct {
	service *extensions.Service
}

// NewExtensionHandler creates an ExtensionHandler.
func NewExtensionHandler(service *extensions.Service) *ExtensionHandler {
	return &ExtensionHandler{service: service}
}

// available reports whether the handler is backed by a service.
//
// Every entry point checks this first and returns 503 when it is not. A nil
// service means the feature was not wired up (or wiring failed), and a Gin
// handler that dereferences nil panics — which in this server would abort the
// request goroutine and can take the process down. Failing closed with 503 is
// both safer and more diagnosable than a panic.
func (h *ExtensionHandler) available(c *gin.Context) bool {
	if h == nil || h.service == nil {
		c.JSON(http.StatusServiceUnavailable, response.Error("Extensions service is not available"))
		return false
	}
	return true
}

// tenantIDFromContext reads the tenant that middleware.Tenant() resolved.
//
// Returns an empty string when absent so each handler can fail closed with 401:
// the platform has no default tenant, and inventing one here would be a
// cross-tenant data leak.
func tenantIDFromContext(c *gin.Context) string {
	return middleware.GetTenantID(c)
}

// List handles GET /api/extensions.
func (h *ExtensionHandler) List(c *gin.Context) {
	// Tenant first, then availability: the security check must not be skipped
	// because a dependency happens to be missing.
	tenantID := tenantIDFromContext(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	if !h.available(c) {
		return
	}

	exts, err := h.service.ListExtensions(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to list extensions"))
		return
	}
	// Built as summaries rather than assigning into the model slice: TokenHash
	// is the credential's stored form and must never leave the server, and a
	// type mix-up here would silently drop rows.
	out := make([]extensionSummary, 0, len(exts))
	for _, e := range exts {
		out = append(out, toExtensionSummary(e))
	}

	c.JSON(http.StatusOK, response.Success(out))
}

// GeneratePairingCode handles POST /api/extensions/pairing/generate.
func (h *ExtensionHandler) GeneratePairingCode(c *gin.Context) {
	// Tenant first, then availability. The security check must not be skipped
	// because a dependency happens to be missing.
	tenantID := tenantIDFromContext(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}
	userID := c.GetString("userID")

	if !h.available(c) {
		return
	}

	code, expiresAt, err := h.service.StartPairing(c.Request.Context(), tenantID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Failed to generate pairing code"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"code":        code,
		"expires_at":  expiresAt.UTC().Format("2006-01-02T15:04:05Z"),
		"ttl_seconds": int(extensions.PairingCodeTTL.Seconds()),
	}))
}

// pairRequestDTO is the pairing confirmation body.
//
// Note the absence of a tenant_id field: the tenant comes from the pairing code,
// which was created by an authenticated user in this tenant's schema.
type pairRequestDTO struct {
	Code             string   `json:"code" binding:"required"`
	ExtensionID      string   `json:"extension_id" binding:"required"`
	Hostname         string   `json:"hostname"`
	BrowserInfo      string   `json:"browser_info"`
	ChromeVersion    string   `json:"chrome_version"`
	ExtensionVersion string   `json:"extension_version"`
	ProtocolVersion  string   `json:"protocol_version"`
	Capabilities     []string `json:"capabilities"`
}

// ConfirmPairing handles POST /api/extensions/pairing/confirm.
//
// This endpoint is deliberately NOT behind middleware.Auth(): the extension has
// no JWT, only the pairing code it was given. It cannot sit behind
// middleware.Tenant() either, because the tenant is what we are trying to
// discover — the code is the only handle the browser holds. The tenant is
// therefore resolved FROM the code, and the code's uniqueness (40 bits of
// CSPRNG, single-use, 5-minute TTL) is what keeps that unambiguous.
//
// This is why the route is registered on the unauthenticated group in
// routes/extension_routes.go rather than the protected one.
func (h *ExtensionHandler) ConfirmPairing(c *gin.Context) {
	// Validate the body before checking service availability: a malformed
	// request is a client error regardless of whether the feature is wired up,
	// and answering 503 for bad input would misdirect the caller.
	var body pairRequestDTO
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid request: "+err.Error()))
		return
	}

	if !h.available(c) {
		return
	}

	tenantID, err := h.service.ResolvePairingTenant(c.Request.Context(), body.Code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, response.Error("Invalid or expired pairing code"))
		return
	}

	token, ext, err := h.service.ConfirmPairing(c.Request.Context(), tenantID, extensions.PairRequest{
		Code:             body.Code,
		ExtensionID:      body.ExtensionID,
		Hostname:         body.Hostname,
		BrowserInfo:      body.BrowserInfo,
		ChromeVersion:    body.ChromeVersion,
		ExtensionVersion: body.ExtensionVersion,
		ProtocolVersion:  body.ProtocolVersion,
		Capabilities:     body.Capabilities,
	})
	if err != nil {
		// Deliberately generic: distinguishing "unknown" from "expired" from
		// "already used" would let a caller probe for valid codes.
		c.JSON(http.StatusUnauthorized, response.Error("Invalid or expired pairing code"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"token":        token,
		"extension_id": ext.ExtensionID,
		"capabilities": ext.Capabilities,
		"ws_path":      extensions.WSPath,
	}))
}

// Unpair handles DELETE /api/extensions/:extension_id.
func (h *ExtensionHandler) Unpair(c *gin.Context) {
	tenantID := tenantIDFromContext(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenant_id"))
		return
	}

	// Validate before reaching the service: an invalid identifier is a client
	// error regardless of service availability, and validating first keeps
	// untrusted input out of any query.
	extensionID := c.Param("extension_id")
	if err := extensions.ValidateExtensionID(extensionID); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("Invalid extension id"))
		return
	}

	if !h.available(c) {
		return
	}

	if err := h.service.Unpair(c.Request.Context(), tenantID, extensionID); err != nil {
		c.JSON(http.StatusNotFound, response.Error("Extension not found"))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{"extension_id": extensionID}))
}

// extensionSummary is the API representation of a paired extension.
//
// TokenHash is deliberately absent — it must never leave the server.
type extensionSummary struct {
	ExtensionID      string   `json:"extension_id"`
	Hostname         string   `json:"hostname"`
	BrowserInfo      string   `json:"browser_info"`
	ChromeVersion    string   `json:"chrome_version"`
	ExtensionVersion string   `json:"extension_version"`
	Capabilities     []string `json:"capabilities"`
	Status           string   `json:"status"`
	LastSeen         string   `json:"last_seen,omitempty"`
	PairedAt         string   `json:"paired_at"`
}

// toExtensionSummary converts a model row into the API shape.
//
// Written as an explicit mapping rather than serialising the model directly so
// TokenHash can never be included by accident if the model gains fields.
func toExtensionSummary(e models.Extension) extensionSummary {
	caps := []string(e.Capabilities)
	if caps == nil {
		caps = []string{}
	}
	lastSeen := ""
	if e.LastSeen != nil {
		lastSeen = e.LastSeen.UTC().Format("2006-01-02T15:04:05Z")
	}
	return extensionSummary{
		ExtensionID:      e.ExtensionID,
		Hostname:         e.Hostname,
		BrowserInfo:      e.BrowserInfo,
		ChromeVersion:    e.ChromeVersion,
		ExtensionVersion: e.ExtensionVersion,
		Capabilities:     caps,
		Status:           e.Status,
		LastSeen:         lastSeen,
		PairedAt:         e.PairedAt.UTC().Format("2006-01-02T15:04:05Z"),
	}
}
