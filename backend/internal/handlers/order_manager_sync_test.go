package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestOrderManagerHandler_SyncAll(t *testing.T) {
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
			name: "valid_tenant_id_returns_success_with_empty_sync",
			setupContext: func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
			},
			// SyncAll gracefully handles missing credentials and returns success with 0 orders
			expectedStatus: http.StatusOK,
			checkError:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewOrderManagerHandler()

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/orders/sync-all", handler.SyncAll)

			req, _ := http.NewRequest("POST", "/api/orders/sync-all", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)

			if tt.checkError {
				assert.Equal(t, false, resp["success"])
			} else {
				assert.Equal(t, true, resp["success"])
			}
		})
	}
}
