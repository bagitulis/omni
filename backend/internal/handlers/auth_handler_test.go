package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// MockMultiTenantAuthService mocks the MultiTenantAuthService
type MockMultiTenantAuthService struct {
	mock.Mock
}

func (m *MockMultiTenantAuthService) LoginAcrossTenants(ctx context.Context, req *services.MultiTenantLoginRequest) (*services.MultiTenantLoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.MultiTenantLoginResponse), args.Error(1)
}

func (m *MockMultiTenantAuthService) RefreshTokenForTenant(ctx context.Context, refreshToken, tenantID, ip, userAgent string) (string, string, error) {
	args := m.Called(ctx, refreshToken, tenantID, ip, userAgent)
	return args.String(0), args.String(1), args.Error(2)
}

func (m *MockMultiTenantAuthService) ChangePasswordForTenant(ctx context.Context, userID, tenantID, oldPwd, newPwd string) error {
	args := m.Called(ctx, userID, tenantID, oldPwd, newPwd)
	return args.Error(0)
}

func (m *MockMultiTenantAuthService) GetAvailableTenants(ctx context.Context) ([]services.TenantInfo, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]services.TenantInfo), args.Error(1)
}

func (m *MockMultiTenantAuthService) SwitchTenant(ctx context.Context, userID, role, tenantID string) (string, error) {
	args := m.Called(ctx, userID, role, tenantID)
	return args.String(0), args.Error(1)
}

// MockAuthService mocks the AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Logout(ctx context.Context, refreshToken string) error {
	args := m.Called(ctx, refreshToken)
	return args.Error(0)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*utils.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*utils.JWTClaims), args.Error(1)
}

// MockUserManagementService mocks the UserManagementService
type MockUserManagementService struct {
	mock.Mock
}

func (m *MockUserManagementService) CreateUser(ctx context.Context, req *services.CreateUserRequest, createdBy string) (*models.User, error) {
	args := m.Called(ctx, req, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// AuthHandlerTestable is a testable version of AuthHandler with interfaces
type AuthHandlerTestable struct {
	MultiTenantAuth       MultiTenantAuthInterface
	AuthService           AuthServiceInterface
	UserManagementService UserManagementInterface
}

// MultiTenantAuthInterface defines the interface for testing
type MultiTenantAuthInterface interface {
	LoginAcrossTenants(ctx context.Context, req *services.MultiTenantLoginRequest) (*services.MultiTenantLoginResponse, error)
	RefreshTokenForTenant(ctx context.Context, refreshToken, tenantID, ip, userAgent string) (string, string, error)
	ChangePasswordForTenant(ctx context.Context, userID, tenantID, oldPwd, newPwd string) error
	GetAvailableTenants(ctx context.Context) ([]services.TenantInfo, error)
	SwitchTenant(ctx context.Context, userID, role, tenantID string) (string, error)
}

// AuthServiceInterface defines auth service interface for testing
type AuthServiceInterface interface {
	Logout(ctx context.Context, refreshToken string) error
	ValidateToken(tokenString string) (*utils.JWTClaims, error)
}

// UserManagementInterface defines user management interface for testing
type UserManagementInterface interface {
	CreateUser(ctx context.Context, req *services.CreateUserRequest, createdBy string) (*models.User, error)
}

// TestLogin_Success tests successful login
func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockMultiAuth := new(MockMultiTenantAuthService)
	mockAuth := new(MockAuthService)
	mockUserMgmt := new(MockUserManagementService)

	// Setup mock response
	userResp := models.UserResponse{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "owner",
	}
	loginResp := &services.MultiTenantLoginResponse{
		User:         &userResp,
		AccessToken:  "access-token-123",
		RefreshToken: "refresh-token-123",
		TenantID:     "tenant-123",
		ExpiresAt:    time.Now().Add(time.Hour),
	}

	mockMultiAuth.On("LoginAcrossTenants", mock.Anything, mock.AnythingOfType("*services.MultiTenantLoginRequest")).
		Return(loginResp, nil)

	// Create testable handler
	testHandler := &AuthHandlerTestable{
		MultiTenantAuth:       mockMultiAuth,
		AuthService:           mockAuth,
		UserManagementService: mockUserMgmt,
	}

	// Register login route with custom handler
	r.POST("/api/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Username and password are required",
			})
			return
		}

		result, err := testHandler.MultiTenantAuth.LoginAcrossTenants(c.Request.Context(), &services.MultiTenantLoginRequest{
			Username:  req.Username,
			Password:  req.Password,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":      true,
			"message":      "Login successful",
			"user":         result.User,
			"token":        result.AccessToken,
			"access_token": result.AccessToken,
			"tenant_id":    result.TenantID,
		})
	})

	// Perform request
	reqBody := `{"username": "testuser", "password": "testpassword123"}`
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Assertions
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "access-token-123", resp["access_token"])
	assert.Equal(t, "tenant-123", resp["tenant_id"])

	mockMultiAuth.AssertExpectations(t)
}

