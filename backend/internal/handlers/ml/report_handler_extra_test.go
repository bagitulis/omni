package ml

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestReportHandler_GetLatest(t *testing.T) {
	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports/shopee/latest", nil)
		c.Params = gin.Params{{Key: "platform", Value: "shopee"}}

		handler.GetLatest(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("missing platform returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports//latest", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "platform", Value: ""}}

		handler.GetLatest(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "platform")
	})

	t.Run("with valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports/shopee/latest", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "platform", Value: "shopee"}}

		handler.GetLatest(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestReportHandler_GetByFilename(t *testing.T) {
	t.Run("missing tenant ID returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports/shopee/report.html", nil)
		c.Params = gin.Params{
			{Key: "platform", Value: "shopee"},
			{Key: "filename", Value: "report.html"},
		}

		handler.GetByFilename(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("missing platform returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports//report.html", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{
			{Key: "platform", Value: ""},
			{Key: "filename", Value: "report.html"},
		}

		handler.GetByFilename(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "platform")
	})

	t.Run("missing filename returns 400", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports/shopee/", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{
			{Key: "platform", Value: "shopee"},
			{Key: "filename", Value: ""},
		}

		handler.GetByFilename(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Contains(t, resp["error"], "platform")
	})

	t.Run("with valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewReportHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/api/ml/reports/shopee/report.html", nil)
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{
			{Key: "platform", Value: "shopee"},
			{Key: "filename", Value: "report.html"},
		}

		handler.GetByFilename(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}
