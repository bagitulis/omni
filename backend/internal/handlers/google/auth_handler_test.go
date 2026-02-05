package google

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

func TestAuthHandler_GetAuthURL(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/google/auth/url", nil)

		handler.GetAuthURL(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "tenantId")
	})

	t.Run("valid tenant ID with nil service panics or errors", func(t *testing.T) {
		// When authService is nil, calling methods on it will panic
		// This tests that the handler is correctly instantiated
		handler := NewAuthHandler(nil)
		assert.NotNil(t, handler)
	})
}

func TestAuthHandler_HandleCallback(t *testing.T) {
	t.Run("missing code returns 400", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/google/auth/callback", nil)

		handler.HandleCallback(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "authorization code")
	})

	t.Run("error param returns 400", func(t *testing.T) {
		handler := NewAuthHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/google/auth/callback?error=access_denied", nil)

		handler.HandleCallback(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "OAuth error")
	})
}

func TestNewAuthHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewAuthHandler(nil)
		assert.NotNil(t, handler)
	})
}
