package master_product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestStagingImportHandler_ImportFromShopee(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewStagingImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/staging/shopee", nil)

		handler.ImportFromShopee(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewStagingImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/staging/shopee", nil)
		c.Set("tenant_id", "test-tenant")

		handler.ImportFromShopee(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestStagingImportHandler_ImportFromTiktok(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewStagingImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/staging/tiktok", nil)

		handler.ImportFromTiktok(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewStagingImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/staging/tiktok", nil)
		c.Set("tenant_id", "test-tenant")

		handler.ImportFromTiktok(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestStagingImportHandler_ImportFromLazada(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewStagingImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/staging/lazada", nil)

		handler.ImportFromLazada(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewStagingImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/staging/lazada", nil)
		c.Set("tenant_id", "test-tenant")

		handler.ImportFromLazada(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNewStagingImportHandler(t *testing.T) {
	t.Run("creates handler with base path", func(t *testing.T) {
		handler := NewStagingImportHandler("/test/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/test/path", handler.basePath)
	})

	t.Run("creates handler with empty base path", func(t *testing.T) {
		handler := NewStagingImportHandler("")
		assert.NotNil(t, handler)
	})
}
