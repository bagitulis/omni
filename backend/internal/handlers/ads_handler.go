package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/services/ads"
	"gorm.io/gorm"
)

// AdsHandler handles ads report endpoints
type AdsHandler struct {
	fallbackDB *gorm.DB
}

// NewAdsHandler creates a new ads handler
func NewAdsHandler(db *gorm.DB) *AdsHandler {
	return &AdsHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
func (h *AdsHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// parseDateRange extracts date range from query params
func parseDateRange(c *gin.Context) (time.Time, time.Time) {
	// Default to last 30 days if not provided
	end := time.Now()
	start := end.AddDate(0, 0, -30)

	if s := c.Query("startDate"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			start = t
		}
	}
	if s := c.Query("endDate"); s != "" {
		if t, err := time.Parse("2006-01-02", s); err == nil {
			end = t
		}
	}
	return start, end
}

// validateTenantID validates tenant ID from context
func validateTenantID(c *gin.Context) (string, bool) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return "", false
	}
	return tenantID, true
}

// newShopeeService creates a Shopee ads service
func (h *AdsHandler) newShopeeService(c *gin.Context) (*ads.ShopeeAdsService, string, error) {
	tenantID, ok := validateTenantID(c)
	if !ok {
		return nil, "", nil
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return nil, "", err
	}

	return ads.NewShopeeAdsService(db, tenantID), tenantID, nil
}

// newTiktokService creates a TikTok ads service
func (h *AdsHandler) newTiktokService(c *gin.Context) (*ads.TiktokAdsService, string, error) {
	tenantID, ok := validateTenantID(c)
	if !ok {
		return nil, "", nil
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return nil, "", err
	}

	return ads.NewTiktokAdsService(db, tenantID), tenantID, nil
}
