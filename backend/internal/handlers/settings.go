package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

// SettingsHandler handles general settings endpoints
type SettingsHandler struct {
	fallbackDB *gorm.DB
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{fallbackDB: db}
}

// generateID generates a unique ID using xid
func generateID() string {
	return xid.New().String()
}

// getDB returns the appropriate database for the current request
func (h *SettingsHandler) getDB(c *gin.Context) (*gorm.DB, error) {
	return GetTenantDBFromContext(c, h.fallbackDB)
}

// GetInventorySettings handles GET /api/settings/inventory
func (h *SettingsHandler) GetInventorySettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	var settings models.InventorySettings
	err = db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		settings = models.InventorySettings{TenantID: tenantID}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "settings": settings})
}

// UpdateInventorySettings handles PUT /api/settings/inventory
func (h *SettingsHandler) UpdateInventorySettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var req InventorySettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	var settings models.InventorySettings
	err = db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		settings = models.InventorySettings{TenantID: tenantID}
	}

	settings.SpreadsheetID = req.SpreadsheetID
	settings.SheetName = req.SheetName
	settings.AutoSync = req.AutoSync
	settings.SyncIntervalSec = req.SyncInterval

	if err := db.WithContext(ctx).Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "settings": settings})
}

// GetGoogleSheetsSettings handles GET /api/settings/google-sheets
func (h *SettingsHandler) GetGoogleSheetsSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	var settings models.GoogleSheetsSettings
	// Schema-level multi-tenancy: just get first record (only one per schema)
	err = db.WithContext(ctx).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		// Return empty settings object
		settings = models.GoogleSheetsSettings{}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "settings": settings})
}

// UpdateGoogleSheetsSettings handles PUT /api/settings/google-sheets
func (h *SettingsHandler) UpdateGoogleSheetsSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var req GoogleSheetsSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	var settings models.GoogleSheetsSettings
	// Schema-level multi-tenancy: just get first record (only one per schema)
	err = db.WithContext(ctx).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		// Create new with generated ID
		settings = models.GoogleSheetsSettings{
			ID: generateID(),
		}
	}

	settings.SpreadsheetID = req.DefaultSpreadsheetID

	if err := db.WithContext(ctx).Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "settings": settings})
}

// GetGeneralSettings handles GET /api/settings/general
func (h *SettingsHandler) GetGeneralSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	var settings models.GeneralSettings
	err = db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		// Return default settings
		settings = models.GeneralSettings{
			TenantID:             tenantID,
			Language:             "en",
			Timezone:             "Asia/Jakarta",
			NotificationsEmail:   true,
			NotificationsBrowser: true,
			AutoSync:             true,
			SyncInterval:         "30",
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

// UpdateGeneralSettings handles POST /api/settings/general
func (h *SettingsHandler) UpdateGeneralSettings(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}

	db, err := h.getDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	var req GeneralSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	var settings models.GeneralSettings
	err = db.WithContext(ctx).Where("tenant_id = ?", tenantID).First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		settings = models.GeneralSettings{
			ID:       generateID(),
			TenantID: tenantID,
		}
	}

	settings.Language = req.Language
	settings.Timezone = req.Timezone
	settings.NotificationsEmail = req.NotificationsEmail
	settings.NotificationsBrowser = req.NotificationsBrowser
	settings.AutoSync = req.AutoSync
	settings.SyncInterval = req.SyncInterval

	if err := db.WithContext(ctx).Save(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": settings})
}

// InventorySettingsRequest represents inventory settings request
type InventorySettingsRequest struct {
	SpreadsheetID string `json:"spreadsheet_id"`
	SheetName     string `json:"sheet_name"`
	AutoSync      bool   `json:"auto_sync"`
	SyncInterval  int    `json:"sync_interval"`
}

// GoogleSheetsSettingsRequest represents Google Sheets settings request
type GoogleSheetsSettingsRequest struct {
	ServiceAccountEmail  string `json:"service_account_email"`
	DefaultSpreadsheetID string `json:"default_spreadsheet_id"`
	IsConnected          bool   `json:"is_connected"`
}

// GeneralSettingsRequest represents general settings request
type GeneralSettingsRequest struct {
	Language             string `json:"language"`
	Timezone             string `json:"timezone"`
	NotificationsEmail   bool   `json:"notifications_email"`
	NotificationsBrowser bool   `json:"notifications_browser"`
	AutoSync             bool   `json:"auto_sync"`
	SyncInterval         string `json:"sync_interval"`
}
