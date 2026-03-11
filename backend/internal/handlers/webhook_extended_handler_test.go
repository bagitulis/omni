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

func TestWebhookExtendedHandler_ShopeeWebhookTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		tenantID       string
		body           string
		expectedStatus int
	}{
		{
			name:           "missing_tenant_id",
			tenantID:       "",
			body:           `{"code": 0, "data": {}}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid_tenant_id",
			tenantID:       "test-tenant",
			body:           `{"code": 0, "data": {"order_sn": "123"}}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewWebhookExtendedHandler(nil, nil, nil, "")

			r.POST("/api/webhooks/:tenantId/shopee", handler.ShopeeWebhookTenant)

			url := "/api/webhooks/" + tt.tenantID + "/shopee"
			if tt.tenantID == "" {
				url = "/api/webhooks//shopee"
			}

			req, _ := http.NewRequest("POST", url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.True(t, resp["success"].(bool))
			}
		})
	}
}

func TestWebhookExtendedHandler_LazadaWebhookTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		tenantID       string
		body           string
		expectedStatus int
	}{
		{
			name:           "missing_tenant_id",
			tenantID:       "",
			body:           `{"message_type": "order_status_update"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid_tenant_id",
			tenantID:       "test-tenant",
			body:           `{"message_type": "order_status_update", "data": {}}`,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewWebhookExtendedHandler(nil, nil, nil, "")

			r.POST("/api/webhooks/:tenantId/lazada", handler.LazadaWebhookTenant)

			url := "/api/webhooks/" + tt.tenantID + "/lazada"
			if tt.tenantID == "" {
				url = "/api/webhooks//lazada"
			}

			req, _ := http.NewRequest("POST", url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.True(t, resp["success"].(bool))
			}
		})
	}
}

func TestWebhookExtendedHandler_TiktokWebhookTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		tenantID       string
		body           string
		signature      string
		timestamp      string
		expectedStatus int
	}{
		{
			name:           "missing_tenant_id",
			tenantID:       "",
			body:           `{"type": "ORDER_STATUS_CHANGE"}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "valid_tenant_id",
			tenantID:       "test-tenant",
			body:           `{"type": "ORDER_STATUS_CHANGE", "data": {}}`,
			signature:      "test-sig",
			timestamp:      "1234567890",
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewWebhookExtendedHandler(nil, nil, nil, "")

			r.POST("/api/webhooks/:tenantId/tiktok", handler.TiktokWebhookTenant)

			url := "/api/webhooks/" + tt.tenantID + "/tiktok"
			if tt.tenantID == "" {
				url = "/api/webhooks//tiktok"
			}

			req, _ := http.NewRequest("POST", url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			if tt.signature != "" {
				req.Header.Set("x-tts-signature", tt.signature)
				req.Header.Set("x-tts-timestamp", tt.timestamp)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.True(t, resp["success"].(bool))
			}
		})
	}
}

func TestWebhookExtendedHandler_TestWebhook(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		body           map[string]interface{}
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name: "missing_tenant_id",
			body: map[string]interface{}{
				"platform":   "shopee",
				"event_type": "order_status_update",
			},
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid_request",
			body: map[string]interface{}{
				"platform":   "shopee",
				"event_type": "order_status_update",
				"data":       map[string]interface{}{"order_sn": "123"},
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing_required_fields",
			body: map[string]interface{}{
				"data": map[string]interface{}{},
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewWebhookExtendedHandler(nil, nil, nil, "")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/webhooks/test", handler.TestWebhook)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/webhooks/test", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.True(t, resp["success"].(bool))
			}
		})
	}
}

func TestWebhookExtendedHandler_GetWebhookConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name: "missing_tenant_id",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid_tenant_returns_config",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewWebhookExtendedHandler(nil, nil, nil, "")

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/webhooks/config", handler.GetWebhookConfig)

			req, _ := http.NewRequest("GET", "/api/webhooks/config", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var resp map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &resp)
				assert.True(t, resp["success"].(bool))

				data := resp["data"].(map[string]interface{})
				assert.NotNil(t, data["shopee"])
				assert.NotNil(t, data["lazada"])
				assert.NotNil(t, data["tiktok"])
			}
		})
	}
}
