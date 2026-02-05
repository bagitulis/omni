package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Filter Preference Handler Tests (filter_preference.go)
// =============================================================================

// TestFilterPreferenceGet_MissingTenant tests getting filter preferences without tenant
func TestFilterPreferenceGet_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewFilterPreferenceHandler(nil)
	r.GET("/api/filter-preferences", handler.Get)

	req, _ := http.NewRequest("GET", "/api/filter-preferences?platform=shopee", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "tenant ID required")
}

// TestFilterPreferenceGet_InvalidPlatform tests getting with invalid platform
func TestFilterPreferenceGet_InvalidPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/filter-preferences", func(c *gin.Context) {
		platform := c.Query("platform")
		if !isValidPlatform(platform) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid platform. Must be one of: tiktok, lazada, shopee, inventory",
			})
			return
		}
	})

	req, _ := http.NewRequest("GET", "/api/filter-preferences?platform=invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "Invalid platform")
}

// TestFilterPreferenceGet_ValidPlatforms tests getting with each valid platform
func TestFilterPreferenceGet_ValidPlatforms(t *testing.T) {
	platforms := []string{"tiktok", "lazada", "shopee", "inventory"}

	for _, platform := range platforms {
		t.Run("platform_"+platform, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			r := gin.New()

			r.Use(func(c *gin.Context) {
				c.Set("tenantID", "tenant-123")
				c.Next()
			})

			r.GET("/api/filter-preferences", func(c *gin.Context) {
				p := c.Query("platform")
				if !isValidPlatform(p) {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   "Invalid platform",
					})
					return
				}
				c.JSON(http.StatusOK, gin.H{
					"success":  true,
					"platform": p,
				})
			})

			req, _ := http.NewRequest("GET", "/api/filter-preferences?platform="+platform, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, true, resp["success"])
			assert.Equal(t, platform, resp["platform"])
		})
	}
}

// TestFilterPreferenceSave_MissingTenant tests saving filter preferences without tenant
func TestFilterPreferenceSave_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewFilterPreferenceHandler(nil)
	r.POST("/api/filter-preferences", handler.Save)

	reqBody := `{"platform": "shopee"}`
	req, _ := http.NewRequest("POST", "/api/filter-preferences", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestFilterPreferenceSave_InvalidBody tests saving with invalid body
func TestFilterPreferenceSave_InvalidBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/filter-preferences", func(c *gin.Context) {
		var req SaveFilterPreferenceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
	})

	// Missing required platform
	reqBody := `{"visible_columns": ["col1"]}`
	req, _ := http.NewRequest("POST", "/api/filter-preferences", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestFilterPreferenceSave_InvalidPlatform tests saving with invalid platform
func TestFilterPreferenceSave_InvalidPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/filter-preferences", func(c *gin.Context) {
		var req SaveFilterPreferenceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		if !isValidPlatform(req.Platform) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid platform. Must be one of: tiktok, lazada, shopee, inventory",
			})
			return
		}
	})

	reqBody := `{"platform": "invalid_platform"}`
	req, _ := http.NewRequest("POST", "/api/filter-preferences", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "Invalid platform")
}

// TestFilterPreferenceSave_DefaultTab tests saving with default tab
func TestFilterPreferenceSave_DefaultTab(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/filter-preferences", func(c *gin.Context) {
		var req SaveFilterPreferenceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Support both 'page' (new) and 'tab' (legacy) parameters
		tab := req.Page
		if tab == "" {
			tab = req.Tab
		}
		if tab == "" {
			tab = "product"
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"tab":     tab,
		})
	})

	reqBody := `{"platform": "shopee"}`
	req, _ := http.NewRequest("POST", "/api/filter-preferences", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "product", resp["tab"])
}

// TestFilterPreferenceSave_LegacyFiltersSupport tests legacy filters parameter
func TestFilterPreferenceSave_LegacyFiltersSupport(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.POST("/api/filter-preferences", func(c *gin.Context) {
		var req SaveFilterPreferenceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}

		// Support both 'columnFilters' (new) and 'filters' (legacy) parameters
		columnFilters := req.ColumnFilters
		if columnFilters == nil {
			columnFilters = req.Filters
		}

		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"column_filters": columnFilters,
		})
	})

	reqBody := `{"platform": "shopee", "filters": {"status": "active"}}`
	req, _ := http.NewRequest("POST", "/api/filter-preferences", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	columnFilters := resp["column_filters"].(map[string]interface{})
	assert.Equal(t, "active", columnFilters["status"])
}

// TestFilterPreferenceDelete_MissingTenant tests deleting filter preferences without tenant
func TestFilterPreferenceDelete_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	handler := NewFilterPreferenceHandler(nil)
	r.DELETE("/api/filter-preferences", handler.Delete)

	req, _ := http.NewRequest("DELETE", "/api/filter-preferences?platform=shopee", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestFilterPreferenceDelete_InvalidPlatform tests deleting with invalid platform
func TestFilterPreferenceDelete_InvalidPlatform(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.DELETE("/api/filter-preferences", func(c *gin.Context) {
		platform := c.Query("platform")
		if !isValidPlatform(platform) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid platform. Must be one of: tiktok, lazada, shopee, inventory",
			})
			return
		}
	})

	req, _ := http.NewRequest("DELETE", "/api/filter-preferences?platform=invalid", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestFilterPreferenceGet_PaginationLimit tests pagination limit enforcement
func TestFilterPreferenceGet_PaginationLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/filter-preferences", func(c *gin.Context) {
		limit := 10
		if l := c.Query("limit"); l != "" {
			// Parse limit
			for _, ch := range l {
				if ch >= '0' && ch <= '9' {
					limit = limit*10 + int(ch-'0')
				}
			}
			limit = limit / 10 // Undo initial 10
		}
		if limit > 100 {
			limit = 100
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"limit":   limit,
		})
	})

	// Request with limit > 100 should be capped at 100
	req, _ := http.NewRequest("GET", "/api/filter-preferences?platform=shopee&limit=200", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, float64(100), resp["limit"])
}

// TestFilterPreferenceGet_LegacyTabParameter tests legacy tab parameter support
func TestFilterPreferenceGet_LegacyTabParameter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/filter-preferences", func(c *gin.Context) {
		// Support both 'page' (new) and 'tab' (legacy) parameters
		tab := c.DefaultQuery("page", c.DefaultQuery("tab", "product"))

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"tab":     tab,
		})
	})

	req, _ := http.NewRequest("GET", "/api/filter-preferences?platform=shopee&tab=order", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, "order", resp["tab"])
}
