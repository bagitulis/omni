package inventory

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSheetHandler_ExportToSheet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewSheetHandler(nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/export-to-sheet", strings.NewReader(`{"spreadsheet_id": "abc"}`))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.ExportToSheet(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
		assert.Contains(t, resp["error"], "tenant_id")
	})

	t.Run("invalid JSON returns 400", func(t *testing.T) {
		handler := NewSheetHandler(nil, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/inventory/export-to-sheet", strings.NewReader(`invalid`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenant_id", "test-tenant")

		handler.ExportToSheet(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestNewSheetHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewSheetHandler(nil, nil)
		assert.NotNil(t, handler)
	})
}
