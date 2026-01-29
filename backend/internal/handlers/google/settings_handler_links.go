package google

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
)

// SpreadsheetLink represents a saved spreadsheet link
type SpreadsheetLink struct {
	Type          string `json:"type"` // wallet, shipping, inventory, order
	SpreadsheetID string `json:"spreadsheet_id"`
	URL           string `json:"url"`
	Title         string `json:"title"`
}

// SaveLinksRequest represents the request body from frontend
type SaveLinksRequest struct {
	Inventory *string `json:"inventory"`
	Wallet    *string `json:"wallet"`
	Shipping  *string `json:"shipping"`
	Order     *string `json:"order"`
}

// SaveLinks handles POST /api/google/settings/save-links
// Saves spreadsheet links for different purposes
func (h *SettingsHandler) SaveLinks(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	var req SaveLinksRequest
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

	// Convert to LinksByType for service
	links := &google.LinksByType{
		Inventory: getStringValue(req.Inventory),
		Wallet:    getStringValue(req.Wallet),
		Shipping:  getStringValue(req.Shipping),
		Order:     getStringValue(req.Order),
	}

	settingsService := google.NewSettingsService(db, tenantID)
	if err := settingsService.SaveSpreadsheetLinks(c.Request.Context(), links); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Links saved successfully",
	})
}

// getStringValue safely extracts string value from pointer
func getStringValue(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// GetSavedLinks handles GET /api/google/settings/saved-links
// Returns saved spreadsheet links
func (h *SettingsHandler) GetSavedLinks(c *gin.Context) {
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
	links, err := settingsService.GetSpreadsheetLinks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    links,
	})
}

// ValidateLink handles POST /api/google/settings/validate-link
// Validates spreadsheet link and detects sheets
// Compatible with Node.js backend format: expects {spreadsheetUrl, type}
func (h *SettingsHandler) ValidateLink(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId - authentication required",
		})
		return
	}

	var req struct {
		SpreadsheetURL string `json:"spreadsheet_url"`
		URL            string `json:"url"` // Fallback for backward compatibility
		Type           string `json:"type"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Use spreadsheetUrl if provided, otherwise fall back to url
	urlToValidate := req.SpreadsheetURL
	if urlToValidate == "" {
		urlToValidate = req.URL
	}

	if urlToValidate == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Spreadsheet URL is required",
		})
		return
	}

	// Validate type if provided
	validTypes := map[string]bool{"inventory": true, "wallet": true, "shipping": true, "order": true}
	if req.Type != "" && !validTypes[req.Type] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Valid spreadsheet type is required (inventory, wallet, shipping, order)",
		})
		return
	}

	// Extract spreadsheet ID from URL
	spreadsheetID := extractSpreadsheetID(urlToValidate)
	if spreadsheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid spreadsheet URL format",
		})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	info, err := sheetsService.GetSpreadsheetInfo(c.Request.Context(), spreadsheetID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid spreadsheet URL or no access. Make sure the service account has access to this spreadsheet.",
		})
		return
	}

	// Convert sheets to format expected by frontend
	sheets := make([]gin.H, len(info.Sheets))
	for i, sheet := range info.Sheets {
		sheets[i] = gin.H{
			"name":        sheet.Title,
			"sheetId":     sheet.ID,
			"index":       sheet.Index,
			"columnCount": sheet.Cols,
			"rowCount":    sheet.Rows,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"spreadsheetId": info.ID,
			"name":          info.Title,
			"sheets":        sheets,
			"type":          req.Type,
		},
	})
}
