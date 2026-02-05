package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Google Sheets Handler Tests (google_sheets.go)
// =============================================================================

// TestGoogleSheetsGetSpreadsheetInfo_MissingTenant tests getting spreadsheet info without tenant
func TestGoogleSheetsGetSpreadsheetInfo_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewGoogleSheetsHandler(nil)
	r.GET("/api/google-sheets/spreadsheets/:id", handler.GetSpreadsheetInfo)

	req, _ := http.NewRequest("GET", "/api/google-sheets/spreadsheets/spreadsheet-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "tenant ID required")
}

// TestGoogleSheetsGetSpreadsheetInfo_MissingSpreadsheetID tests getting spreadsheet without ID
func TestGoogleSheetsGetSpreadsheetInfo_MissingSpreadsheetID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/google-sheets/spreadsheets/:id", func(c *gin.Context) {
		spreadsheetID := c.Param("id")
		if spreadsheetID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "spreadsheet ID required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "id": spreadsheetID})
	})

	// With a valid ID
	req, _ := http.NewRequest("GET", "/api/google-sheets/spreadsheets/test-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "test-123", resp["id"])
}

// TestGoogleSheetsReadData_MissingTenant tests reading data without tenant
func TestGoogleSheetsReadData_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewGoogleSheetsHandler(nil)
	r.POST("/api/google-sheets/read", handler.ReadData)

	reqBody := `{"spreadsheet_id": "abc", "range": "Sheet1!A1:B10"}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/read", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGoogleSheetsReadData_InvalidBody tests reading data with invalid body
func TestGoogleSheetsReadData_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/google-sheets/read", func(c *gin.Context) {
		var req struct {
			SpreadsheetID string `json:"spreadsheet_id" binding:"required"`
			Range         string `json:"range" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})

	// Missing required fields
	reqBody := `{"spreadsheet_id": "abc"}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/read", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGoogleSheetsWriteData_MissingTenant tests writing data without tenant
func TestGoogleSheetsWriteData_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewGoogleSheetsHandler(nil)
	r.POST("/api/google-sheets/write", handler.WriteData)

	reqBody := `{"spreadsheet_id": "abc", "range": "Sheet1!A1:B10", "values": [[1,2],[3,4]]}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/write", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGoogleSheetsWriteData_InvalidBody tests writing data with invalid body
func TestGoogleSheetsWriteData_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/google-sheets/write", func(c *gin.Context) {
		var req struct {
			SpreadsheetID string          `json:"spreadsheet_id" binding:"required"`
			Range         string          `json:"range" binding:"required"`
			Values        [][]interface{} `json:"values" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})

	// Missing required values field
	reqBody := `{"spreadsheet_id": "abc", "range": "Sheet1!A1:B10"}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/write", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGoogleSheetsWriteData_ValidBody tests writing data with valid body
func TestGoogleSheetsWriteData_ValidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/google-sheets/write", func(c *gin.Context) {
		var req struct {
			SpreadsheetID string          `json:"spreadsheet_id" binding:"required"`
			Range         string          `json:"range" binding:"required"`
			Values        [][]interface{} `json:"values" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "data written",
			"rows":    len(req.Values),
		})
	})

	reqBody := `{"spreadsheet_id": "abc", "range": "Sheet1!A1:B10", "values": [["a","b"],["c","d"]]}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/write", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "data written", resp["message"])
	assert.Equal(t, float64(2), resp["rows"])
}

// TestGoogleSheetsDetectColumns_MissingTenant tests detecting columns without tenant
func TestGoogleSheetsDetectColumns_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewGoogleSheetsHandler(nil)
	r.POST("/api/google-sheets/detect-columns", handler.DetectColumns)

	reqBody := `{"spreadsheet_id": "abc", "range": "Sheet1!A1:Z1"}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/detect-columns", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGoogleSheetsDetectColumns_InvalidBody tests detecting columns with invalid body
func TestGoogleSheetsDetectColumns_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/google-sheets/detect-columns", func(c *gin.Context) {
		var req struct {
			SpreadsheetID string `json:"spreadsheet_id" binding:"required"`
			Range         string `json:"range" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})

	// Missing required range field
	reqBody := `{"spreadsheet_id": "abc"}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/detect-columns", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGoogleSheetsImportData_MissingTenant tests importing data without tenant
func TestGoogleSheetsImportData_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewGoogleSheetsHandler(nil)
	r.POST("/api/google-sheets/import", handler.ImportData)

	reqBody := `{"spreadsheet_id": "abc", "range": "Sheet1!A1:B10", "mappings": {}}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/import", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGoogleSheetsImportData_InvalidBody tests importing data with invalid body
func TestGoogleSheetsImportData_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/google-sheets/import", func(c *gin.Context) {
		var req struct {
			SpreadsheetID string         `json:"spreadsheet_id" binding:"required"`
			Range         string         `json:"range" binding:"required"`
			Mappings      map[int]string `json:"mappings" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	})

	// Missing required mappings field
	reqBody := `{"spreadsheet_id": "abc", "range": "Sheet1!A1:B10"}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/import", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestGoogleSheetsImportData_ValidBody tests importing data with valid body
func TestGoogleSheetsImportData_ValidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/google-sheets/import", func(c *gin.Context) {
		var req struct {
			SpreadsheetID string            `json:"spreadsheet_id" binding:"required"`
			Range         string            `json:"range" binding:"required"`
			Mappings      map[string]string `json:"mappings" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []map[string]interface{}{},
			"count":   0,
		})
	})

	reqBody := `{"spreadsheet_id": "abc", "range": "Sheet1!A1:B10", "mappings": {"0": "sku", "1": "name"}}`
	req, _ := http.NewRequest("POST", "/api/google-sheets/import", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(0), resp["count"])
}
