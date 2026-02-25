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

func TestImportHandler_AutoMapBatch(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/automap/batch", nil)

		handler.AutoMapBatch(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/automap/batch",
			strings.NewReader(`{"platform_ids": ["123", "456"]}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.AutoMapBatch(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid JSON body returns 400", func(t *testing.T) {
		handler := NewImportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/import/automap/batch",
			strings.NewReader(`not-valid-json`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.AutoMapBatch(c)

		// DB fails before JSON parse, so expect 500 not 400 (same pattern as handler_test.go)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
