package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPlatformAuthHandler_GetTiktokShops(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		expectedStatus int
		checkError     bool
	}{
		{
			name: "missing_tenant_id_returns_401",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			checkError:     true,
		},
		{
			name: "valid_tenant_id_without_db_returns_error",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/platform-auth/tiktok/shops", handler.GetTiktokShops)

			req, _ := http.NewRequest("GET", "/api/platform-auth/tiktok/shops", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkError {
				assert.Equal(t, false, resp["success"])
			}
		})
	}
}

func TestPlatformAuthHandler_GetActiveTiktokShop(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		expectedStatus int
		checkError     bool
	}{
		{
			name: "missing_tenant_id_returns_401",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			checkError:     true,
		},
		{
			name: "valid_tenant_id_without_db_returns_error",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/platform-auth/tiktok/active-shop", handler.GetActiveTiktokShop)

			req, _ := http.NewRequest("GET", "/api/platform-auth/tiktok/active-shop", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkError {
				assert.Equal(t, false, resp["success"])
			}
		})
	}
}

func TestPlatformAuthHandler_GetStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		expectedStatus int
		checkError     bool
	}{
		{
			name: "missing_tenant_id_returns_401",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			checkError:     true,
		},
		{
			name: "valid_tenant_id_without_db_returns_error",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/platform-auth/status", handler.GetStatus)

			req, _ := http.NewRequest("GET", "/api/platform-auth/status", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkError {
				assert.Equal(t, false, resp["success"])
			}
		})
	}
}

func TestPlatformAuthHandler_GetLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		expectedStatus int
		checkError     bool
	}{
		{
			name: "missing_tenant_id_returns_401",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			checkError:     true,
		},
		{
			name: "valid_tenant_id_without_db_returns_error",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewPlatformAuthHandler(nil, "http://localhost:3000", "./data")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/platform-auth/logs", handler.GetLogs)

			req, _ := http.NewRequest("GET", "/api/platform-auth/logs", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkError {
				assert.Equal(t, false, resp["success"])
			}
		})
	}
}
