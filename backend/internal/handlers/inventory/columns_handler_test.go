package inventory

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestColumnsHandler_GetAvailableColumns(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewColumnsHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/columns/available", nil)

		handler.GetAvailableColumns(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "tenant_id")
	})
}

func TestColumnsHandler_GetSelectedColumns(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewColumnsHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/inventory/columns/selected", nil)

		handler.GetSelectedColumns(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "tenant_id")
	})
}

func TestColumnsHandler_SaveSelectedColumns(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewColumnsHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/columns/selected", strings.NewReader(`{"columns": ["SKU", "Name"]}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.SaveSelectedColumns(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "tenant_id")
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := NewColumnsHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/columns/selected", strings.NewReader(`invalid`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.SaveSelectedColumns(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewColumnsHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewColumnsHandler(nil)
		assert.NotNil(t, handler)
	})
}
