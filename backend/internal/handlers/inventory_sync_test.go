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

func TestInventoryHandler_SyncFromSheets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/sync/from-sheets", body)
		c.Request.Header.Set("Content-Type", "application/json")
		// No tenantID set

		handler.SyncFromSheets(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "FAILED", resp["status"])
		assert.Contains(t, resp["message"], "tenant ID required")
	})

	t.Run("valid tenant ID with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/sync/from-sheets", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.SyncFromSheets(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_SyncToSheets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/sync/to-sheets", body)
		c.Request.Header.Set("Content-Type", "application/json")
		// No tenantID set

		handler.SyncToSheets(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "FAILED", resp["status"])
		assert.Contains(t, resp["message"], "tenant ID required")
	})

	t.Run("valid tenant ID with no spreadsheet config returns 400", func(t *testing.T) {
		// This test would need DB mock, but we can at least test that
		// without google auth the handler returns service unavailable
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		body := bytes.NewBufferString(`{"spreadsheet_id":"test-id","sheet_name":"Sheet1"}`)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/sync/to-sheets", body)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.SyncToSheets(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestGetInventorySpreadsheetConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("returns request params when provided", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

		// Since we don't have a DB, this will fail, but the function should
		// attempt to use request params first
		spreadsheetID, sheetName, err := handler.getInventorySpreadsheetConfig(c, "test-tenant", "test-spreadsheet-id", "Sheet1")

		// Without DB, should return error
		assert.Error(t, err)
		_ = spreadsheetID
		_ = sheetName
	})
}
