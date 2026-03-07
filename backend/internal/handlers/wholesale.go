package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto/response"
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
// Returns settings in admin_fee-based format matching frontend expectations
func (h *WholesaleHandler) GetSettings(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	settings, err := service.GetSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	// Use response.Success() for consistent "data" wrapper
	// Frontend extractSettingsPayload() looks for response.data first
	c.JSON(http.StatusOK, response.Success(settings))
}

// UpdateSettings handles PUT /api/wholesale/settings
// Accepts: { admin_fee, min_order_1, max_order_1, max_order_tier_3 }
func (h *WholesaleHandler) UpdateSettings(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	var settings models.WholesaleSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	// Validate
	if settings.MinOrder1 < 2 {
		c.JSON(http.StatusBadRequest, response.Error("min_order_1 must be at least 2"))
		return
	}
	if settings.MaxOrder1 <= settings.MinOrder1 {
		c.JSON(http.StatusBadRequest, response.Error("max_order_1 must be greater than min_order_1"))
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	if err := service.UpdateSettings(c.Request.Context(), &settings); err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"message": "settings updated",
	}))
}

// Calculate handles POST /api/wholesale/calculate
// Accepts: { base_price: number }
func (h *WholesaleHandler) Calculate(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	var req models.WholesaleCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	service := wholesale.NewWholesaleService(db, tenantID)
	result, err := service.CalculateTiers(c.Request.Context(), req.BasePrice)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(result))
}

// Apply handles POST /api/wholesale/apply
func (h *WholesaleHandler) Apply(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	var req models.WholesaleApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	// Get API from context
	apiKey := req.Platform + "API"
	api, exists := c.Get(apiKey)
	if !exists {
		c.JSON(http.StatusBadRequest, response.Error(req.Platform+" API not configured"))
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

	c.JSON(http.StatusOK, response.Success(gin.H{
		"total":         len(results),
		"success_count": successCount,
		"results":       results,
	}))
}
