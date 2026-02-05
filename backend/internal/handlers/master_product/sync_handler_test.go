package master_product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSyncHandler_Sync(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewSyncHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/1/sync", nil)
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.Sync(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("invalid ID returns 500 when DB unavailable", func(t *testing.T) {
		// Handler gets DB before validating ID, so returns 500 (DB error) not 400
		handler := NewSyncHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/master-products/invalid/sync", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "invalid"}}

		handler.Sync(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewSyncHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewSyncHandler("/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/path", handler.basePath)
	})
}
