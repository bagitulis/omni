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

func TestPriceHandler_UpdatePrice(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-price", strings.NewReader(`{"sku": "SKU001"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.UpdatePrice(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "tenantId")
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := NewPriceHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/update-price", strings.NewReader(`invalid`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.UpdatePrice(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewPriceHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewPriceHandler(nil)
		assert.NotNil(t, handler)
	})
}
