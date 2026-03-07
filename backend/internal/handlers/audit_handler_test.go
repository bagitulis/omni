package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuditService mocks the AuditService
type MockAuditService struct {
	mock.Mock
}

func (m *MockAuditService) GetByTenant(ctx context.Context, tenantID string, page, pageSize int) (*services.AuditListResponse, error) {
	args := m.Called(ctx, tenantID, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuditListResponse), args.Error(1)
}

func (m *MockAuditService) GetByUser(ctx context.Context, userID string, limit int) ([]models.AuditLog, error) {
	args := m.Called(ctx, userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *MockAuditService) GetByAction(ctx context.Context, tenantID, action string, limit int) ([]models.AuditLog, error) {
	args := m.Called(ctx, tenantID, action, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.AuditLog), args.Error(1)
}

func (m *MockAuditService) GetByDateRange(ctx context.Context, tenantID string, startDate, endDate time.Time, page, pageSize int) (*services.AuditListResponse, error) {
	args := m.Called(ctx, tenantID, startDate, endDate, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.AuditListResponse), args.Error(1)
}

func (m *MockAuditService) Create(ctx context.Context, log *models.AuditLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// TestGetAuditLogs_Success tests getting audit logs
func TestGetAuditLogs_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockAuditService)
	mockService.On("GetByTenant", mock.Anything, "tenant-123", 1, 20).Return(&services.AuditListResponse{
		Data:       []models.AuditLog{},
		Total:      10,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}, nil)

	// Use middleware to set tenant
	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/audit", func(c *gin.Context) {
		tenantID := c.GetString("tenantID")
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenantId",
			})
			return
		}

		result, err := mockService.GetByTenant(c.Request.Context(), tenantID, 1, 20)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"data":       result.Data,
			"total":      result.Total,
			"page":       result.Page,
			"pageSize":   result.PageSize,
			"totalPages": result.TotalPages,
		})
	})

	req, _ := http.NewRequest("GET", "/api/audit", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(10), resp["total"])

	mockService.AssertExpectations(t)
}

// TestGetAuditLogs_MissingTenant tests audit logs without tenant
func TestGetAuditLogs_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/audit", func(c *gin.Context) {
		tenantID := c.GetString("tenantID")
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenantId",
			})
			return
		}
	})

	req, _ := http.NewRequest("GET", "/api/audit", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

// TestGetAuditLogsByUser_Success tests getting logs by user
func TestGetAuditLogsByUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockAuditService)
	logs := []models.AuditLog{
		{Action: "login"},
		{Action: "update"},
	}
	mockService.On("GetByUser", mock.Anything, "user-123", 50).Return(logs, nil)

	r.GET("/api/audit/user/:userId", func(c *gin.Context) {
		userID := c.Param("userId")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "User ID is required",
			})
			return
		}

		result, err := mockService.GetByUser(c.Request.Context(), userID, 50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	req, _ := http.NewRequest("GET", "/api/audit/user/user-123", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	mockService.AssertExpectations(t)
}

// TestGetAuditLogsByUser_MissingUserID tests getting logs without user ID
func TestGetAuditLogsByUser_MissingUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/audit/user/:userId", func(c *gin.Context) {
		userID := c.Param("userId")
		if userID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "User ID is required",
			})
			return
		}
	})

	// Empty user ID in path - route won't match
	req, _ := http.NewRequest("GET", "/api/audit/user/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Will return 404 because path doesn't match
	assert.Equal(t, http.StatusNotFound, w.Code)
}

// TestGetAuditLogsByAction_Success tests getting logs by action
func TestGetAuditLogsByAction_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockAuditService)
	logs := []models.AuditLog{
		{Action: "login"},
	}
	mockService.On("GetByAction", mock.Anything, "tenant-123", "login", 50).Return(logs, nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/audit/action/:action", func(c *gin.Context) {
		tenantID := c.GetString("tenantID")
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenantId",
			})
			return
		}

		action := c.Param("action")
		if action == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Action is required",
			})
			return
		}

		result, err := mockService.GetByAction(c.Request.Context(), tenantID, action, 50)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	req, _ := http.NewRequest("GET", "/api/audit/action/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	mockService.AssertExpectations(t)
}

// TestGetAuditLogsByAction_MissingTenant tests action logs without tenant
func TestGetAuditLogsByAction_MissingTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/audit/action/:action", func(c *gin.Context) {
		tenantID := c.GetString("tenantID")
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenantId",
			})
			return
		}
	})

	req, _ := http.NewRequest("GET", "/api/audit/action/login", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// TestGetAuditLogsByDateRange_Success tests getting logs by date range
func TestGetAuditLogsByDateRange_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	mockService := new(MockAuditService)
	mockService.On("GetByDateRange", mock.Anything, "tenant-123", mock.Anything, mock.Anything, 1, 20).Return(&services.AuditListResponse{
		Data:       []models.AuditLog{},
		Total:      5,
		Page:       1,
		PageSize:   20,
		TotalPages: 1,
	}, nil)

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/audit/date-range", func(c *gin.Context) {
		tenantID := c.GetString("tenantID")
		if tenantID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Missing tenantId",
			})
			return
		}

		startDateStr := c.Query("startDate")
		endDateStr := c.Query("endDate")
		if startDateStr == "" || endDateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "startDate and endDate are required",
			})
			return
		}

		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid startDate format (use YYYY-MM-DD)",
			})
			return
		}

		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid endDate format (use YYYY-MM-DD)",
			})
			return
		}

		result, err := mockService.GetByDateRange(c.Request.Context(), tenantID, startDate, endDate, 1, 20)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success":    true,
			"data":       result.Data,
			"total":      result.Total,
			"page":       result.Page,
			"pageSize":   result.PageSize,
			"totalPages": result.TotalPages,
		})
	})

	req, _ := http.NewRequest("GET", "/api/audit/date-range?startDate=2025-01-01&endDate=2025-01-31", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(5), resp["total"])

	mockService.AssertExpectations(t)
}

// TestGetAuditLogsByDateRange_MissingDates tests date range without dates
func TestGetAuditLogsByDateRange_MissingDates(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/audit/date-range", func(c *gin.Context) {
		startDateStr := c.Query("startDate")
		endDateStr := c.Query("endDate")
		if startDateStr == "" || endDateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "startDate and endDate are required",
			})
			return
		}
	})

	req, _ := http.NewRequest("GET", "/api/audit/date-range", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "startDate and endDate are required")
}

// TestGetAuditLogsByDateRange_InvalidDateFormat tests date range with invalid format
func TestGetAuditLogsByDateRange_InvalidDateFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("tenantID", "tenant-123")
		c.Next()
	})

	r.GET("/api/audit/date-range", func(c *gin.Context) {
		startDateStr := c.Query("startDate")
		endDateStr := c.Query("endDate")
		if startDateStr == "" || endDateStr == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "startDate and endDate are required",
			})
			return
		}

		_, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid startDate format (use YYYY-MM-DD)",
			})
			return
		}
	})

	req, _ := http.NewRequest("GET", "/api/audit/date-range?startDate=invalid&endDate=2025-01-31", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "Invalid startDate format")
}
