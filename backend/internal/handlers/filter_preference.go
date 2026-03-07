package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// FilterPreferenceHandler handles filter preference endpoints
type FilterPreferenceHandler struct {
	fallbackDB *gorm.DB
}

// NewFilterPreferenceHandler creates a new filter preference handler
func NewFilterPreferenceHandler(db *gorm.DB) *FilterPreferenceHandler {
	return &FilterPreferenceHandler{fallbackDB: db}
}

// getDB returns the appropriate database for the current request
func (h *FilterPreferenceHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// ValidPlatforms defines valid platforms for filter preferences
var validPlatforms = []string{"tiktok", "lazada", "shopee", "inventory"}

func isValidPlatform(platform string) bool {
	for _, p := range validPlatforms {
		if p == platform {
			return true
		}
	}
	return false
}

// Get handles GET /api/filter-preferences
// Query params: platform (required), tab/page (optional, default "product")
func (h *FilterPreferenceHandler) Get(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	platform := c.Query("platform")
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid platform. Must be one of: tiktok, lazada, shopee, inventory",
		})
		return
	}

	// Support both 'page' (new) and 'tab' (legacy) parameters
	tab := c.DefaultQuery("page", c.DefaultQuery("tab", "product"))

	// Parse pagination
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)
	if limit > 100 {
		limit = 100
	}

	ctx := c.Request.Context()
	var pref models.FilterPreference
	err = db.WithContext(ctx).Where("tenant_id = ? AND platform = ? AND tab = ?", tenantID, platform, tab).
		First(&pref).Error

	if err == gorm.ErrRecordNotFound {
		// Return empty preference structure (wrapped in "data" to match Node.js format)
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"platform":        platform,
				"tab":             tab,
				"visible_columns": []string{},
				"column_filters":  map[string]interface{}{},
				"search_query":    "",
				"locked_columns":  []string{},
				"items":           []interface{}{},
			},
			"pagination": gin.H{
				"total":   0,
				"limit":   limit,
				"offset":  offset,
				"hasMore": false,
			},
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Parse JSON fields
	var visibleColumns []string
	var columnFilters map[string]interface{}
	var lockedColumns []string

	if pref.VisibleColumns != "" {
		json.Unmarshal([]byte(pref.VisibleColumns), &visibleColumns)
	}
	if pref.ColumnFilters != "" {
		json.Unmarshal([]byte(pref.ColumnFilters), &columnFilters)
	}
	if pref.LockedColumns != "" {
		json.Unmarshal([]byte(pref.LockedColumns), &lockedColumns)
	}

	// Set cache headers (no cache for filter prefs - they change frequently)
	c.Header("Cache-Control", "no-cache, must-revalidate")
	c.Header("ETag", "\""+platform+"_"+tab+"\"")
	c.Header("Pragma", "no-cache")

	// Return with "data" wrapper to match Node.js format
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"platform":        platform,
			"tab":             tab,
			"visible_columns": visibleColumns,
			"column_filters":  columnFilters,
			"search_query":    pref.SearchQuery,
			"locked_columns":  lockedColumns,
			"items":           []interface{}{},
		},
		"pagination": gin.H{
			"total":   0,
			"limit":   limit,
			"offset":  offset,
			"hasMore": false,
		},
	})
}

// SaveFilterPreferenceRequest represents the request body for saving filter preferences
type SaveFilterPreferenceRequest struct {
	Platform       string                 `json:"platform" binding:"required"`
	Page           string                 `json:"page"` // new parameter
	Tab            string                 `json:"tab"`  // legacy parameter
	VisibleColumns []string               `json:"visible_columns"`
	ColumnFilters  map[string]interface{} `json:"column_filters"`
	Filters        map[string]interface{} `json:"filters"` // legacy parameter
	SearchQuery    string                 `json:"search_query"`
	LockedColumns  []string               `json:"locked_columns"`
}

// Save handles POST /api/filter-preferences
func (h *FilterPreferenceHandler) Save(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var req SaveFilterPreferenceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if !isValidPlatform(req.Platform) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid platform. Must be one of: tiktok, lazada, shopee, inventory",
		})
		return
	}

	// Support both 'page' (new) and 'tab' (legacy) parameters
	tab := req.Page
	if tab == "" {
		tab = req.Tab
	}
	if tab == "" {
		tab = "product"
	}

	// Support both 'columnFilters' (new) and 'filters' (legacy) parameters
	columnFilters := req.ColumnFilters
	if columnFilters == nil {
		columnFilters = req.Filters
	}
	if columnFilters == nil {
		columnFilters = map[string]interface{}{}
	}

	// Serialize JSON fields
	visibleColumnsJSON, _ := json.Marshal(req.VisibleColumns)
	columnFiltersJSON, _ := json.Marshal(columnFilters)
	lockedColumnsJSON, _ := json.Marshal(req.LockedColumns)

	ctx := c.Request.Context()
	// Check if exists
	var existing models.FilterPreference
	findErr := db.WithContext(ctx).Where("tenant_id = ? AND platform = ? AND tab = ?", tenantID, req.Platform, tab).
		First(&existing).Error

	if findErr == gorm.ErrRecordNotFound {
		// Create new
		pref := models.FilterPreference{
			ID:             uuid.New().String(),
			TenantID:       tenantID,
			Platform:       req.Platform,
			Tab:            tab,
			VisibleColumns: string(visibleColumnsJSON),
			ColumnFilters:  string(columnFiltersJSON),
			SearchQuery:    req.SearchQuery,
			LockedColumns:  string(lockedColumnsJSON),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := db.WithContext(ctx).Create(&pref).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"success": true, "message": "Filter preference saved"})
		return
	}

	if findErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": findErr.Error()})
		return
	}

	// Update existing
	existing.VisibleColumns = string(visibleColumnsJSON)
	existing.ColumnFilters = string(columnFiltersJSON)
	existing.SearchQuery = req.SearchQuery
	existing.LockedColumns = string(lockedColumnsJSON)
	existing.UpdatedAt = time.Now()

	if err := db.WithContext(ctx).Save(&existing).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Filter preference updated"})
}

// Delete handles DELETE /api/filter-preferences
// Query params: platform (required), tab/page (optional)
func (h *FilterPreferenceHandler) Delete(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenantId"})
		return
	}

	db, dbErr := h.getDB(c)
	if dbErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": dbErr.Error()})
		return
	}

	platform := c.Query("platform")
	if !isValidPlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid platform. Must be one of: tiktok, lazada, shopee, inventory",
		})
		return
	}

	// Support both 'page' (new) and 'tab' (legacy) parameters
	tab := c.DefaultQuery("page", c.DefaultQuery("tab", "product"))

	ctx := c.Request.Context()
	result := db.WithContext(ctx).Where("tenant_id = ? AND platform = ? AND tab = ?", tenantID, platform, tab).
		Delete(&models.FilterPreference{})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": result.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Filter preference deleted"})
}
