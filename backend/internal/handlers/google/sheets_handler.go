// Package google handles Google API related endpoints
package google

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
)

// SheetsHandler handles Google Sheets data operations
type SheetsHandler struct {
	authService *google.AuthService
}

// NewSheetsHandler creates a new sheets handler
func NewSheetsHandler(authService *google.AuthService) *SheetsHandler {
	return &SheetsHandler{authService: authService}
}

// ListSpreadsheets handles GET /api/google/sheets/list
// Returns list of available spreadsheets
func (h *SheetsHandler) ListSpreadsheets(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	spreadsheets, err := sheetsService.ListSpreadsheets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spreadsheets,
	})
}

// GetSpreadsheetData handles GET /api/google/sheets/data
// Fetches all inventory data from configured Google Sheet
func (h *SheetsHandler) GetSpreadsheetData(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	spreadsheetID := c.Query("spreadsheet_id")
	sheetName := c.Query("sheet_name")
	startRow := c.DefaultQuery("start_row", "2")

	if spreadsheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "spreadsheet_id is required",
		})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)

	var data [][]interface{}
	var err error

	if sheetName != "" {
		rangeStr := sheetName + "!A" + startRow + ":ZZ"
		data, err = sheetsService.ReadRange(c.Request.Context(), spreadsheetID, rangeStr)
	} else {
		// Get first sheet
		info, infoErr := sheetsService.GetSpreadsheetInfo(c.Request.Context(), spreadsheetID)
		if infoErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   infoErr.Error(),
			})
			return
		}
		if len(info.Sheets) > 0 {
			rangeStr := info.Sheets[0].Title + "!A" + startRow + ":ZZ"
			data, err = sheetsService.ReadRange(c.Request.Context(), spreadsheetID, rangeStr)
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
		"count":   len(data),
	})
}

// GetWorksheets handles GET /api/google/sheets/worksheets/:spreadsheetId
// Returns list of worksheets in a spreadsheet
func (h *SheetsHandler) GetWorksheets(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	spreadsheetID := c.Param("spreadsheetId")
	if spreadsheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "spreadsheetId is required",
		})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	info, err := sheetsService.GetSpreadsheetInfo(c.Request.Context(), spreadsheetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    info.Sheets,
	})
}

// GetColumnHeaders handles GET /api/google/sheets/columns/:spreadsheetId/:sheetName
// Returns column headers from a specific sheet
func (h *SheetsHandler) GetColumnHeaders(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	spreadsheetID := c.Param("spreadsheetId")
	sheetName := c.Param("sheetName")

	if spreadsheetID == "" || sheetName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "spreadsheetId and sheetName are required",
		})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	rangeStr := sheetName + "!1:1"
	data, err := sheetsService.ReadRange(c.Request.Context(), spreadsheetID, rangeStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	headers := make([]string, 0)
	if len(data) > 0 {
		for _, h := range data[0] {
			if str, ok := h.(string); ok {
				headers = append(headers, str)
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    headers,
	})
}

// CreateSpreadsheet handles POST /api/google/sheets/create
// Creates a new Google Spreadsheet
func (h *SheetsHandler) CreateSpreadsheet(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Missing tenantId",
		})
		return
	}

	var req struct {
		Title string `json:"title" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	spreadsheet, err := sheetsService.CreateSpreadsheet(c.Request.Context(), req.Title)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spreadsheet,
	})
}

// RefreshSpreadsheets handles GET /api/google/sheets/refresh
// Refreshes list of available spreadsheets
func (h *SheetsHandler) RefreshSpreadsheets(c *gin.Context) {
	// Same as ListSpreadsheets but forces cache refresh
	h.ListSpreadsheets(c)
}

// extractSpreadsheetID extracts spreadsheet ID from URL or returns as-is
func extractSpreadsheetID(input string) string {
	// If already an ID (no slashes), return as-is
	if !regexp.MustCompile(`/`).MatchString(input) {
		return input
	}

	// Extract from Google Sheets URL
	// Format: https://docs.google.com/spreadsheets/d/{SPREADSHEET_ID}/edit...
	re := regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9-_]+)`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1]
	}

	return input
}
