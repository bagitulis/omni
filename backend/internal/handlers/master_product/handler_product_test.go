package master_product

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Delete(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/master-products/1", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Delete(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodDelete, "/api/master-products/1", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Delete(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestHandler_UpdateWithBody(t *testing.T) {
	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/master-products/1",
			strings.NewReader(`{"name": "Updated Product"}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Update(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
