package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/sku"
	"gorm.io/gorm"
)

// SKUCheckHandler handles SKU check endpoints
type SKUCheckHandler struct {
	db *gorm.DB
}

// NewSKUCheckHandler creates a new SKU check handler
func NewSKUCheckHandler(db *gorm.DB) *SKUCheckHandler {
	return &SKUCheckHandler{db: db}
}

// CheckSingle handles GET /api/sku/check/:sku
func (h *SKUCheckHandler) CheckSingle(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	skuCode := c.Param("sku")
	if skuCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU required"})
		return
	}

	// Build API map from context
	apis := h.buildAPIsFromContext(c)
	if len(apis) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no platform APIs configured"})
		return
	}

	service := sku.NewCheckService(h.db, tenantID)
	result := service.CheckSingle(c.Request.Context(), skuCode, apis)

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// CheckBatch handles POST /api/sku/batch-check
func (h *SKUCheckHandler) CheckBatch(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	var req struct {
		SKUs []string `json:"skus" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(req.SKUs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "max 100 SKUs per request"})
		return
	}

	apis := h.buildAPIsFromContext(c)
	if len(apis) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no platform APIs configured"})
		return
	}

	service := sku.NewCheckService(h.db, tenantID)
	results := service.CheckBatch(c.Request.Context(), req.SKUs, apis)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    results,
		"count":   len(results),
	})
}

// GetCachedStatus handles GET /api/sku/status/:sku
func (h *SKUCheckHandler) GetCachedStatus(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing tenantId"})
		return
	}

	skuCode := c.Param("sku")
	if skuCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "SKU required"})
		return
	}

	service := sku.NewCheckService(h.db, tenantID)
	cached := service.GetCachedStatus(skuCode)
	if cached == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not in cache"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": cached, "cached": true})
}

// buildAPIsFromContext extracts platform APIs from context
func (h *SKUCheckHandler) buildAPIsFromContext(c *gin.Context) map[string]sku.PlatformSKUChecker {
	apis := make(map[string]sku.PlatformSKUChecker)

	if api, ok := c.Get("shopeeAPI"); ok {
		if checker, ok := api.(sku.PlatformSKUChecker); ok {
			apis["shopee"] = checker
		}
	}
	if api, ok := c.Get("lazadaAPI"); ok {
		if checker, ok := api.(sku.PlatformSKUChecker); ok {
			apis["lazada"] = checker
		}
	}
	if api, ok := c.Get("tiktokAPI"); ok {
		if checker, ok := api.(sku.PlatformSKUChecker); ok {
			apis["tiktok"] = checker
		}
	}

	return apis
}
