// Package google handles Google API related endpoints
package google

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/handlers"
	"github.com/omni/backend/internal/services/google"
	"gorm.io/gorm"
)

// SettingsHandler handles Google Sheets settings endpoints
type SettingsHandler struct {
	authService *google.AuthService
	db          *gorm.DB
}

// NewSettingsHandler creates a new settings handler
func NewSettingsHandler(authService *google.AuthService, db *gorm.DB) *SettingsHandler {
	return &SettingsHandler{authService: authService, db: db}
}

// getTenantDB returns the proper tenant DB with schema set
func (h *SettingsHandler) getTenantDB(c *gin.Context) (*gorm.DB, error) {
	return handlers.GetTenantDBFromContext(c, h.db)
}

// DetailedSettings represents the Google Sheets settings structure
type DetailedSettings struct {
	WalletSpreadsheetID      string   `json:"wallet_spreadsheet_id"`
	ShippingSpreadsheetID    string   `json:"shipping_spreadsheet_id"`
	InventorySpreadsheetID   string   `json:"inventory_spreadsheet_id"`
	OrderSpreadsheetID       string   `json:"order_spreadsheet_id"`
	InventorySheetName       string   `json:"inventory_sheet_name"`
	WalletSheetName          string   `json:"wallet_sheet_name"`
	ShippingSheetName        string   `json:"shipping_sheet_name"`
	OrderSheetName           string   `json:"order_sheet_name"`
	InventorySelectedColumns []string `json:"inventory_selected_columns"`
}

// GetDetailedSettings handles GET /api/google/settings/detailed
// Returns current detailed Google Sheets settings
func (h *SettingsHandler) GetDetailedSettings(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	db, err := h.getTenantDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	settingsService := google.NewSettingsService(db, tenantID)
	settings, err := settingsService.GetDetailedSettings(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

// UpdateDetailedSettings handles POST /api/google/settings/update-detailed
// Updates detailed Google Sheets settings
func (h *SettingsHandler) UpdateDetailedSettings(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	var req DetailedSettings
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	db, err := h.getTenantDB(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	settingsService := google.NewSettingsService(db, tenantID)
	if err := settingsService.UpdateDetailedSettings(c.Request.Context(), &google.DetailedSettingsInput{
		WalletSpreadsheetID:      req.WalletSpreadsheetID,
		ShippingSpreadsheetID:    req.ShippingSpreadsheetID,
		InventorySpreadsheetID:   req.InventorySpreadsheetID,
		OrderSpreadsheetID:       req.OrderSpreadsheetID,
		InventorySheetName:       req.InventorySheetName,
		WalletSheetName:          req.WalletSheetName,
		ShippingSheetName:        req.ShippingSheetName,
		OrderSheetName:           req.OrderSheetName,
		InventorySelectedColumns: req.InventorySelectedColumns,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Settings updated successfully",
	})
}

// TestConnection handles GET /api/google/settings/test
// Tests Google Sheets connection
func (h *SettingsHandler) TestConnection(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	spreadsheetID := c.Query("spreadsheet_id")

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	err := sheetsService.TestConnection(c.Request.Context(), spreadsheetID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"connected": false,
			"error":     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"connected": true,
	})
}
