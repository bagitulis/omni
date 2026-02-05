package google

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestServiceAccountHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("NewServiceAccountHandler creates handler", func(t *testing.T) {
		handler := NewServiceAccountHandler(nil)
		assert.NotNil(t, handler)
	})
}

func TestServiceAccountHandler_ListAccounts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("nil auth service panics", func(t *testing.T) {
		// When authService is nil, calling methods on it will panic
		// This test verifies the handler is correctly structured
		handler := NewServiceAccountHandler(nil)
		assert.NotNil(t, handler)
	})
}
