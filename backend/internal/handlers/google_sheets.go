package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/google"
)

// GoogleSheetsHandler handles Google Sheets endpoints
type GoogleSheetsHandler struct {
	authService *google.AuthService
}

// NewGoogleSheetsHandler creates a new handler
func NewGoogleSheetsHandler(authService *google.AuthService) *GoogleSheetsHandler {
	return &GoogleSheetsHandler{authService: authService}
}

// GetSpreadsheetInfo handles GET /api/google-sheets/spreadsheets/:id
func (h *GoogleSheetsHandler) GetSpreadsheetInfo(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	spreadsheetID := c.Param("id")
	if spreadsheetID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "spreadsheet ID required"})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	info, err := sheetsService.GetSpreadsheetInfo(c.Request.Context(), spreadsheetID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": info})
}

// ReadData handles POST /api/google-sheets/read
func (h *GoogleSheetsHandler) ReadData(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	var req struct {
		SpreadsheetID string `json:"spreadsheet_id" binding:"required"`
		Range         string `json:"range" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	data, err := sheetsService.ReadRange(c.Request.Context(), req.SpreadsheetID, req.Range)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// WriteData handles POST /api/google-sheets/write
func (h *GoogleSheetsHandler) WriteData(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	var req struct {
		SpreadsheetID string          `json:"spreadsheet_id" binding:"required"`
		Range         string          `json:"range" binding:"required"`
		Values        [][]interface{} `json:"values" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	if err := sheetsService.WriteRange(c.Request.Context(), req.SpreadsheetID, req.Range, req.Values); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "data written"})
}

// DetectColumns handles POST /api/google-sheets/detect-columns
func (h *GoogleSheetsHandler) DetectColumns(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	var req struct {
		SpreadsheetID string `json:"spreadsheet_id" binding:"required"`
		Range         string `json:"range" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	dataOps := google.NewDataOperations(sheetsService)
	detected, err := dataOps.AutoDetectColumns(c.Request.Context(), req.SpreadsheetID, req.Range)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": detected})
}

// ImportData handles POST /api/google-sheets/import
func (h *GoogleSheetsHandler) ImportData(c *gin.Context) {
	tenantID := c.GetString("tenantID")
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "tenant ID required"})
		return
	}

	var req struct {
		SpreadsheetID string         `json:"spreadsheet_id" binding:"required"`
		Range         string         `json:"range" binding:"required"`
		Mappings      map[int]string `json:"mappings" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sheetsService := google.NewSheetsService(h.authService, tenantID)
	dataOps := google.NewDataOperations(sheetsService)
	data, err := dataOps.ImportData(c.Request.Context(), req.SpreadsheetID, req.Range, req.Mappings)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data, "count": len(data)})
}
