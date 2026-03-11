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

func TestSpreadsheetRegistryHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		query          string
		setupContext   func(c *gin.Context)
		expectedStatus int
		expectedError  bool
	}{
		{
			name:  "missing_tenant_id",
			query: "",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  true,
		},
		{
			name:  "valid_tenant_db_not_available",
			query: "",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:  "with_purpose_filter",
			query: "?purpose=inventory",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
		{
			name:  "with_active_filter",
			query: "?active=true",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewSpreadsheetRegistryHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/spreadsheets", handler.List)

			req, _ := http.NewRequest("GET", "/api/spreadsheets"+tt.query, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSpreadsheetRegistryHandler_Get(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		id             string
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name: "missing_tenant_id",
			id:   "1",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid_id_db_not_available",
			id:   "1",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewSpreadsheetRegistryHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.GET("/api/spreadsheets/:id", handler.Get)

			url := "/api/spreadsheets/" + tt.id
			req, _ := http.NewRequest("GET", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSpreadsheetRegistryHandler_Register(t *testing.T) {
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
				"name":           "Test Spreadsheet",
				"spreadsheet_id": "abc123",
			},
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid_request_db_not_available",
			body: map[string]interface{}{
				"name":           "Test Spreadsheet",
				"spreadsheet_id": "abc123",
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewSpreadsheetRegistryHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/spreadsheets", handler.Register)

			body, _ := json.Marshal(tt.body)
			req, _ := http.NewRequest("POST", "/api/spreadsheets", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSpreadsheetRegistryHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		id             string
		body           map[string]interface{}
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name: "missing_tenant_id",
			id:   "1",
			body: map[string]interface{}{
				"name": "Updated Spreadsheet",
			},
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid_request_db_not_available",
			id:   "1",
			body: map[string]interface{}{
				"name": "Updated Spreadsheet",
			},
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewSpreadsheetRegistryHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.PUT("/api/spreadsheets/:id", handler.Update)

			body, _ := json.Marshal(tt.body)
			url := "/api/spreadsheets/" + tt.id
			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSpreadsheetRegistryHandler_Delete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		id             string
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name: "missing_tenant_id",
			id:   "1",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid_id_db_not_available",
			id:   "1",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewSpreadsheetRegistryHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.DELETE("/api/spreadsheets/:id", handler.Delete)

			url := "/api/spreadsheets/" + tt.id
			req, _ := http.NewRequest("DELETE", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestSpreadsheetRegistryHandler_MarkSynced(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		id             string
		setupContext   func(c *gin.Context)
		expectedStatus int
	}{
		{
			name: "missing_tenant_id",
			id:   "1",
			setupContext: func(c *gin.Context) {
				// No tenantID set
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "valid_id_db_not_available",
			id:   "1",
			setupContext: func(c *gin.Context) {
				c.Set("tenant_id", "test-tenant")
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			handler := NewSpreadsheetRegistryHandler(nil)

			r.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			r.POST("/api/spreadsheets/:id/synced", handler.MarkSynced)

			url := "/api/spreadsheets/" + tt.id + "/synced"
			req, _ := http.NewRequest("POST", url, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
