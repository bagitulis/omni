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

func TestInventoryHandler_GetConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/config", nil)
		// No tenantID set

		handler.GetConfig(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenantId")
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/config", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetConfig(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_GetList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/list", nil)
		// No tenantID set

		handler.GetList(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenantId")
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/list", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetList(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_GetRecordByKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/SKU001", nil)
		c.Params = gin.Params{{Key: "keyValue", Value: "SKU001"}}
		// No tenantID set

		handler.GetRecordByKey(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenantId")
	})

	t.Run("missing keyValue returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "keyValue", Value: ""}}

		handler.GetRecordByKey(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "keyValue is required")
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/SKU001", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "keyValue", Value: "SKU001"}}

		handler.GetRecordByKey(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_UpdateRecordByKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"quantity": 100}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/inventory/SKU001", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = gin.Params{{Key: "keyValue", Value: "SKU001"}}
		// No tenantID set

		handler.UpdateRecordByKey(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenantId")
	})

	t.Run("missing keyValue returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"quantity": 100}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/inventory/", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "keyValue", Value: ""}}

		handler.UpdateRecordByKey(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "keyValue is required")
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/inventory/SKU001", bytes.NewReader([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "keyValue", Value: "SKU001"}}

		handler.UpdateRecordByKey(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Invalid request body")
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"quantity": 100}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/inventory/SKU001", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "keyValue", Value: "SKU001"}}

		handler.UpdateRecordByKey(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_CreateRecord(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"keyValue": "SKU001", "quantity": 100}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		// No tenantID set

		handler.CreateRecord(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenantId")
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory", bytes.NewReader([]byte("invalid json")))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.CreateRecord(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Invalid request body")
	})

	t.Run("missing keyValue returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"quantity": 100}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.CreateRecord(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "keyValue is required")
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		body := map[string]interface{}{"keyValue": "SKU001", "quantity": 100}
		bodyBytes, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory", bytes.NewReader(bodyBytes))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.CreateRecord(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestInventoryHandler_DeleteRecordByKey(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/inventory/SKU001", nil)
		c.Params = gin.Params{{Key: "keyValue", Value: "SKU001"}}
		// No tenantID set

		handler.DeleteRecordByKey(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "Missing tenantId")
	})

	t.Run("missing keyValue returns 400", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/inventory/", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "keyValue", Value: ""}}

		handler.DeleteRecordByKey(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "keyValue is required")
	})

	t.Run("valid request with no DB returns error", func(t *testing.T) {
		handler := NewInventoryHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/inventory/SKU001", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "keyValue", Value: "SKU001"}}

		handler.DeleteRecordByKey(c)

		// Without a valid DB, should return error
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

// Test helper functions
func TestParseCommaSeparatedColumns(t *testing.T) {
	t.Run("empty string returns empty slice", func(t *testing.T) {
		result := parseCommaSeparatedColumns("")
		assert.Equal(t, []string{}, result)
	})

	t.Run("single column", func(t *testing.T) {
		result := parseCommaSeparatedColumns("SKU")
		assert.Equal(t, []string{"SKU"}, result)
	})

	t.Run("multiple columns", func(t *testing.T) {
		result := parseCommaSeparatedColumns("SKU, Name, Price")
		assert.Equal(t, []string{"SKU", "Name", "Price"}, result)
	})

	t.Run("columns with extra spaces", func(t *testing.T) {
		result := parseCommaSeparatedColumns("  SKU  ,  Name  ,  Price  ")
		assert.Equal(t, []string{"SKU", "Name", "Price"}, result)
	})

	t.Run("empty segments are ignored", func(t *testing.T) {
		result := parseCommaSeparatedColumns("SKU,,Name,,,Price")
		assert.Equal(t, []string{"SKU", "Name", "Price"}, result)
	})
}

func TestExtractSpreadsheetIDFromURL(t *testing.T) {
	t.Run("empty string returns empty", func(t *testing.T) {
		result := extractSpreadsheetIDFromURL("")
		assert.Equal(t, "", result)
	})

	t.Run("plain ID returns as-is", func(t *testing.T) {
		result := extractSpreadsheetIDFromURL("1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms")
		assert.Equal(t, "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms", result)
	})

	t.Run("full Google Sheets URL extracts ID", func(t *testing.T) {
		url := "https://docs.google.com/spreadsheets/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit#gid=0"
		result := extractSpreadsheetIDFromURL(url)
		assert.Equal(t, "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms", result)
	})

	t.Run("URL with different path extracts ID", func(t *testing.T) {
		url := "https://docs.google.com/spreadsheets/d/abc123_XYZ-def/view"
		result := extractSpreadsheetIDFromURL(url)
		assert.Equal(t, "abc123_XYZ-def", result)
	})
}
