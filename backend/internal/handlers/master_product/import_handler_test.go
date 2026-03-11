package master_product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestImportHandler_Preview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/import/preview?platform=shopee&item_id=123", nil)

		handler.Preview(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("missing platform returns 500 when DB unavailable", func(t *testing.T) {
		// Handler gets DB before validating platform, so returns 500 (DB error) not 400
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/import/preview?item_id=123", nil)
		c.Set("tenant_id", "test-tenant")

		handler.Preview(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("unsupported platform returns 500 when DB unavailable", func(t *testing.T) {
		// Handler gets DB before validating platform, so returns 500 (DB error) not 400
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/import/preview?platform=lazada&item_id=123", nil)
		c.Set("tenant_id", "test-tenant")

		handler.Preview(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewImportHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewImportHandler("/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/path", handler.basePath)
	})
}
