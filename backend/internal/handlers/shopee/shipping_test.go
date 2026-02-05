package shopee

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockShopeeAPIClient implements shopeeService.APIClient interface for testing
type MockShopeeAPIClient struct {
	mock.Mock
}

// Implement APIClient interface methods
func (m *MockShopeeAPIClient) GetWalletBalance(ctx context.Context) (map[string]interface{}, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockShopeeAPIClient) GetShippingOptions(ctx context.Context, orderSn string) (map[string]interface{}, error) {
	args := m.Called(ctx, orderSn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockShopeeAPIClient) GetTrackingInfo(ctx context.Context, orderSn string) (map[string]interface{}, error) {
	args := m.Called(ctx, orderSn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func (m *MockShopeeAPIClient) GetShipmentInfo(ctx context.Context, orderSn string) (map[string]interface{}, error) {
	args := m.Called(ctx, orderSn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]interface{}), args.Error(1)
}

func TestShopeeShippingHandler_GetOptions_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.GET("/api/shopee/shipping/options", handler.GetOptions)

	req, _ := http.NewRequest("GET", "/api/shopee/shipping/options?orderSn=ORDER123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeShippingHandler_GetOptions_MissingOrderSN(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/shipping/options", handler.GetOptions)

	req, _ := http.NewRequest("GET", "/api/shopee/shipping/options", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeShippingHandler_ArrangeShipment_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.POST("/api/shopee/shipping/arrange", handler.ArrangeShipment)

	body := `{"order_sn": "ORDER123"}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/arrange", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeShippingHandler_ArrangeShipment_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/shopee/shipping/arrange", handler.ArrangeShipment)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/shopee/shipping/arrange", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeShippingHandler_GetTracking_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.GET("/api/shopee/shipping/tracking/:orderSn", handler.GetTracking)

	req, _ := http.NewRequest("GET", "/api/shopee/shipping/tracking/ORDER123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeShippingHandler_GetTracking_EmptyOrderSN(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/shopee/shipping/tracking/:orderSn", handler.GetTracking)

	// Path without orderSn will not match this route - tests 404
	req, _ := http.NewRequest("GET", "/api/shopee/shipping/tracking/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Route requires :orderSn param, so empty will 404
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestShopeeShippingHandler_GetShipment_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.GET("/api/shopee/shipping/info/:orderSn", handler.GetShipment)

	req, _ := http.NewRequest("GET", "/api/shopee/shipping/info/ORDER123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeShippingHandler_GetShippingLabel_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.GET("/api/shopee/shipping/label/:orderSn", handler.GetShippingLabel)

	req, _ := http.NewRequest("GET", "/api/shopee/shipping/label/ORDER123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeShippingHandler_DownloadShippingLabel_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	mockClient := new(MockShopeeAPIClient)

	handler := NewShippingHandler(func(tenantID string) shopeeService.APIClient {
		return mockClient
	})

	r.GET("/api/shopee/shipping/download/:orderSn", handler.DownloadShippingLabel)

	req, _ := http.NewRequest("GET", "/api/shopee/shipping/download/ORDER123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}
