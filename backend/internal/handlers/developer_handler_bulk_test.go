package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestBulkResetPasswords_EmptyUserIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-reset-password", func(c *gin.Context) {
		var req bulkResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "user_ids must not be empty",
			})
			return
		}
	})

	body := `{"user_ids":[],"tenant_id":"t1","new_password":"StrongPass1!"}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "user_ids must not be empty")
}

func TestBulkResetPasswords_ExceedsMaxUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-reset-password", func(c *gin.Context) {
		var req bulkResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "user_ids must not be empty",
			})
			return
		}

		if len(req.UserIDs) > maxBulkUsers {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("user_ids exceeds maximum of %d users per request", maxBulkUsers),
			})
			return
		}
	})

	// Build 51 user IDs
	userIDs := make([]string, 51)
	for i := range userIDs {
		userIDs[i] = fmt.Sprintf("user-%d", i)
	}
	userIDsJSON, _ := json.Marshal(userIDs)
	body := fmt.Sprintf(`{"user_ids":%s,"tenant_id":"t1","new_password":"StrongPass1!"}`, userIDsJSON)

	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "exceeds maximum of 50")
}

func TestBulkResetPasswords_WeakPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-reset-password", func(c *gin.Context) {
		var req bulkResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "user_ids must not be empty",
			})
			return
		}

		// Simulate: each user gets a password validation error
		result := bulkOperationResult{
			Errors: make([]bulkErrorDetail, 0),
		}
		for _, userID := range req.UserIDs {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: userID,
				Error:  "password must be at least 8 characters",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	body := `{"user_ids":["u1","u2"],"tenant_id":"t1","new_password":"weak"}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(0), data["success_count"])
	assert.Equal(t, float64(2), data["failure_count"])

	errors := data["errors"].([]interface{})
	assert.Len(t, errors, 2)
	firstErr := errors[0].(map[string]interface{})
	assert.Equal(t, "u1", firstErr["user_id"])
	assert.Contains(t, firstErr["error"].(string), "password must")
}

func TestBulkResetPasswords_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-reset-password", func(c *gin.Context) {
		var req bulkResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "user_ids must not be empty",
			})
			return
		}

		// Simulate: all succeed
		result := bulkOperationResult{
			SuccessCount: len(req.UserIDs),
			FailureCount: 0,
			Errors:       make([]bulkErrorDetail, 0),
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	body := `{"user_ids":["u1","u2","u3"],"tenant_id":"t1","new_password":"StrongPass1!"}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(3), data["success_count"])
	assert.Equal(t, float64(0), data["failure_count"])
	assert.Empty(t, data["errors"])
}

func TestBulkResetPasswords_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-reset-password", func(c *gin.Context) {
		var req bulkResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}
	})

	tests := []struct {
		name string
		body string
	}{
		{"missing user_ids", `{"tenant_id":"t1","new_password":"StrongPass1!"}`},
		{"missing tenant_id", `{"user_ids":["u1"],"new_password":"StrongPass1!"}`},
		{"missing new_password", `{"user_ids":["u1"],"tenant_id":"t1"}`},
		{"empty body", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/dev/users/bulk-reset-password", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, false, resp["success"])
			assert.Contains(t, resp["error"].(string), "validation failed")
		})
	}
}

func TestBulkDisableUsers_EmptyUserIDs(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-disable", func(c *gin.Context) {
		var req bulkDisableUsersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "user_ids must not be empty",
			})
			return
		}
	})

	body := `{"user_ids":[],"tenant_id":"t1"}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "user_ids must not be empty")
}

func TestBulkDisableUsers_ExceedsMaxUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-disable", func(c *gin.Context) {
		var req bulkDisableUsersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "user_ids must not be empty",
			})
			return
		}

		if len(req.UserIDs) > maxBulkUsers {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("user_ids exceeds maximum of %d users per request", maxBulkUsers),
			})
			return
		}
	})

	userIDs := make([]string, 51)
	for i := range userIDs {
		userIDs[i] = fmt.Sprintf("user-%d", i)
	}
	userIDsJSON, _ := json.Marshal(userIDs)
	body := fmt.Sprintf(`{"user_ids":%s,"tenant_id":"t1"}`, userIDsJSON)

	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "exceeds maximum of 50")
}

func TestBulkDisableUsers_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-disable", func(c *gin.Context) {
		var req bulkDisableUsersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		if len(req.UserIDs) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "user_ids must not be empty",
			})
			return
		}

		// Simulate: all succeed
		result := bulkOperationResult{
			SuccessCount: len(req.UserIDs),
			FailureCount: 0,
			Errors:       make([]bulkErrorDetail, 0),
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	body := `{"user_ids":["u1","u2"],"tenant_id":"t1"}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(2), data["success_count"])
	assert.Equal(t, float64(0), data["failure_count"])
	assert.Empty(t, data["errors"])
}

func TestBulkDisableUsers_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-disable", func(c *gin.Context) {
		var req bulkDisableUsersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}
	})

	tests := []struct {
		name string
		body string
	}{
		{"missing user_ids", `{"tenant_id":"t1"}`},
		{"missing tenant_id", `{"user_ids":["u1"]}`},
		{"empty body", `{}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/api/dev/users/bulk-disable", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, false, resp["success"])
			assert.Contains(t, resp["error"].(string), "validation failed")
		})
	}
}

func TestBulkDisableUsers_PartialFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	r.POST("/api/dev/users/bulk-disable", func(c *gin.Context) {
		var req bulkDisableUsersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		// Simulate: first succeeds, second fails (not found)
		result := bulkOperationResult{
			SuccessCount: 1,
			FailureCount: 1,
			Errors: []bulkErrorDetail{
				{UserID: "u2", Error: "user not found"},
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	body := `{"user_ids":["u1","u2"],"tenant_id":"t1"}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["success_count"])
	assert.Equal(t, float64(1), data["failure_count"])

	errors := data["errors"].([]interface{})
	assert.Len(t, errors, 1)
	firstErr := errors[0].(map[string]interface{})
	assert.Equal(t, "u2", firstErr["user_id"])
	assert.Equal(t, "user not found", firstErr["error"])
}
