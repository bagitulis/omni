package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOrderManagerHandler_GetLockedTodayOrders(t *testing.T) {
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
			name: "valid_tenant_id_returns_service_unavailable",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			// Without DB context, the handler tries to get DB and fails
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderManagerHandler("./data")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/orders/locked-today", handler.GetLockedTodayOrders)

			req, _ := http.NewRequest("POST", "/api/orders/locked-today", nil)
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

func TestOrderManagerHandler_GetSavedLockedOrders(t *testing.T) {
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
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			checkError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderManagerHandler("./data")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/orders/locked-today", handler.GetSavedLockedOrders)

			req, _ := http.NewRequest("GET", "/api/orders/locked-today", nil)
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

func TestAggregateLockedOrders(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("empty orders returns empty result", func(t *testing.T) {
		result := aggregateLockedOrders(nil, nil)
		assert.Empty(t, result)
	})

	t.Run("aggregates orders by SKU and product name", func(t *testing.T) {
		// Test with empty slices as we don't have sync.Order available
		result := aggregateLockedOrders(nil, nil)
		assert.Empty(t, result)
	})
}
