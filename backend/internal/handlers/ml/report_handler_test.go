package ml

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestReportHandler_Generate(t *testing.T) {
	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/ml/reports/generate", strings.NewReader("platform=shopee"))
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		handler.Generate(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("missing platform returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/ml/reports/generate", nil)
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		c.Set("tenantID", "test-tenant")

		handler.Generate(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "platform")
	})

	t.Run("invalid platform returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/api/ml/reports/generate", strings.NewReader("platform=invalid"))
		c.Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		c.Request.PostForm = map[string][]string{"platform": {"invalid"}}
		c.Set("tenantID", "test-tenant")

		handler.Generate(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "platform")
	})
}

func TestReportHandler_List(t *testing.T) {
	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports/shopee/list", nil)
		c.Params = gin.Params{{Key: "platform", Value: "shopee"}}

		handler.List(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("missing platform returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports//list", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "platform", Value: ""}}

		handler.List(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "platform")
	})
}

func TestNewReportHandler(t *testing.T) {
	t.Run("creates handler", func(t *testing.T) {
		handler := NewReportHandler("/path")
		assert.NotNil(t, handler)
		assert.Equal(t, "/path", handler.basePath)
	})
}
