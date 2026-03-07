// Package google handles Google API related endpoints
package google

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
	"gorm.io/gorm"
)

// SheetConfigHandler handles sheet configuration endpoints
type SheetConfigHandler struct {
	db          *gorm.DB
	authService *google.AuthService
}

// NewSheetConfigHandler creates a new sheet config handler
func NewSheetConfigHandler(db *gorm.DB, authService *google.AuthService) *SheetConfigHandler {
	return &SheetConfigHandler{db: db, authService: authService}
}

// SheetConfig represents a sheet configuration
type SheetConfig struct {
	SheetID      string            `json:"sheet_id"`
	Type         string            `json:"type"` // inventory, wallet, shipping, order
	Mappings     map[string]string `json:"mappings"`
	HeaderRow    int               `json:"header_row"`
	DataStartRow int               `json:"data_start_row"`
}

// SaveConfig handles POST /api/google/sheet-config/save
// Saves sheet configuration
func (h *SheetConfigHandler) SaveConfig(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req SheetConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	configService := google.NewSheetConfigService(h.db, tenantID)
	if err := configService.SaveConfig(c.Request.Context(), &google.SheetConfigInput{
		SheetID:      req.SheetID,
		Type:         req.Type,
		Mappings:     req.Mappings,
		HeaderRow:    req.HeaderRow,
		DataStartRow: req.DataStartRow,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration saved successfully",
	})
}

// ListConfigs handles GET /api/google/sheet-config/list
// Returns all sheet configurations for tenant
func (h *SheetConfigHandler) ListConfigs(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	configService := google.NewSheetConfigService(h.db, tenantID)
	configs, err := configService.ListConfigs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    configs,
	})
}

// DeleteConfig handles DELETE /api/google/sheet-config/:sheetId
// Deletes a sheet configuration
func (h *SheetConfigHandler) DeleteConfig(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	sheetID := c.Param("sheetId")
	if sheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "sheetId is required",
		})
		return
	}

	configService := google.NewSheetConfigService(h.db, tenantID)
	if err := configService.DeleteConfig(c.Request.Context(), sheetID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Configuration deleted successfully",
	})
}
