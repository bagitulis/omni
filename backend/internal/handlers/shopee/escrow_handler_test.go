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

func TestEscrowHandler_GetEscrowDetail_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewEscrowHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.POST("/api/shopee/wallet/escrow-detail", handler.GetEscrowDetail)

	body := `{"order_sn": "ORDER123"}`
	req, _ := http.NewRequest("POST", "/api/shopee/wallet/escrow-detail", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestEscrowHandler_GetEscrowDetail_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewEscrowHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/wallet/escrow-detail", handler.GetEscrowDetail)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/shopee/wallet/escrow-detail", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestEscrowHandler_GetEscrowDetail_MissingOrderSN(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewEscrowHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/wallet/escrow-detail", handler.GetEscrowDetail)

	body := `{}`
	req, _ := http.NewRequest("POST", "/api/shopee/wallet/escrow-detail", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestEscrowHandler_GetEscrowDetailBatch_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewEscrowHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.POST("/api/shopee/wallet/escrow-detail-batch", handler.GetEscrowDetailBatch)

	body := `{"order_sn_list": ["ORDER123"]}`
	req, _ := http.NewRequest("POST", "/api/shopee/wallet/escrow-detail-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestEscrowHandler_GetEscrowDetailBatch_EmptyList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewEscrowHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/wallet/escrow-detail-batch", handler.GetEscrowDetailBatch)

	body := `{"order_sn_list": []}`
	req, _ := http.NewRequest("POST", "/api/shopee/wallet/escrow-detail-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestEscrowHandler_GetEscrowDetailBatch_ExceedsMaxSize(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewEscrowHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/wallet/escrow-detail-batch", handler.GetEscrowDetailBatch)

	// Build a list of 51 order SNs (exceeds max of 50)
	orderList := make([]string, 51)
	for i := range orderList {
		orderList[i] = `"ORDER` + strings.Repeat("0", i) + `"`
	}
	body := `{"order_sn_list": [` + strings.Join(orderList, ",") + `]}`
	req, _ := http.NewRequest("POST", "/api/shopee/wallet/escrow-detail-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestEscrowHandler_GetEscrowDetailBatch_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)
	handler := NewEscrowHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/wallet/escrow-detail-batch", handler.GetEscrowDetailBatch)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/shopee/wallet/escrow-detail-batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestNewEscrowHandler(t *testing.T) {
	mockClient := new(MockShopeeAPIClient)
	handler := NewEscrowHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})
	assert.NotNil(t, handler)
	assert.NotNil(t, handler.getAPIClient)
}
