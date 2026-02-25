package master_product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestImportHandler_GetMappingStatus(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/1/mapping", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetMappingStatus(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/1/mapping", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetMappingStatus(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestImportHandler_AutoMap(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/mapping/auto",
			strings.NewReader(`{"seller_sku": "SKU-001"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.AutoMap(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/mapping/auto",
			strings.NewReader(`{"seller_sku": "SKU-001"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.AutoMap(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestImportHandler_ManualLink(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/mapping/link",
			strings.NewReader(`{"master_sku_id": 1, "platform": "shopee", "platform_item_id": "123"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ManualLink(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/mapping/link",
			strings.NewReader(`{"master_sku_id": 1, "platform": "shopee", "platform_item_id": "123"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.ManualLink(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestImportHandler_Unlink(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/mapping/unlink",
			strings.NewReader(`{"master_sku_id": 1, "platform": "shopee"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Unlink(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/mapping/unlink",
			strings.NewReader(`{"master_sku_id": 1, "platform": "shopee"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.Unlink(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
