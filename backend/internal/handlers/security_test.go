package handlers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestSecurityHandler_NewSecurityHandler tests handler creation
func TestSecurityHandler_NewSecurityHandler(t *testing.T) {
	handler := NewSecurityHandler(nil, nil)
	assert.NotNil(t, handler)
}

// TestSecurityHandler_ReportIssue_MissingBody tests ReportIssue with no body
func TestSecurityHandler_ReportIssue_MissingBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewSecurityHandler(nil, nil)
	r.POST("/api/security/report", handler.ReportIssue)

	req, _ := http.NewRequest("POST", "/api/security/report", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should fail validation with missing body
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSecurityHandler_ReportIssue_MissingRequiredFields tests ReportIssue without required fields
func TestSecurityHandler_ReportIssue_MissingRequiredFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewSecurityHandler(nil, nil)
	r.POST("/api/security/report", handler.ReportIssue)

	body := `{"type": "test"}`
	req, _ := http.NewRequest("POST", "/api/security/report", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Missing description should fail
	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// TestSecurityHandler_ReportIssue_InvalidJSON tests ReportIssue with invalid JSON
func TestSecurityHandler_ReportIssue_InvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenant_id", "test-tenant")
		c.Next()
	})

	handler := NewSecurityHandler(nil, nil)
	r.POST("/api/security/report", handler.ReportIssue)

	body := `{invalid json`
	req, _ := http.NewRequest("POST", "/api/security/report", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Invalid JSON should fail
	assert.Equal(t, http.StatusBadRequest, w.Code)
}
