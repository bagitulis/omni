package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Marketplace Sync History Handler Tests
// =============================================================================

func setupSyncHistoryRouter() (*gin.Engine, *MarketplaceSyncHistoryHandler) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// Handler with nil repo — validation fires before repo is touched
	handler := NewMarketplaceSyncHistoryHandler(nil)
	return r, handler
}

func TestSyncHistoryCreate_MissingTenant(t *testing.T) {
	r, handler := setupSyncHistoryRouter()
	r.POST("/api/marketplace-sync-history", handler.Create)

	body := `{"sku":"SKU-1","platform":"shopee","operation":"stock_update","status":"success"}`
	req, _ := http.NewRequest("POST", "/api/marketplace-sync-history", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "tenant_id")
}

func TestSyncHistoryCreate_MissingRequiredFields(t *testing.T) {
	r, handler := setupSyncHistoryRouter()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-test")
		c.Next()
	})
	r.POST("/api/marketplace-sync-history", handler.Create)

	// Missing all required fields
	body := `{}`
	req, _ := http.NewRequest("POST", "/api/marketplace-sync-history", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestSyncHistoryCreate_InvalidPlatform(t *testing.T) {
	r, handler := setupSyncHistoryRouter()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-test")
		c.Next()
	})
	r.POST("/api/marketplace-sync-history", handler.Create)

	body := `{"sku":"SKU-1","platform":"amazon","operation":"stock_update","status":"success"}`
	req, _ := http.NewRequest("POST", "/api/marketplace-sync-history", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "Invalid platform")
	assert.Contains(t, resp["error"].(string), "amazon")
	assert.Contains(t, resp["error"].(string), "shopee")
}

func TestSyncHistoryCreate_InvalidOperation(t *testing.T) {
	r, handler := setupSyncHistoryRouter()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-test")
		c.Next()
	})
	r.POST("/api/marketplace-sync-history", handler.Create)

	body := `{"sku":"SKU-1","platform":"shopee","operation":"delete_product","status":"success"}`
	req, _ := http.NewRequest("POST", "/api/marketplace-sync-history", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "Invalid operation")
	assert.Contains(t, resp["error"].(string), "delete_product")
	assert.Contains(t, resp["error"].(string), "stock_update")
}

func TestSyncHistoryCreate_InvalidStatus(t *testing.T) {
	r, handler := setupSyncHistoryRouter()
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-test")
		c.Next()
	})
	r.POST("/api/marketplace-sync-history", handler.Create)

	body := `{"sku":"SKU-1","platform":"shopee","operation":"stock_update","status":"pending"}`
	req, _ := http.NewRequest("POST", "/api/marketplace-sync-history", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "Invalid status")
	assert.Contains(t, resp["error"].(string), "pending")
	assert.Contains(t, resp["error"].(string), "success")
}

func TestSyncHistoryList_MissingTenant(t *testing.T) {
	r, handler := setupSyncHistoryRouter()
	r.GET("/api/marketplace-sync-history", handler.List)

	req, _ := http.NewRequest("GET", "/api/marketplace-sync-history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "tenant_id")
}

func TestSyncHistoryCreate_AllValidPlatforms(t *testing.T) {
	for _, platform := range []string{"shopee", "tiktok", "lazada"} {
		t.Run(platform, func(t *testing.T) {
			r, handler := setupSyncHistoryRouter()
			r.Use(gin.Recovery()) // Recover from nil repo panic
			r.Use(func(c *gin.Context) {
				c.Set("tenantID", "tenant-test")
				c.Next()
			})
			r.POST("/api/marketplace-sync-history", handler.Create)

			body := `{"sku":"SKU-1","platform":"` + platform + `","operation":"stock_update","status":"success"}`
			req, _ := http.NewRequest("POST", "/api/marketplace-sync-history", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			// Should NOT be 400 — validation passes (500 from nil repo panic is expected)
			assert.NotEqual(t, http.StatusBadRequest, w.Code, "Platform %q should be accepted", platform)
		})
	}
}

func TestSyncHistoryCreate_AllValidOperations(t *testing.T) {
	for _, op := range []string{"stock_update", "price_update", "wholesale_update", "mpq_update", "clone"} {
		t.Run(op, func(t *testing.T) {
			r, handler := setupSyncHistoryRouter()
			r.Use(gin.Recovery())
			r.Use(func(c *gin.Context) {
				c.Set("tenantID", "tenant-test")
				c.Next()
			})
			r.POST("/api/marketplace-sync-history", handler.Create)

			body := `{"sku":"SKU-1","platform":"shopee","operation":"` + op + `","status":"success"}`
			req, _ := http.NewRequest("POST", "/api/marketplace-sync-history", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusBadRequest, w.Code, "Operation %q should be accepted", op)
		})
	}
}

func TestSyncHistoryCreate_AllValidStatuses(t *testing.T) {
	for _, status := range []string{"success", "failed", "partial"} {
		t.Run(status, func(t *testing.T) {
			r, handler := setupSyncHistoryRouter()
			r.Use(gin.Recovery())
			r.Use(func(c *gin.Context) {
				c.Set("tenantID", "tenant-test")
				c.Next()
			})
			r.POST("/api/marketplace-sync-history", handler.Create)

			body := `{"sku":"SKU-1","platform":"shopee","operation":"stock_update","status":"` + status + `"}`
			req, _ := http.NewRequest("POST", "/api/marketplace-sync-history", bytes.NewBufferString(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.NotEqual(t, http.StatusBadRequest, w.Code, "Status %q should be accepted", status)
		})
	}
}
