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

func TestWalletReportHandler_GetWalletReport_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	r.POST("/wallet/report", handler.GetWalletReport)

	body := `{"month": 1, "year": 2024}`
	req, _ := http.NewRequest("POST", "/wallet/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletReportHandler_GetWalletReport_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	r.POST("/wallet/report", handler.GetWalletReport)

	// Missing month and year
	body := `{}`
	req, _ := http.NewRequest("POST", "/wallet/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletReportHandler_GetWalletReport_InvalidMonth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	r.POST("/wallet/report", handler.GetWalletReport)

	// month=0 is invalid (binding:"required,min=1,max=12")
	body := `{"month": 0, "year": 2024}`
	req, _ := http.NewRequest("POST", "/wallet/report", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletReportHandler_ExportWallet_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	r.POST("/wallet/export", handler.ExportWallet)

	body := `{"month": 3, "year": 2024}`
	req, _ := http.NewRequest("POST", "/wallet/export", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletReportHandler_ExportWallet_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	r.POST("/wallet/export", handler.ExportWallet)

	// Missing year (binding:"required,min=2020")
	body := `{"month": 3}`
	req, _ := http.NewRequest("POST", "/wallet/export", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletReportHandler_ExportToSheets_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	r.POST("/wallet/export-to-sheets", handler.ExportToSheets)

	body := `{"month": 3, "year": 2024, "spreadsheet_id": "abc", "sheet_name": "Sheet1"}`
	req, _ := http.NewRequest("POST", "/wallet/export-to-sheets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletReportHandler_ExportToSheets_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	r.POST("/wallet/export-to-sheets", handler.ExportToSheets)

	// Missing spreadsheet_id and sheet_name (both binding:"required")
	body := `{"month": 3, "year": 2024}`
	req, _ := http.NewRequest("POST", "/wallet/export-to-sheets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletReportHandler_ExportToSheets_NilGoogleAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	// nil google auth — handler returns 503 ServiceUnavailable
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	r.POST("/wallet/export-to-sheets", handler.ExportToSheets)

	body := `{"month": 3, "year": 2024, "spreadsheet_id": "abc123", "sheet_name": "Sheet1"}`
	req, _ := http.NewRequest("POST", "/wallet/export-to-sheets", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestNewWalletReportHandler_NotNil(t *testing.T) {
	mockClient := new(MockShopeeAPIClient)
	handler := NewWalletReportHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	}, nil)
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.getAPIClient)
	assert.Nil(t, handler.googleAuthService)
}
