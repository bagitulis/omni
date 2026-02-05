package google

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestQuotaHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("NewQuotaHandler creates handler", func(t *testing.T) {
		handler := NewQuotaHandler(nil)
		assert.NotNil(t, handler)
	})
}
