package inventory

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestDataHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewDataHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/data", nil)

		handler.List(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "tenantId")
	})
}

func TestDataHandler_GetBySKU(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewDataHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/data/SKU001", nil)
		c.Params = gin.Params{{Key: "sku", Value: "SKU001"}}

		handler.GetBySKU(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "tenantId")
	})

	t.Run("missing SKU returns 400", func(t *testing.T) {
		handler := NewDataHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/data/", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "sku", Value: ""}}

		handler.GetBySKU(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "SKU")
	})
}

func TestNewDataHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewDataHandler(nil)
		assert.NotNil(t, handler)
	})
}
