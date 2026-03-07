package google

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	googleService "github.com/omni/backend/internal/services/google"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthHandler_GetAuthStatus(t *testing.T) {
	t.Run("returns unauthenticated when no credentials", func(t *testing.T) {
		authService := googleService.NewAuthService(nil)
		handler := NewAuthHandler(authService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/google/auth/status", nil)

		handler.GetAuthStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["success"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, false, data["authenticated"])
	})

	t.Run("returns authenticated when credentials provided", func(t *testing.T) {
		authService := googleService.NewAuthService([]byte(`{"type":"service_account"}`))
		handler := NewAuthHandler(authService)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/google/auth/status", nil)

		handler.GetAuthStatus(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, true, resp["success"])
		data := resp["data"].(map[string]interface{})
		assert.Equal(t, true, data["authenticated"])
	})
}

func TestNewAuthHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewAuthHandler(nil)
		assert.NotNil(t, handler)
	})
}
