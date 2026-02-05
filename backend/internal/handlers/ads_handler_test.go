package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestParseDateRange tests the date range parsing utility
func TestParseDateRange_Default(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/test", func(c *gin.Context) {
		start, end := parseDateRange(c)
		c.JSON(http.StatusOK, gin.H{
			"start": start.Format("2006-01-02"),
			"end":   end.Format("2006-01-02"),
		})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	// Default should return dates (we just check they exist)
	assert.NotEmpty(t, resp["start"])
	assert.NotEmpty(t, resp["end"])
}

// TestParseDateRange_WithParams tests date range with query params
func TestParseDateRange_WithParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/test", func(c *gin.Context) {
		start, end := parseDateRange(c)
		c.JSON(http.StatusOK, gin.H{
			"start": start.Format("2006-01-02"),
			"end":   end.Format("2006-01-02"),
		})
	})

	req, _ := http.NewRequest("GET", "/test?startDate=2025-01-01&endDate=2025-01-31", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "2025-01-01", resp["start"])
	assert.Equal(t, "2025-01-31", resp["end"])
}

// TestValidateTenantID_Success tests tenant ID validation success
func TestValidateTenantID_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Set tenant ID in middleware
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/test", func(c *gin.Context) {
		tenantID, ok := validateTenantID(c)
		if !ok {
			return
		}
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "tenant-123", resp["tenant_id"])
}

// TestValidateTenantID_Missing tests tenant ID validation failure
func TestValidateTenantID_Missing(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// No tenant ID set
	r.GET("/test", func(c *gin.Context) {
		tenantID, ok := validateTenantID(c)
		if !ok {
			return // Error already sent
		}
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "tenantId")
}

// TestGetShopeeAds_MissingTenant tests Shopee ads endpoint without tenant
func TestGetShopeeAds_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/shopee", handler.GetShopeeAds)

	req, _ := http.NewRequest("GET", "/api/ads/shopee", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail because no tenant ID
	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "tenantId")
}

// TestGetShopeeAdsSummary_MissingTenant tests Shopee summary endpoint without tenant
func TestGetShopeeAdsSummary_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/shopee/summary", handler.GetShopeeAdsSummary)

	req, _ := http.NewRequest("GET", "/api/ads/shopee/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetShopeeAdsTrends_MissingTenant tests Shopee trends endpoint without tenant
func TestGetShopeeAdsTrends_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/shopee/trends", handler.GetShopeeAdsTrends)

	req, _ := http.NewRequest("GET", "/api/ads/shopee/trends", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetShopeeProductPerformance_MissingTenant tests Shopee performance without tenant
func TestGetShopeeProductPerformance_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/shopee/performance", handler.GetShopeeProductPerformance)

	req, _ := http.NewRequest("GET", "/api/ads/shopee/performance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestUploadShopeeAds_MissingTenant tests Shopee upload without tenant
func TestUploadShopeeAds_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.POST("/api/ads/shopee/upload", handler.UploadShopeeAds)

	req, _ := http.NewRequest("POST", "/api/ads/shopee/upload", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail because no tenant ID
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetTiktokAds_MissingTenant tests TikTok ads endpoint without tenant
func TestGetTiktokAds_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/tiktok", handler.GetTiktokAds)

	req, _ := http.NewRequest("GET", "/api/ads/tiktok", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetTiktokAdsSummary_MissingTenant tests TikTok summary endpoint without tenant
func TestGetTiktokAdsSummary_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/tiktok/summary", handler.GetTiktokAdsSummary)

	req, _ := http.NewRequest("GET", "/api/ads/tiktok/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetTiktokAdsTrends_MissingTenant tests TikTok trends without tenant
func TestGetTiktokAdsTrends_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/tiktok/trends", handler.GetTiktokAdsTrends)

	req, _ := http.NewRequest("GET", "/api/ads/tiktok/trends", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetTiktokProductPerformance_MissingTenant tests TikTok performance without tenant
func TestGetTiktokProductPerformance_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/tiktok/performance", handler.GetTiktokProductPerformance)

	req, _ := http.NewRequest("GET", "/api/ads/tiktok/performance", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetTiktokPredictions_MissingTenant tests TikTok predictions without tenant
func TestGetTiktokPredictions_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.GET("/api/ads/tiktok/predictions", handler.GetTiktokPredictions)

	req, _ := http.NewRequest("GET", "/api/ads/tiktok/predictions", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestUploadTiktokAds_MissingTenant tests TikTok upload without tenant
func TestUploadTiktokAds_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewAdsHandler(nil)

	r.POST("/api/ads/tiktok/upload", handler.UploadTiktokAds)

	req, _ := http.NewRequest("POST", "/api/ads/tiktok/upload", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// NOTE: TestUploadShopeeAds_MissingFile and TestUploadTiktokAds_MissingFile
// require a DB mock to test properly. The handler validates tenant and creates
// a service (which needs DB) BEFORE checking for the file. These tests are
// skipped until a proper mock DB can be injected.