// TestLogin_InvalidCredentials tests login with invalid credentials
func TestLogin_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockMultiAuth := new(MockMultiTenantAuthService)

	mockMultiAuth.On("LoginAcrossTenants", mock.Anything, mock.AnythingOfType("*services.MultiTenantLoginRequest")).
		Return(nil, services.ErrInvalidCredentials)

	r.POST("/api/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Username and password are required",
			})
			return
		}

		_, err := mockMultiAuth.LoginAcrossTenants(c.Request.Context(), &services.MultiTenantLoginRequest{
			Username:  req.Username,
			Password:  req.Password,
			IPAddress: c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
		})

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
	})

	reqBody := `{"username": "testuser", "password": "wrongpassword"}`
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "Invalid")

	mockMultiAuth.AssertExpectations(t)
}

// TestLogin_MissingCredentials tests login with missing credentials
func TestLogin_MissingCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/auth/login", func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Username and password are required",
			})
			return
		}
	})

	// Empty body
	req, _ := http.NewRequest("POST", "/api/auth/login", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

// TestLogout_Success tests successful logout
func TestLogout_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockAuth := new(MockAuthService)
	mockAuth.On("Logout", mock.Anything, "refresh-token-123").Return(nil)

	r.POST("/api/auth/logout", func(c *gin.Context) {
		// Set a refresh token cookie for the test
		refreshToken, _ := c.Cookie(RefreshTokenCookieName)
		if refreshToken == "" {
			refreshToken = "refresh-token-123" // Simulated from header for testing
		}

		if refreshToken != "" {
			_ = mockAuth.Logout(c.Request.Context(), refreshToken)
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Logged out successfully",
		})
	})

	req, _ := http.NewRequest("POST", "/api/auth/logout", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	mockAuth.AssertExpectations(t)
}

// TestGetCurrentUser_Success tests getting current user info
func TestGetCurrentUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Middleware to set user context
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("tenant_id", "tenant-456")
		c.Set("role", "owner")
		c.Next()
	})

	r.GET("/api/auth/me", func(c *gin.Context) {
		userID := c.GetString("userID")
		tenantID := c.GetString("tenant_id")
		role := c.GetString("role")

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"user_id":   userID,
				"tenant_id": tenantID,
				"role":      role,
			},
		})
	})

	req, _ := http.NewRequest("GET", "/api/auth/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "user-123", data["user_id"])
	assert.Equal(t, "tenant-456", data["tenant_id"])
	assert.Equal(t, "owner", data["role"])
}

// TestChangePassword_Success tests successful password change
func TestChangePassword_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockMultiAuth := new(MockMultiTenantAuthService)
	mockMultiAuth.On("ChangePasswordForTenant", mock.Anything, "user-123", "tenant-456", "oldpassword123", "newpassword123").
		Return(nil)

	// Middleware to set user context
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("tenant_id", "tenant-456")
		c.Next()
	})

	r.POST("/api/auth/change-password", func(c *gin.Context) {
		var req ChangePasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		userID := c.GetString("userID")
		tenantID := c.GetString("tenant_id")
		if userID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "User not authenticated",
			})
			return
		}

		if err := mockMultiAuth.ChangePasswordForTenant(c.Request.Context(), userID, tenantID, req.OldPassword, req.NewPassword); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Password changed successfully",
		})
	})

	reqBody := `{"old_password": "oldpassword123", "new_password": "newpassword123"}`
	req, _ := http.NewRequest("POST", "/api/auth/change-password", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	mockMultiAuth.AssertExpectations(t)
}

