package shopee

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestWalletHandler_GetBalance_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewWalletHandler(nil, "/tmp/test-wallet")
	r.GET("/wallet/balance", handler.GetBalance)

	req, _ := http.NewRequest("GET", "/wallet/balance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetBalance_WithTenantID_NoDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	handler := NewWalletHandler(nil, "/tmp/nonexistent-wallet-path")
	r.GET("/wallet/balance", handler.GetBalance)

	req, _ := http.NewRequest("GET", "/wallet/balance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Fails at service level (no credentials in DB)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetTransactions_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewWalletHandler(nil, "/tmp/test-wallet")
	r.GET("/wallet/transactions", handler.GetTransactions)

	req, _ := http.NewRequest("GET", "/wallet/transactions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetTransactions_InvalidStartDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	handler := NewWalletHandler(nil, "/tmp/test-wallet")
	r.GET("/wallet/transactions", handler.GetTransactions)

	req, _ := http.NewRequest("GET", "/wallet/transactions?start_date=not-a-date", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetTransactions_InvalidEndDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	handler := NewWalletHandler(nil, "/tmp/test-wallet")
	r.GET("/wallet/transactions", handler.GetTransactions)

	req, _ := http.NewRequest("GET", "/wallet/transactions?end_date=2024/01/15", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetTransactions_ValidDateFormat_NoDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	handler := NewWalletHandler(nil, "/tmp/nonexistent-wallet-path")
	r.GET("/wallet/transactions", handler.GetTransactions)

	req, _ := http.NewRequest("GET", "/wallet/transactions?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Valid date format passes — fails at service/DB level
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetNetIncome_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewWalletHandler(nil, "/tmp/test-wallet")
	r.GET("/wallet/income", handler.GetNetIncome)

	req, _ := http.NewRequest("GET", "/wallet/income", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetNetIncome_InvalidStartDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	handler := NewWalletHandler(nil, "/tmp/test-wallet")
	r.GET("/wallet/income", handler.GetNetIncome)

	req, _ := http.NewRequest("GET", "/wallet/income?start_date=baddate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetNetIncome_InvalidEndDate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	handler := NewWalletHandler(nil, "/tmp/test-wallet")
	r.GET("/wallet/income", handler.GetNetIncome)

	req, _ := http.NewRequest("GET", "/wallet/income?end_date=15-01-2024", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestWalletHandler_GetNetIncome_ValidDates_NoDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	handler := NewWalletHandler(nil, "/tmp/nonexistent-wallet-path")
	r.GET("/wallet/income", handler.GetNetIncome)

	req, _ := http.NewRequest("GET", "/wallet/income?start_date=2024-01-01&end_date=2024-01-31", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Dates parse OK — fails at service level
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestNewWalletHandler_NotNil(t *testing.T) {
	handler := NewWalletHandler(nil, "/tmp/test")
	assert.NotNil(t, handler)
}
