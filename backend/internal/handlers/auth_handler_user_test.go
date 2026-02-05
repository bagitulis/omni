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

// TestChangePassword_MissingBody tests ChangePassword with missing body
func TestChangePassword_MissingBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/auth/change-password", nil)

	h := &AuthHandler{}
	h.ChangePassword(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestChangePassword_MissingUserID_Direct tests ChangePassword without userID in context
func TestChangePassword_MissingUserID_Direct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := ChangePasswordRequest{OldPassword: "old", NewPassword: "newpassword123"}
	jsonBody, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/api/auth/change-password", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")
	// No userID set

	h := &AuthHandler{}
	h.ChangePassword(c)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// TestGetCurrentUser_Success_Direct tests GetCurrentUser with user context
func TestGetCurrentUser_Success_Direct(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/auth/me", nil)
	c.Set("userID", "test-user-id")
	c.Set("tenantID", "test-tenant-id")
	c.Set("role", "admin")

	h := &AuthHandler{}
	h.GetCurrentUser(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

// TestRegister_MissingBody tests Register with missing body
func TestRegister_MissingBody(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("POST", "/api/auth/register", nil)

	h := &AuthHandler{}
	h.Register(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// TestRegister_NoUserManagementService tests Register with nil service
func TestRegister_NoUserManagementService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}
	jsonBody, _ := json.Marshal(body)
	c.Request, _ = http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(jsonBody))
	c.Request.Header.Set("Content-Type", "application/json")

	h := &AuthHandler{}
	h.Register(c)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500 (no service), got %d", w.Code)
	}
}
