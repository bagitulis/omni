package master_product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestHandler_List(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products", nil)

		handler.List(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant ID with no DB returns error", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products?page=1&limit=20", nil)
		c.Set("tenant_id", "test-tenant")

		handler.List(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_GetByID(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid ID returns 500 when DB unavailable", func(t *testing.T) {
		// Handler gets DB before validating ID, so returns 500 (DB error) not 400
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/abc", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "abc"}}

		handler.GetByID(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestHandler_Update(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/master-products/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("invalid ID returns 500 when DB unavailable", func(t *testing.T) {
		// Handler gets DB before validating ID, so returns 500 (DB error) not 400
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/master-products/invalid", nil)
		c.Set("tenant_id", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.Update(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_UpdateSku(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/master-products/skus/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.UpdateSku(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestNewHandler(t *testing.T) {
	t.Run("creates handler with base path", func(t *testing.T) {
		handler := NewHandler("/test/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/test/path", handler.basePath)
	})
}
