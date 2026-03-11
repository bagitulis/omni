package shopee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	"github.com/stretchr/testify/assert"
)

func TestShippingFeeHandler_ProcessShippingFee_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)

	r.POST("/api/shopee/shipping/process-fee", handler.ProcessShippingFee)

	body := `{"order_sn_list": ["ORDER123"]}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/process-fee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShippingFeeHandler_ProcessShippingFee_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/shipping/process-fee", handler.ProcessShippingFee)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/process-fee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShippingFeeHandler_ProcessShippingFee_EmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/shipping/process-fee", handler.ProcessShippingFee)

	body := `{"order_sn_list": []}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/process-fee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShippingFeeHandler_ExportShippingFee_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)

	r.POST("/api/shopee/shipping/export-fee", handler.ExportShippingFee)

	body := `{"month": 1, "year": 2024}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/export-fee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShippingFeeHandler_ExportShippingFee_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/shipping/export-fee", handler.ExportShippingFee)

	// Missing year field
	body := `{"month": 1}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/export-fee", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShippingFeeHandler_ExportToSheets_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)

	r.POST("/api/shopee/shipping/export-to-sheets", handler.ExportToSheets)

	body := `{"month": 1, "year": 2024, "spreadsheet_id": "SHEET123", "sheet_name": "Sheet1"}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/export-to-sheets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShippingFeeHandler_ExportToSheets_NoGoogleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	// Pass nil for google auth service
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/shipping/export-to-sheets", handler.ExportToSheets)

	body := `{"month": 1, "year": 2024, "spreadsheet_id": "SHEET123", "sheet_name": "Sheet1"}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/export-to-sheets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShippingFeeHandler_ExportToSheets_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/shipping/export-to-sheets", handler.ExportToSheets)

	// Missing spreadsheet_id and sheet_name
	body := `{"month": 1, "year": 2024}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/export-to-sheets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestNewShippingFeeHandler(t *testing.T) {
	mockClient := new(MockShopeeAPIClient)
	handler := NewShippingFeeHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.getAPIClient)
	assert.Nil(t, handler.googleAuthService)
}
