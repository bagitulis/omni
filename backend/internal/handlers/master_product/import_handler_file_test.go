package master_product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestImportHandler_DownloadTemplate(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/import/template", nil)

		handler.DownloadTemplate(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/import/template?format=xlsx", nil)
		c.Set("tenantID", "test-tenant")

		handler.DownloadTemplate(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("valid tenant with csv format but no DB returns 500", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/master-products/import/template?format=csv", nil)
		c.Set("tenantID", "test-tenant")

		handler.DownloadTemplate(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
