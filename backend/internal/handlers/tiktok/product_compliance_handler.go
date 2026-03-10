package tiktok

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
)

// ComplianceHandler handles TikTok product compliance endpoints.
// All endpoints return 501 — TikTok Compliance API SDK not yet implemented.
type ComplianceHandler struct {
	basePath string
}

// NewComplianceHandler creates a new compliance handler
func NewComplianceHandler(basePath string) *ComplianceHandler {
	return &ComplianceHandler{basePath: basePath}
}

// GetCompliance handles GET /api/tiktok/products/compliance/:productId
func (h *ComplianceHandler) GetCompliance(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Compliance API not yet implemented"))
}

// UpdateCompliance handles PUT /api/tiktok/products/compliance/:productId
func (h *ComplianceHandler) UpdateCompliance(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Compliance API not yet implemented"))
}

// PublishGlobal handles POST /api/tiktok/products/compliance/:productId/publish-global
func (h *ComplianceHandler) PublishGlobal(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Global Publishing API not yet implemented"))
}

// GetGlobalProducts handles GET /api/tiktok/products/global-products
func (h *ComplianceHandler) GetGlobalProducts(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, response.Error("TikTok Global Products API not yet implemented"))
}
