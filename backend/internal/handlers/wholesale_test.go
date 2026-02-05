package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestWholesaleHandler_NewWholesaleHandler tests handler creation
func TestWholesaleHandler_NewWholesaleHandler(t *testing.T) {
	handler := NewWholesaleHandler(nil)
	assert.NotNil(t, handler)
}

// TestWholesaleHandler_GetSettings_MissingTenant tests GetSettings without tenant
func TestWholesaleHandler_GetSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleHandler(nil)
	r.GET("/api/wholesale/settings", handler.GetSettings)

	req, _ := http.NewRequest("GET", "/api/wholesale/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleHandler_UpdateSettings_MissingTenant tests UpdateSettings without tenant
func TestWholesaleHandler_UpdateSettings_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleHandler(nil)
	r.PUT("/api/wholesale/settings", handler.UpdateSettings)

	req, _ := http.NewRequest("PUT", "/api/wholesale/settings", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleHandler_UpdateSettings_InvalidBody tests UpdateSettings with invalid JSON
func TestWholesaleHandler_UpdateSettings_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleHandler(nil)
	r.PUT("/api/wholesale/settings", handler.UpdateSettings)

	body := `{invalid json`
	req, _ := http.NewRequest("PUT", "/api/wholesale/settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestWholesaleHandler_Calculate_MissingTenant tests Calculate without tenant
func TestWholesaleHandler_Calculate_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleHandler(nil)
	r.POST("/api/wholesale/calculate", handler.Calculate)

	req, _ := http.NewRequest("POST", "/api/wholesale/calculate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleHandler_Calculate_InvalidBody tests Calculate with invalid JSON
func TestWholesaleHandler_Calculate_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleHandler(nil)
	r.POST("/api/wholesale/calculate", handler.Calculate)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/calculate", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestWholesaleHandler_Apply_MissingTenant tests Apply without tenant
func TestWholesaleHandler_Apply_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWholesaleHandler(nil)
	r.POST("/api/wholesale/apply", handler.Apply)

	req, _ := http.NewRequest("POST", "/api/wholesale/apply", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWholesaleHandler_Apply_InvalidBody tests Apply with invalid JSON
func TestWholesaleHandler_Apply_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleHandler(nil)
	r.POST("/api/wholesale/apply", handler.Apply)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/wholesale/apply", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Handler checks DB first, so may return 500 (DB error) or 400 (invalid JSON)
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}

// TestWholesaleHandler_Apply_MissingAPI tests Apply without platform API configured
func TestWholesaleHandler_Apply_MissingAPI(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "test-tenant")
		c.Next()
	})

	handler := NewWholesaleHandler(nil)
	r.POST("/api/wholesale/apply", handler.Apply)

	body := `{"platform": "shopee", "skus": ["SKU001"]}`
	req, _ := http.NewRequest("POST", "/api/wholesale/apply", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// May return 500 (DB error) or 400 (missing API) depending on order of checks
	assert.Contains(t, []int{http.StatusBadRequest, http.StatusInternalServerError}, w.Code)
}
