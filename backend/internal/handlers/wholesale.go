package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/wholesale"
	"gorm.io/gorm"
)

// WholesaleHandler handles wholesale endpoints
type WholesaleHandler struct {
	fallbackDB *gorm.DB
}

// NewWholesaleHandler creates a new wholesale handler
func NewWholesaleHandler(db *gorm.DB) *WholesaleHandler {
	return &WholesaleHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
func (h *WholesaleHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// GetSettings handles GET /api/wholesale/settings
func (h *WholesaleHandler) GetSettings(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	settings, err := service.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "settings": settings})
}

// UpdateSettings handles PUT /api/wholesale/settings
func (h *WholesaleHandler) UpdateSettings(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var settings models.WholesaleSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	if err := service.UpdateSettings(c.Request.Context(), &settings); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "settings updated"})
}

// Calculate handles POST /api/wholesale/calculate
func (h *WholesaleHandler) Calculate(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req models.WholesaleCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	result, err := service.CalculateTiers(c.Request.Context(), req.OriginalPrice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
}

// Apply handles POST /api/wholesale/apply
func (h *WholesaleHandler) Apply(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req models.WholesaleApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get API from context
	apiKey := req.Platform + "API"
	api, exists := c.Get(apiKey)
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": req.Platform + " API not configured"})
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	results := service.ApplyToProducts(c.Request.Context(), api.(wholesale.PlatformWholesaleAPI), req.Platform, req.SKUs)

	// Count successes
	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       successCount == len(results),
		"total":         len(results),
		"success_count": successCount,
		"results":       results,
	})
}
