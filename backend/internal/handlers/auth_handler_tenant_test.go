package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestGetTenants_NoAuthService tests GetTenants with nil auth service
// This test verifies that calling without proper service panics (expected behavior)
func TestGetTenants_NoAuthService(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// The handler will panic with nil multiTenantAuth
	// We use defer/recover to catch this expected panic
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic with nil auth service:", r)
		} else {
			t.Error("Expected panic with nil auth service, but no panic occurred")
		}
	}()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/auth/tenants", nil)

	h := &AuthHandler{}
	h.GetTenants(c)
}

// TestSwitchTenant_MissingBody tests SwitchTenant with missing body
func TestSwitchTenant_MissingBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/auth/switch-tenant", nil)

	h := &AuthHandler{}
	h.SwitchTenant(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestSwitchTenant_EmptyTenantID tests SwitchTenant with empty tenant ID in body
func TestSwitchTenant_EmptyTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := SwitchTenantRequest{TenantID: ""}
	jsonBody, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/api/auth/switch-tenant", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &AuthHandler{}
	h.SwitchTenant(c)

	// SwitchTenantRequest.TenantID has binding:"required", so empty string should fail
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestVerifyToken_MissingToken_Direct tests VerifyToken with missing token
func TestVerifyToken_MissingToken_Direct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/auth/verify", nil)

	h := &AuthHandler{}
	h.VerifyToken(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}
