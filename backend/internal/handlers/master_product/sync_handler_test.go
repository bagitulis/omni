package master_product

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	masterProductService "github.com/omni/backend/internal/services/master_product"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockSyncService struct{}

func (m *mockSyncService) SyncToPlatform(_ context.Context, _ string, _ uint, _ string) (*masterProductService.SyncResult, error) {
	return &masterProductService.SyncResult{}, nil
}

func (m *mockSyncService) GetSyncStatus(_ context.Context, _ string, _ uint) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

type mockMasterProductImageService struct {
	refreshResult *masterProductService.RefreshProductImagesResult
	refreshErr    error
	lastTenantID  string
	lastProductID uint
	lastForce     bool
}

func (m *mockMasterProductImageService) BackfillImages(_ context.Context, _ string, _ int, _ bool) (*masterProductService.BackfillImagesResult, error) {
	return &masterProductService.BackfillImagesResult{}, nil
}

func (m *mockMasterProductImageService) RefreshProductImages(_ context.Context, tenantID string, productID uint, force bool) (*masterProductService.RefreshProductImagesResult, error) {
	m.lastTenantID = tenantID
	m.lastProductID = productID
	m.lastForce = force
	if m.refreshErr != nil {
		return nil, m.refreshErr
	}
	if m.refreshResult != nil {
		return m.refreshResult, nil
	}
	return &masterProductService.RefreshProductImagesResult{}, nil
}

func TestSyncHandler_Sync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewSyncHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/1/sync", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Sync(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("invalid ID returns 400", func(t *testing.T) {
		handler := NewSyncHandler("")
		handler.getTenantDB = func(_ string, _ string) (*gorm.DB, error) {
			return nil, nil
		}
		handler.newSyncService = func(_ *gorm.DB, _ string) syncService {
			return &mockSyncService{}
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/invalid/sync", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.Sync(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewSyncHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewSyncHandler("/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/path", handler.basePath)
	})
}

func TestSyncHandler_RefreshProductImages(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewSyncHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/1/images/refresh", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.RefreshProductImages(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("invalid product ID returns 400", func(t *testing.T) {
		handler := NewSyncHandler("")
		handler.getTenantDB = func(_ string, _ string) (*gorm.DB, error) {
			return nil, nil
		}
		handler.newMasterProductService = func(_ *gorm.DB) masterProductImageService {
			return &mockMasterProductImageService{}
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/invalid/images/refresh", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.RefreshProductImages(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("passes force payload and returns service response fields", func(t *testing.T) {
		mockService := &mockMasterProductImageService{
			refreshResult: &masterProductService.RefreshProductImagesResult{
				MasterProductID:    1,
				PreviousImageCount: 1,
				ImageCount:         3,
				Updated:            true,
				Forced:             true,
				Images:             []string{"/uploads/a.webp", "/uploads/b.webp", "/uploads/c.webp"},
			},
		}

		handler := NewSyncHandler("")
		handler.getTenantDB = func(_ string, _ string) (*gorm.DB, error) {
			return nil, nil
		}
		handler.newMasterProductService = func(_ *gorm.DB) masterProductImageService {
			return mockService
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(
			http.MethodPost,
			"/api/master-products/1/images/refresh",
			bytes.NewBufferString(`{"force":true}`),
		)
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.RefreshProductImages(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test-tenant", mockService.lastTenantID)
		assert.Equal(t, uint(1), mockService.lastProductID)
		assert.True(t, mockService.lastForce)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["success"])

		data, ok := resp["data"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, float64(1), data["master_product_id"])
		assert.Equal(t, float64(1), data["previous_image_count"])
		assert.Equal(t, float64(3), data["image_count"])
		assert.Equal(t, true, data["updated"])
		assert.Equal(t, true, data["forced"])
	})

	t.Run("empty request body defaults force to false", func(t *testing.T) {
		mockService := &mockMasterProductImageService{}

		handler := NewSyncHandler("")
		handler.getTenantDB = func(_ string, _ string) (*gorm.DB, error) {
			return nil, nil
		}
		handler.newMasterProductService = func(_ *gorm.DB) masterProductImageService {
			return mockService
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/2/images/refresh", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "2"}}

		handler.RefreshProductImages(c)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, uint(2), mockService.lastProductID)
		assert.False(t, mockService.lastForce)
	})

	t.Run("maps product-not-found error to 404", func(t *testing.T) {
		mockService := &mockMasterProductImageService{refreshErr: masterProductService.ErrProductNotFound}

		handler := NewSyncHandler("")
		handler.getTenantDB = func(_ string, _ string) (*gorm.DB, error) {
			return nil, nil
		}
		handler.newMasterProductService = func(_ *gorm.DB) masterProductImageService {
			return mockService
		}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/99/images/refresh", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "99"}}

		handler.RefreshProductImages(c)

		assert.Equal(t, http.StatusNotFound, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Equal(t, "Product not found", resp["error"])
	})
}
