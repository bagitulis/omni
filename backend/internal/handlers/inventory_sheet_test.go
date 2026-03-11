package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestInventoryHandler_GetSyncStatus_MissingTenant tests GetSyncStatus without tenant
func TestInventoryHandler_GetSyncStatus_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/sync-status", handler.GetSyncStatus)

	req, _ := http.NewRequest("POST", "/api/inventory/sync-status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestInventoryHandler_GetSyncStatus_MissingSheetID tests GetSyncStatus without sheet_id
func TestInventoryHandler_GetSyncStatus_MissingSheetID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/sync-status", handler.GetSyncStatus)

	body := `{}`
	req, _ := http.NewRequest("POST", "/api/inventory/sync-status", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "sheetId")
}

// TestInventoryHandler_ExportToSheet_MissingTenant tests ExportToSheet without tenant
func TestInventoryHandler_ExportToSheet_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/export-to-sheet", handler.ExportToSheet)

	req, _ := http.NewRequest("POST", "/api/inventory/export-to-sheet", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestInventoryHandler_ExportToSheet_MissingData tests ExportToSheet without required data
func TestInventoryHandler_ExportToSheet_MissingData(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/export-to-sheet", handler.ExportToSheet)

	body := `{"sheet_id": "sheet-123"}`
	req, _ := http.NewRequest("POST", "/api/inventory/export-to-sheet", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestInventoryHandler_ExportToSheet_NoGoogleAuth tests ExportToSheet without Google auth
func TestInventoryHandler_ExportToSheet_NoGoogleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	// Handler without googleAuth configured
	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/export-to-sheet", handler.ExportToSheet)

	body := `{"sheet_id": "sheet-123", "data": [{"col1": "val1"}]}`
	req, _ := http.NewRequest("POST", "/api/inventory/export-to-sheet", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "Google Sheets")
}

// TestInventoryHandler_ImportFromSheet_MissingTenant tests ImportFromSheet without tenant
func TestInventoryHandler_ImportFromSheet_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/import-from-sheet", handler.ImportFromSheet)

	req, _ := http.NewRequest("POST", "/api/inventory/import-from-sheet", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestInventoryHandler_ImportFromSheet_MissingSheetID tests ImportFromSheet without sheet_id
func TestInventoryHandler_ImportFromSheet_MissingSheetID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/import-from-sheet", handler.ImportFromSheet)

	body := `{}`
	req, _ := http.NewRequest("POST", "/api/inventory/import-from-sheet", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestInventoryHandler_ImportFromSheet_NoGoogleAuth tests ImportFromSheet without Google auth
func TestInventoryHandler_ImportFromSheet_NoGoogleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewInventoryHandler(nil)
	r.POST("/api/inventory/import-from-sheet", handler.ImportFromSheet)

	body := `{"sheet_id": "sheet-123"}`
	req, _ := http.NewRequest("POST", "/api/inventory/import-from-sheet", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

// TestSyncStatusRequest_Structure tests SyncStatusRequest struct
func TestSyncStatusRequest_Structure(t *testing.T) {
	req := SyncStatusRequest{
		SheetID: "test-sheet-id",
	}
	assert.Equal(t, "test-sheet-id", req.SheetID)
}

// TestExportToSheetRequest_Structure tests ExportToSheetRequest struct
func TestExportToSheetRequest_Structure(t *testing.T) {
	req := ExportToSheetRequest{
		SheetID: "test-sheet-id",
		Data:    []map[string]interface{}{{"col1": "val1"}},
	}
	assert.Equal(t, "test-sheet-id", req.SheetID)
	assert.Len(t, req.Data, 1)
}

// TestImportFromSheetRequest_Structure tests ImportFromSheetRequest struct
func TestImportFromSheetRequest_Structure(t *testing.T) {
	req := ImportFromSheetRequest{
		SheetID:   "test-sheet-id",
		SheetName: "Sheet1",
	}
	assert.Equal(t, "test-sheet-id", req.SheetID)
	assert.Equal(t, "Sheet1", req.SheetName)
}