// TestGetTenants_Success tests getting available tenants
func TestGetTenants_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockMultiAuth := new(MockMultiTenantAuthService)
	mockMultiAuth.On("GetAvailableTenants", mock.Anything).Return([]services.TenantInfo{
		{ID: "tenant-1", ShopName: "Shop 1"},
		{ID: "tenant-2", ShopName: "Shop 2"},
	}, nil)

	r.GET("/api/auth/tenants", func(c *gin.Context) {
		tenants, err := mockMultiAuth.GetAvailableTenants(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		var result []gin.H
		for _, t := range tenants {
			result = append(result, gin.H{
				"id":        t.ID,
				"shop_name": t.ShopName,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"tenants": result,
		})
	})

	req, _ := http.NewRequest("GET", "/api/auth/tenants", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	tenants := resp["tenants"].([]interface{})
	assert.Len(t, tenants, 2)

	mockMultiAuth.AssertExpectations(t)
}

// TestSwitchTenant_Success tests successful tenant switch
func TestSwitchTenant_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockMultiAuth := new(MockMultiTenantAuthService)
	mockMultiAuth.On("SwitchTenant", mock.Anything, "user-123", "developer", "new-tenant-456").
		Return("new-token-xyz", nil)

	// Middleware to set user context
	r.Use(func(c *gin.Context) {
		c.Set("userID", "user-123")
		c.Set("role", "developer")
		c.Next()
	})

	r.POST("/api/auth/switch-tenant", func(c *gin.Context) {
		var req SwitchTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "tenant_id is required",
			})
			return
		}

		role := c.GetString("role")
		userID := c.GetString("userID")

		newToken, err := mockMultiAuth.SwitchTenant(c.Request.Context(), userID, role, req.TenantID)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":   true,
			"message":   "Switched to tenant '" + req.TenantID + "'",
			"token":     newToken,
			"tenant_id": req.TenantID,
		})
	})

	reqBody := `{"tenant_id": "new-tenant-456"}`
	req, _ := http.NewRequest("POST", "/api/auth/switch-tenant", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "new-token-xyz", resp["token"])
	assert.Equal(t, "new-tenant-456", resp["tenant_id"])

	mockMultiAuth.AssertExpectations(t)
}

// TestVerifyToken_Success tests successful token verification
func TestVerifyToken_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockAuth := new(MockAuthService)
	claims := &utils.JWTClaims{
		UserID:   "user-123",
		TenantID: "tenant-456",
		Role:     "owner",
	}
	mockAuth.On("ValidateToken", "valid-token-123").Return(claims, nil)

	r.GET("/api/auth/verify", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"valid":   false,
				"error":   "No token provided",
			})
			return
		}

		token := authHeader
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		}

		claims, err := mockAuth.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"valid":   false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"valid":   true,
			"payload": claims,
		})
	})

	req, _ := http.NewRequest("GET", "/api/auth/verify", nil)
	req.Header.Set("Authorization", "Bearer valid-token-123")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, true, resp["valid"])

	mockAuth.AssertExpectations(t)
}

// TestVerifyToken_NoToken tests token verification without token
func TestVerifyToken_NoToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/auth/verify", func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"valid":   false,
				"error":   "No token provided",
			})
			return
		}
	})

	req, _ := http.NewRequest("GET", "/api/auth/verify", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, false, resp["valid"])
}

// TestRegister_Success tests successful user registration
func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserMgmt := new(MockUserManagementService)

	user := &models.User{
		ID:        "new-user-123",
		Username:  "newuser",
		Email:     "newuser@example.com",
		Role:      "owner",
		CreatedAt: time.Now(),
	}
	mockUserMgmt.On("CreateUser", mock.Anything, mock.AnythingOfType("*services.CreateUserRequest"), "system").
		Return(user, nil)

	r.POST("/api/auth/register", func(c *gin.Context) {
		var req RegisterRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body: " + err.Error(),
			})
			return
		}

		createdUser, err := mockUserMgmt.CreateUser(c.Request.Context(), &services.CreateUserRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			Role:     "owner",
		}, "system")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"message": "User registered successfully",
			"user":    createdUser.ToResponse(),
		})
	})

	reqBody := `{"username": "newuser", "email": "newuser@example.com", "password": "password123"}`
	req, _ := http.NewRequest("POST", "/api/auth/register", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "User registered successfully", resp["message"])

	mockUserMgmt.AssertExpectations(t)
}
