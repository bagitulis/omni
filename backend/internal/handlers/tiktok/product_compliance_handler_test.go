package tiktok

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestProductComplianceHandler_Constructor(t *testing.T) {
	handler := NewProductComplianceHandler("/data/path")
	assert.NotNil(t, handler)
	assert.Equal(t, "/data/path", handler.basePath)
}

func TestProductComplianceHandler_GetCompliance_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.GET("/api/tiktok/products/:productId/compliance", handler.GetCompliance)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/PROD123/compliance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductComplianceHandler_GetCompliance_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/:productId/compliance", handler.GetCompliance)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/PROD123/compliance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data, ok := resp["data"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "PROD123", data["product_id"])
	assert.Equal(t, "compliant", data["status"])
}

func TestProductComplianceHandler_UpdateCompliance_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.POST("/api/tiktok/products/:productId/compliance", handler.UpdateCompliance)

	body := `{"action": "fix"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/PROD123/compliance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductComplianceHandler_UpdateCompliance_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/:productId/compliance", handler.UpdateCompliance)

	body := `{invalid json}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/PROD123/compliance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductComplianceHandler_UpdateCompliance_MissingActionField(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/:productId/compliance", handler.UpdateCompliance)

	// action is required binding field
	body := `{}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/PROD123/compliance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductComplianceHandler_UpdateCompliance_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/:productId/compliance", handler.UpdateCompliance)

	body := `{"action": "fix_violations"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/PROD123/compliance", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
}

func TestProductComplianceHandler_GetGlobalProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.GET("/api/tiktok/products/global-products", handler.GetGlobalProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/global-products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductComplianceHandler_GetGlobalProducts_WithTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.GET("/api/tiktok/products/global-products", handler.GetGlobalProducts)

	req, _ := http.NewRequest("GET", "/api/tiktok/products/global-products", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// DB not available in test environment
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductComplianceHandler_PublishGlobal_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.POST("/api/tiktok/products/publish-global", handler.PublishGlobal)

	body := `{"product_id": "PROD123", "markets": ["ID"]}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/publish-global", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductComplianceHandler_PublishGlobal_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/publish-global", handler.PublishGlobal)

	// Missing markets (required)
	body := `{"product_id": "PROD123"}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/publish-global", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestProductComplianceHandler_PublishGlobal_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	handler := NewProductComplianceHandler("/test/path")

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	r.POST("/api/tiktok/products/publish-global", handler.PublishGlobal)

	body := `{"product_id": "PROD123", "markets": ["ID", "MY"]}`
	req, _ := http.NewRequest("POST", "/api/tiktok/products/publish-global", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
}

func TestComplianceStatus_JSONFields(t *testing.T) {
	status := ComplianceStatus{
		ProductID:         "PROD001",
		Status:            "compliant",
		Violations:        []ComplianceViolation{},
		LastCheckedAt:     1700000000,
		RecommendedAction: "none",
	}

	data, err := json.Marshal(status)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "PROD001", decoded["product_id"])
	assert.Equal(t, "compliant", decoded["status"])
	assert.Equal(t, float64(1700000000), decoded["last_checked_at"])
	assert.Equal(t, "none", decoded["recommended_action"])
}

func TestComplianceViolation_JSONFields(t *testing.T) {
	v := ComplianceViolation{
		Code:        "PROHIBITED_CONTENT",
		Description: "Product contains prohibited content",
		Field:       "title",
		Severity:    "high",
	}

	data, err := json.Marshal(v)
	assert.NoError(t, err)

	var decoded map[string]interface{}
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)

	assert.Equal(t, "PROHIBITED_CONTENT", decoded["code"])
	assert.Equal(t, "Product contains prohibited content", decoded["description"])
	assert.Equal(t, "title", decoded["field"])
	assert.Equal(t, "high", decoded["severity"])
}

func TestCurrentTimestamp(t *testing.T) {
	ts := currentTimestamp()
	assert.Greater(t, ts, int64(0))
}
