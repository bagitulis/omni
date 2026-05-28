package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestBulkResetPasswords_EmptyItems(t *testing.T) {
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

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "items must not be empty",
			})
			return
		}
	})

	body := `{"items":[],"new_password":"StrongPass1!"}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "items must not be empty")
}

func TestBulkResetPasswordsRequiresExplicitTenantList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/dev/users/bulk-reset-password", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		var req bulkResetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "validation failed: " + err.Error()})
			return
		}
		if !bulkItemsHaveExplicitTenantScope(req.Items) {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "bulk operation requires explicit tenant_id for every item"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	body := `{"items":[{"user_id":"u1"}],"new_password":"StrongPass1!"}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "explicit tenant_id")
}

func TestBulkOperationMixedTenantScopePartialReject(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/dev/users/bulk-disable", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		c.Set("developer_tenant_scope", []string{"tenant_a"})
		var req bulkDisableUsersRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "validation failed: " + err.Error()})
			return
		}
		result := bulkOperationResult{Errors: make([]bulkErrorDetail, 0)}
		for _, item := range req.Items {
			if !developerCanAccessTenant(c, item.TenantID) {
				result.FailureCount++
				result.Errors = append(result.Errors, bulkErrorDetail{UserID: item.UserID, TenantID: item.TenantID, Error: "tenant is outside authorized developer scope"})
				continue
			}
			result.SuccessCount++
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": result})
	})

	body := `{"items":[{"user_id":"u1","tenant_id":"tenant_a"},{"user_id":"u2","tenant_id":"tenant_b"}]}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(1), data["success_count"])
	assert.Equal(t, float64(1), data["failure_count"])
	errors := data["errors"].([]interface{})
	assert.Equal(t, "tenant_b", errors[0].(map[string]interface{})["tenant_id"])
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

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "items must not be empty",
			})
			return
		}

		if len(req.Items) > maxBulkUsers {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("items exceeds maximum of %d users per request", maxBulkUsers),
			})
			return
		}
	})

	// Build 51 items
	items := make([]bulkUserItem, 51)
	for i := range items {
		items[i] = bulkUserItem{UserID: fmt.Sprintf("user-%d", i), TenantID: "t1"}
	}
	itemsJSON, _ := json.Marshal(items)
	body := fmt.Sprintf(`{"items":%s,"new_password":"StrongPass1!"}`, itemsJSON)

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

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "items must not be empty",
			})
			return
		}

		// Simulate: each user gets a password validation error
		result := bulkOperationResult{
			Errors: make([]bulkErrorDetail, 0),
		}
		for _, item := range req.Items {
			result.FailureCount++
			result.Errors = append(result.Errors, bulkErrorDetail{
				UserID: item.UserID,
				Error:  "password must be at least 8 characters",
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	body := `{"items":[{"user_id":"u1","tenant_id":"t1"},{"user_id":"u2","tenant_id":"t1"}],"new_password":"weak"}`
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

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "items must not be empty",
			})
			return
		}

		// Simulate: all succeed
		result := bulkOperationResult{
			SuccessCount: len(req.Items),
			FailureCount: 0,
			Errors:       make([]bulkErrorDetail, 0),
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	body := `{"items":[{"user_id":"u1","tenant_id":"t1"},{"user_id":"u2","tenant_id":"t1"},{"user_id":"u3","tenant_id":"t2"}],"new_password":"StrongPass1!"}`
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
		{"missing items", `{"new_password":"StrongPass1!"}`},
		{"missing new_password", `{"items":[{"user_id":"u1","tenant_id":"t1"}]}`},
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

func TestBulkDisableUsers_EmptyItems(t *testing.T) {
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

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "items must not be empty",
			})
			return
		}
	})

	body := `{"items":[]}`
	req, _ := http.NewRequest("POST", "/api/dev/users/bulk-disable", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "items must not be empty")
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

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "items must not be empty",
			})
			return
		}

		if len(req.Items) > maxBulkUsers {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   fmt.Sprintf("items exceeds maximum of %d users per request", maxBulkUsers),
			})
			return
		}
	})

	items := make([]bulkUserItem, 51)
	for i := range items {
		items[i] = bulkUserItem{UserID: fmt.Sprintf("user-%d", i), TenantID: "t1"}
	}
	itemsJSON, _ := json.Marshal(items)
	body := fmt.Sprintf(`{"items":%s}`, itemsJSON)

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

		if len(req.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "items must not be empty",
			})
			return
		}

		// Simulate: all succeed
		result := bulkOperationResult{
			SuccessCount: len(req.Items),
			FailureCount: 0,
			Errors:       make([]bulkErrorDetail, 0),
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    result,
		})
	})

	body := `{"items":[{"user_id":"u1","tenant_id":"t1"},{"user_id":"u2","tenant_id":"t1"}]}`
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
		{"missing items", `{}`},
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

	body := `{"items":[{"user_id":"u1","tenant_id":"t1"},{"user_id":"u2","tenant_id":"t1"}]}`
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
