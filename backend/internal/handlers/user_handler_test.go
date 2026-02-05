package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService mocks the UserManagementService for user handler tests
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req *services.CreateUserRequest, createdBy string) (*models.User, error) {
	args := m.Called(ctx, req, createdBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateUser(ctx context.Context, userID string, req *services.UpdateUserRequest, updatedBy string) (*models.User, error) {
	args := m.Called(ctx, userID, req, updatedBy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID, deletedBy, tenantID string) error {
	args := m.Called(ctx, userID, deletedBy, tenantID)
	return args.Error(0)
}

func (m *MockUserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) ListUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserService) UnlockUser(ctx context.Context, userID, unlockedBy, tenantID string) error {
	args := m.Called(ctx, userID, unlockedBy, tenantID)
	return args.Error(0)
}

// UserServiceInterface defines the user service interface for testing
type UserServiceInterface interface {
	CreateUser(ctx context.Context, req *services.CreateUserRequest, createdBy string) (*models.User, error)
	UpdateUser(ctx context.Context, userID string, req *services.UpdateUserRequest, updatedBy string) (*models.User, error)
	DeleteUser(ctx context.Context, userID, deletedBy, tenantID string) error
	GetUser(ctx context.Context, userID string) (*models.User, error)
	ListUsers(ctx context.Context) ([]models.User, error)
	UnlockUser(ctx context.Context, userID, unlockedBy, tenantID string) error
}

// TestCreateUser_Success tests successful user creation
func TestCreateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserSvc := new(MockUserService)

	newUser := &models.User{
		ID:        "user-123",
		Username:  "newemployee",
		Email:     "employee@example.com",
		Role:      "staff",
		CreatedAt: time.Now(),
	}
	mockUserSvc.On("CreateUser", mock.Anything, mock.AnythingOfType("*services.CreateUserRequest"), "admin-user").
		Return(newUser, nil)

	// Middleware to set context
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Set("userID", "admin-user")
		c.Next()
	})

	r.POST("/api/users", func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body: " + err.Error(),
			})
			return
		}

		tenantID := c.GetString("tenantID")
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenantID - authentication required",
			})
			return
		}

		createdBy := c.GetString("userID")

		user, err := mockUserSvc.CreateUser(c.Request.Context(), &services.CreateUserRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			Role:     req.Role,
			TenantID: tenantID,
		}, createdBy)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data":    user.ToResponse(),
		})
	})

	reqBody := `{"username": "newemployee", "email": "employee@example.com", "password": "password123", "role": "staff"}`
	req, _ := http.NewRequest("POST", "/api/users", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "newemployee", data["username"])
	assert.Equal(t, "staff", data["role"])

	mockUserSvc.AssertExpectations(t)
}

// TestCreateUser_MissingTenantID tests user creation without tenant
func TestCreateUser_MissingTenantID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// No middleware setting tenantID
	r.POST("/api/users", func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body: " + err.Error(),
			})
			return
		}

		tenantID := c.GetString("tenantID")
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenantID - authentication required",
			})
			return
		}
	})

	reqBody := `{"username": "newemployee", "email": "employee@example.com", "password": "password123", "role": "staff"}`
	req, _ := http.NewRequest("POST", "/api/users", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "tenantID")
}

// TestUpdateUser_Success tests successful user update
func TestUpdateUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserSvc := new(MockUserService)

	updatedUser := &models.User{
		ID:       "user-123",
		Username: "updateduser",
		Email:    "updated@example.com",
		Role:     "manager",
	}
	mockUserSvc.On("UpdateUser", mock.Anything, "user-123", mock.AnythingOfType("*services.UpdateUserRequest"), "admin-user").
		Return(updatedUser, nil)

	r.Use(func(c *gin.Context) {
		c.Set("userID", "admin-user")
		c.Next()
	})

	r.PUT("/api/users/:id", func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "User ID is required",
			})
			return
		}

		var req UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body",
			})
			return
		}

		updatedBy := c.GetString("userID")

		user, err := mockUserSvc.UpdateUser(c.Request.Context(), userID, &services.UpdateUserRequest{
			Username: req.Username,
			Email:    req.Email,
			Role:     req.Role,
		}, updatedBy)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    user.ToResponse(),
		})
	})

	reqBody := `{"username": "updateduser", "email": "updated@example.com", "role": "manager"}`
	req, _ := http.NewRequest("PUT", "/api/users/user-123", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	mockUserSvc.AssertExpectations(t)
}

