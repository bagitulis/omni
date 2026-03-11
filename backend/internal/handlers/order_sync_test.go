package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOrderSyncHandler_SyncByCategory_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/orders/sync/pending", nil)
	c.Params = gin.Params{{Key: "category", Value: "pending"}}

	h := NewOrderSyncHandler()
	h.SyncByCategory(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestOrderSyncHandler_SyncByCategory_MissingCategory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/orders/sync/", nil)
	c.Set("tenant_id", "test-tenant")

	h := NewOrderSyncHandler()
	h.SyncByCategory(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

func TestOrderSyncHandler_SyncPlatformOrders_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/orders/sync/platform/shopee", nil)
	c.Params = gin.Params{{Key: "platform", Value: "shopee"}}

	h := NewOrderSyncHandler()
	h.SyncPlatformOrders(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestOrderSyncHandler_GetOrdersByCategory_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/orders/pending", nil)
	c.Params = gin.Params{{Key: "category", Value: "pending"}}

	h := NewOrderSyncHandler()
	h.GetOrdersByCategory(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestOrderSyncHandler_GetOrderDetails_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/orders/details/shopee", nil)
	c.Params = gin.Params{{Key: "platform", Value: "shopee"}}

	h := NewOrderSyncHandler()
	h.GetOrderDetails(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestOrderSyncHandler_GetOrderDetails_MissingParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/orders/details/shopee", nil)
	c.Set("tenant_id", "test-tenant")

	h := NewOrderSyncHandler()
	h.GetOrderDetails(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}
