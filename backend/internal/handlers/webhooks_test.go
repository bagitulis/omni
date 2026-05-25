package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/omni/backend/internal/services/webhooks"
)

// TestWebhookHandler_NewWebhookHandler tests handler creation
func TestWebhookHandler_NewWebhookHandler(t *testing.T) {
	handler := NewWebhookHandler(nil, nil, nil, "./data")
	assert.NotNil(t, handler)
}

// TestWebhookHandler_NewWebhookHandlerWithCache tests handler creation with cache
func TestWebhookHandler_NewWebhookHandlerWithCache(t *testing.T) {
	handler := NewWebhookHandlerWithCache(nil, nil, nil, "./data", nil)
	assert.NotNil(t, handler)
}

// TestWebhookHandler_ShopeeWebhook_EmptyBody tests Shopee webhook with empty body
func TestWebhookHandler_ShopeeWebhook_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockProcessor := webhooks.NewShopeeWebhookProcessor(nil, nil)
	handler := NewWebhookHandler(mockProcessor, nil, nil, "./data")
	r.POST("/webhook/shopee", handler.ShopeeWebhook)

	// Provide tenant and empty body: Process fails with parse error -> handler returns 200
	req, _ := http.NewRequest("POST", "/webhook/shopee?tenant_id=testtenant", strings.NewReader(""))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestWebhookHandler_ShopeeWebhook_WithTenant tests Shopee webhook with tenant
func TestWebhookHandler_ShopeeWebhook_WithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockProcessor := webhooks.NewShopeeWebhookProcessor(nil, nil)
	handler := NewWebhookHandler(mockProcessor, nil, nil, "./data")
	r.POST("/webhook/shopee", handler.ShopeeWebhook)

	// Empty body causes Process parse error (non-signature) -> handler returns 200
	req, _ := http.NewRequest("POST", "/webhook/shopee?tenant_id=testtenant", strings.NewReader(""))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 200 (webhooks always acknowledge)
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestWebhookHandler_LazadaWebhook_EmptyBody tests Lazada webhook with empty body
func TestWebhookHandler_LazadaWebhook_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockProcessor := webhooks.NewLazadaWebhookProcessor(nil, nil)
	handler := NewWebhookHandler(nil, mockProcessor, nil, "./data")
	r.POST("/webhook/lazada", handler.LazadaWebhook)

	req, _ := http.NewRequest("POST", "/webhook/lazada?tenant_id=testtenant", strings.NewReader(""))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestWebhookHandler_TiktokWebhook_EmptyBody tests TikTok webhook with empty body
func TestWebhookHandler_TiktokWebhook_EmptyBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockProcessor := webhooks.NewTiktokWebhookProcessor(nil, nil)
	handler := NewWebhookHandler(nil, nil, mockProcessor, "./data")
	r.POST("/webhook/tiktok", handler.TiktokWebhook)

	req, _ := http.NewRequest("POST", "/webhook/tiktok?tenant_id=testtenant", strings.NewReader(""))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestWebhookHandler_TiktokWebhook_WithHeaders tests TikTok webhook with signature headers
func TestWebhookHandler_TiktokWebhook_WithHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockProcessor := webhooks.NewTiktokWebhookProcessor(nil, nil)
	handler := NewWebhookHandler(nil, nil, mockProcessor, "./data")
	r.POST("/webhook/tiktok", handler.TiktokWebhook)

	// Empty body causes Process parse error (non-signature) -> handler returns 200
	req, _ := http.NewRequest("POST", "/webhook/tiktok?tenant_id=testtenant", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-tts-signature", "test-signature")
	req.Header.Set("x-tts-timestamp", fmt.Sprintf("%d", time.Now().Unix()))
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

// TestWebhookHandler_GetLogs_MissingTenant tests GetLogs without tenant
func TestWebhookHandler_GetLogs_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWebhookHandler(nil, nil, nil, "./data")
	r.GET("/api/webhooks/logs", handler.GetLogs)

	req, _ := http.NewRequest("GET", "/api/webhooks/logs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWebhookHandler_GetLogsByPlatform_MissingTenant tests GetLogsByPlatform without tenant
func TestWebhookHandler_GetLogsByPlatform_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWebhookHandler(nil, nil, nil, "./data")
	r.GET("/api/webhooks/logs/:platform", handler.GetLogsByPlatform)

	req, _ := http.NewRequest("GET", "/api/webhooks/logs/shopee", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWebhookHandler_GetStats_MissingTenant tests GetStats without tenant
func TestWebhookHandler_GetStats_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewWebhookHandler(nil, nil, nil, "./data")
	r.GET("/api/webhooks/stats", handler.GetStats)

	req, _ := http.NewRequest("GET", "/api/webhooks/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestWebhookStats_Structure tests WebhookStats struct
func TestWebhookStats_Structure(t *testing.T) {
	stats := WebhookStats{
		Total:      100,
		ByStatus:   map[string]int64{"success": 90, "error": 10},
		ByPlatform: map[string]int64{"shopee": 50, "lazada": 30, "tiktok": 20},
	}

	assert.Equal(t, int64(100), stats.Total)
	assert.Equal(t, int64(90), stats.ByStatus["success"])
	assert.Equal(t, int64(50), stats.ByPlatform["shopee"])
}