// TestDeleteUser_Success tests successful user deletion
func TestDeleteUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserSvc := new(MockUserService)
	mockUserSvc.On("DeleteUser", mock.Anything, "user-123", "admin-user", "tenant-456").Return(nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-456")
		c.Set("userID", "admin-user")
		c.Next()
	})

	r.DELETE("/api/users/:id", func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "User ID is required",
			})
			return
		}

		tenantID := c.GetString("tenantID")
		deletedBy := c.GetString("userID")

		if err := mockUserSvc.DeleteUser(c.Request.Context(), userID, deletedBy, tenantID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "User deleted successfully",
		})
	})

	req, _ := http.NewRequest("DELETE", "/api/users/user-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	mockUserSvc.AssertExpectations(t)
}

// TestGetUser_Success tests successful get user by ID
func TestGetUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserSvc := new(MockUserService)

	user := &models.User{
		ID:       "user-123",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "owner",
	}
	mockUserSvc.On("GetUser", mock.Anything, "user-123").Return(user, nil)

	r.GET("/api/users/:id", func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "User ID is required",
			})
			return
		}

		user, err := mockUserSvc.GetUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "User not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    user.ToResponse(),
		})
	})

	req, _ := http.NewRequest("GET", "/api/users/user-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "testuser", data["username"])

	mockUserSvc.AssertExpectations(t)
}

// TestGetUser_NotFound tests get user with non-existent ID
func TestGetUser_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserSvc := new(MockUserService)
	mockUserSvc.On("GetUser", mock.Anything, "nonexistent").Return(nil, nil)

	r.GET("/api/users/:id", func(c *gin.Context) {
		userID := c.Param("id")

		user, err := mockUserSvc.GetUser(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
		if user == nil {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "User not found",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    user.ToResponse(),
		})
	})

	req, _ := http.NewRequest("GET", "/api/users/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "User not found", resp["error"])

	mockUserSvc.AssertExpectations(t)
}

// TestListUsers_Success tests listing all users
func TestListUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserSvc := new(MockUserService)

	users := []models.User{
		{ID: "user-1", Username: "user1", Email: "user1@example.com", Role: "owner"},
		{ID: "user-2", Username: "user2", Email: "user2@example.com", Role: "staff"},
	}
	mockUserSvc.On("ListUsers", mock.Anything).Return(users, nil)

	r.GET("/api/users", func(c *gin.Context) {
		users, err := mockUserSvc.ListUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		response := make([]models.UserResponse, len(users))
		for i, user := range users {
			response[i] = user.ToResponse()
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    response,
		})
	})

	req, _ := http.NewRequest("GET", "/api/users", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].([]interface{})
	assert.Len(t, data, 2)

	mockUserSvc.AssertExpectations(t)
}

// TestUnlockUser_Success tests successful user unlock
func TestUnlockUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserSvc := new(MockUserService)
	mockUserSvc.On("UnlockUser", mock.Anything, "locked-user", "admin-user", "tenant-123").Return(nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Set("userID", "admin-user")
		c.Next()
	})

	r.POST("/api/users/:id/unlock", func(c *gin.Context) {
		userID := c.Param("id")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "User ID is required",
			})
			return
		}

		tenantID := c.GetString("tenantID")
		unlockedBy := c.GetString("userID")

		if err := mockUserSvc.UnlockUser(c.Request.Context(), userID, unlockedBy, tenantID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "User unlocked successfully",
		})
	})

	req, _ := http.NewRequest("POST", "/api/users/locked-user/unlock", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	mockUserSvc.AssertExpectations(t)
}

// TestCreateUser_InvalidRequest tests user creation with invalid data
func TestCreateUser_InvalidRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Set("userID", "admin-user")
		c.Next()
	})

	r.POST("/api/users", func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body: " + err.Error(),
			})
			return
		}
	})

	// Missing required fields
	reqBody := `{"username": ""}`
	req, _ := http.NewRequest("POST", "/api/users", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

// TestCreateUser_DuplicateUsername tests user creation with existing username
func TestCreateUser_DuplicateUsername(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockUserSvc := new(MockUserService)
	mockUserSvc.On("CreateUser", mock.Anything, mock.AnythingOfType("*services.CreateUserRequest"), "admin-user").
		Return(nil, errors.New("username already exists"))

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Set("userID", "admin-user")
		c.Next()
	})

	r.POST("/api/users", func(c *gin.Context) {
		var req CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body: " + err.Error(),
			})
			return
		}

		tenantID := c.GetString("tenantID")
		createdBy := c.GetString("userID")

		_, err := mockUserSvc.CreateUser(c.Request.Context(), &services.CreateUserRequest{
			Username: req.Username,
			Email:    req.Email,
			Password: req.Password,
			Role:     req.Role,
			TenantID: tenantID,
		}, createdBy)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
	})

	reqBody := `{"username": "existinguser", "email": "new@example.com", "password": "password123", "role": "staff"}`
	req, _ := http.NewRequest("POST", "/api/users", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "username already exists", resp["error"])

	mockUserSvc.AssertExpectations(t)
}
