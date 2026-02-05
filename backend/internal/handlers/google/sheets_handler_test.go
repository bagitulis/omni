package google

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSheetsHandler_ListSpreadsheets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewSheetsHandler(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/google/sheets/list", nil)

		handler.ListSpreadsheets(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "tenantId")
	})
}

func TestNewSheetsHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewSheetsHandler(nil)
		assert.NotNil(t, handler)
	})
}
