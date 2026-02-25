package tiktok

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Constructor test ---

func TestNewProductSearchHandler(t *testing.T) {
	h := NewProductSearchHandler("/some/base/path")
	assert.NotNil(t, h)
	assert.Equal(t, "/some/base/path", h.basePath)
}

// --- SearchProducts handler tests ---

func TestProductSearchHandler_SearchProducts_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductSearchHandler("/test/path")
	r.POST("/products/search", h.SearchProducts)

	req, _ := http.NewRequest(http.MethodPost, "/products/search", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.NotEmpty(t, resp["error"])
}

func TestProductSearchHandler_SearchProducts_WithTenant_ReturnsServerError(t *testing.T) {
	// In test environment there is no real DB/TikTok credentials.
	// Handler should reach getTiktokClient → config.GetTenantDB → fail with 500.
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})
	h := NewProductSearchHandler("/test/path")
	r.POST("/products/search", h.SearchProducts)

	req, _ := http.NewRequest(http.MethodPost, "/products/search", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Must NOT be 401 (tenant was provided)
	assert.NotEqual(t, http.StatusUnauthorized, w.Code)
	// In test env DB is unavailable → 500
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.NotEmpty(t, resp["error"])
}

func TestProductSearchHandler_SearchProducts_ResponseIsJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewProductSearchHandler("/test/path")
	r.POST("/products/search", h.SearchProducts)

	req, _ := http.NewRequest(http.MethodPost, "/products/search", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Response must always be valid JSON regardless of error
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	_, hasSucess := resp["success"]
	assert.True(t, hasSucess, "response must have a 'success' field")
}
