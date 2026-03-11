package analytics

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMLHandler_GetPortfolioHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns error", func(t *testing.T) {
		handler := NewMLHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/ml/portfolio-health?platform=tiktok", nil)

		handler.GetPortfolioHealth(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant ID with no service returns error", func(t *testing.T) {
		handler := NewMLHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/analytics/ml/portfolio-health?platform=tiktok", nil)
		c.Set("tenant_id", "test-tenant")

		handler.GetPortfolioHealth(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNewMLHandler(t *testing.T) {
	t.Run("creates handler with cache", func(t *testing.T) {
		handler := NewMLHandler(nil)
		assert.NotNil(t, handler)
	})
}
