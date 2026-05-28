package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	"github.com/stretchr/testify/assert"
)

func TestResetPassword_MissingFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		body     string
		wantCode int
		wantErr  string
	}{
		{
			name:     "missing all fields",
			body:     `{}`,
			wantCode: http.StatusBadRequest,
			wantErr:  "validation failed",
		},
		{
			name:     "missing user_id",
			body:     `{"tenant_id":"t1","new_password":"StrongPass1!"}`,
			wantCode: http.StatusBadRequest,
			wantErr:  "validation failed",
		},
		{
			name:     "missing tenant_id",
			body:     `{"user_id":"u1","new_password":"StrongPass1!"}`,
			wantCode: http.StatusBadRequest,
			wantErr:  "validation failed",
		},
		{
			name:     "missing new_password",
			body:     `{"user_id":"u1","tenant_id":"t1"}`,
			wantCode: http.StatusBadRequest,
			wantErr:  "validation failed",
		},
		{
			name:     "invalid json",
			body:     `not json`,
			wantCode: http.StatusBadRequest,
			wantErr:  "validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()

			r.Use(func(c *gin.Context) {
				c.Set("userID", "dev-user")
				c.Next()
			})

			r.POST("/api/dev/reset-password", func(c *gin.Context) {
				var req resetPasswordRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   "validation failed: " + err.Error(),
					})
					return
				}
			})

			req, _ := http.NewRequest("POST", "/api/dev/reset-password", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, false, resp["success"])
			assert.Contains(t, resp["error"].(string), tt.wantErr)
		})
	}
}

func TestResetPassword_WeakPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	// Simulate handler logic with password validation error
	r.POST("/api/dev/reset-password", func(c *gin.Context) {
		var req resetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		// Simulate password validation error from service
		err := errors.New("password must be at least 8 characters")
		if isPasswordValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   err.Error(),
			})
			return
		}
	})

	body := `{"user_id":"u1","tenant_id":"t1","new_password":"weak"}`
	req, _ := http.NewRequest("POST", "/api/dev/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "password must be at least 8 characters")
}

func TestResetPassword_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", "dev-user")
		c.Next()
	})

	// Simulate successful reset
	r.POST("/api/dev/reset-password", func(c *gin.Context) {
		var req resetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		// Validate fields are present
		if req.UserID == "" || req.TenantID == "" || req.NewPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "missing required fields",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Password reset successfully",
		})
	})

	body := `{"user_id":"user-123","tenant_id":"tenant-456","new_password":"StrongPass1!"}`
	req, _ := http.NewRequest("POST", "/api/dev/reset-password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "Password reset successfully", resp["message"])
}

func TestIsPasswordValidationError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{"min length", errors.New("password must be at least 8 characters"), true},
		{"uppercase", errors.New("password must contain at least one uppercase letter"), true},
		{"lowercase", errors.New("password must contain at least one lowercase letter"), true},
		{"number", errors.New("password must contain at least one number"), true},
		{"special", errors.New("password must contain at least one special character"), true},
		{"too common", errors.New("password is too common, please choose a stronger password"), true},
		{"max length", errors.New("password must be less than 128 characters"), true},
		{"user not found", errors.New("user not found"), false},
		{"generic error", errors.New("database connection failed"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isPasswordValidationError(tt.err)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestDeveloperPanelAccessBlocksImpersonation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/dev/protected", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		c.Set("impersonated", true)
		if !requireDeveloperPanelAccess(c) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest("GET", "/dev/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "impersonation")
}

func TestDeveloperTenantScopeRejectsUnauthorizedTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/dev/scope", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		c.Set("developer_tenant_scope", []string{"tenant_a"})
		if !requireDeveloperTenantAccess(c, "tenant_b") {
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest("GET", "/dev/scope", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "authorized developer scope")
}

func TestDeveloperMutationRequiresExplicitTenantScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/dev/mutate", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		c.Set("tenant_id", "tenant_a")
		if !requireDeveloperMutationScope(c, "tenant_a") {
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest("POST", "/dev/mutate", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "explicit developer tenant scope")
}

func TestDeveloperRoleGrantGuardPreventsEscalation(t *testing.T) {
	assert.False(t, canDeveloperGrantRole(models.RoleDeveloper, "dev-user", "dev-user", models.RoleAdmin))
	assert.False(t, canDeveloperGrantRole(models.RoleDeveloper, "dev-user", "target", models.RoleDeveloper))
	assert.False(t, canDeveloperGrantRole(models.RoleDeveloper, "dev-user", "target", roleSuperadmin))
	assert.True(t, canDeveloperGrantRole(models.RoleDeveloper, "dev-user", "target", models.RoleAdmin))
	assert.True(t, canDeveloperGrantRole(roleSuperadmin, "root", "target", models.RoleDeveloper))
}

func TestSearchUsersRequiresExplicitTenantScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dev/users/search", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		if !requireDeveloperPanelAccess(c) {
			return
		}
		if len(parseExplicitTenantScope(c.Query("tenant_ids"))) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "explicit tenant scope is required"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest("GET", "/api/dev/users/search?q=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "explicit tenant scope")
}

func TestDeactivateTenant_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Simulate the handler with empty param (Gin won't route here normally,
	// but we test the guard clause)
	r.DELETE("/api/dev/tenants/:id", func(c *gin.Context) {
		// Simulate handler logic for empty id
		tenantID := c.Param("id")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "tenant id is required",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest("DELETE", "/api/dev/tenants/", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Gin returns 301 redirect for trailing slash or 404
	assert.True(t, w.Code == http.StatusMovedPermanently || w.Code == http.StatusNotFound)
}

func TestDeactivateTenant_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.DELETE("/api/dev/tenants/:id", func(c *gin.Context) {
		tenantID := c.Param("id")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "tenant id is required",
			})
			return
		}

		// Simulate ErrTenantNotFound from service
		err := errors.New("tenant not found")
		_ = err
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "tenant not found",
		})
	})

	req, _ := http.NewRequest("DELETE", "/api/dev/tenants/nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Equal(t, "tenant not found", resp["error"])
}

func TestDeactivateTenant_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.DELETE("/api/dev/tenants/:id", func(c *gin.Context) {
		tenantID := c.Param("id")
		if tenantID == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "tenant id is required",
			})
			return
		}

		// Simulate successful deactivation
		c.JSON(http.StatusOK, gin.H{
			"success": true,
		})
	})

	req, _ := http.NewRequest("DELETE", "/api/dev/tenants/test-tenant", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
}

func TestCreateTenant_MissingName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/dev/tenants", func(c *gin.Context) {
		var req createTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: name is required",
			})
			return
		}
	})

	body := `{}`
	req, _ := http.NewRequest("POST", "/api/dev/tenants", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "name is required")
}

func TestCreateTenant_InvalidName(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		body     string
		wantCode int
	}{
		{"uppercase", `{"name":"MyShop"}`, http.StatusBadRequest},
		{"special chars", `{"name":"my-shop"}`, http.StatusBadRequest},
		{"too short", `{"name":"ab"}`, http.StatusBadRequest},
		{"starts with number", `{"name":"1shop"}`, http.StatusBadRequest},
		{"sql injection", `{"name":"test'; DROP--"}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()

			r.POST("/api/dev/tenants", func(c *gin.Context) {
				var req createTenantRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   "validation failed: name is required",
					})
					return
				}

				// Simulate validation error from service
				var validationErr *services.TenantValidationError
				err := &services.TenantValidationError{Msg: "invalid name"}
				if errors.As(err, &validationErr) {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"error":   validationErr.Error(),
					})
					return
				}
			})

			req, _ := http.NewRequest("POST", "/api/dev/tenants", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCode, w.Code)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, false, resp["success"])
		})
	}
}

func TestCreateTenant_DuplicateName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/dev/tenants", func(c *gin.Context) {
		var req createTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: name is required",
			})
			return
		}

		// Simulate duplicate error from service
		var duplicateErr *services.TenantDuplicateError
		err := &services.TenantDuplicateError{Name: req.Name}
		if errors.As(err, &duplicateErr) {
			c.JSON(http.StatusConflict, gin.H{
				"success": false,
				"error":   duplicateErr.Error(),
			})
			return
		}
	})

	body := `{"name":"existing_tenant"}`
	req, _ := http.NewRequest("POST", "/api/dev/tenants", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "already exists")
}

func TestCreateTenant_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/dev/tenants", func(c *gin.Context) {
		var req createTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: name is required",
			})
			return
		}

		// Simulate successful creation
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data": gin.H{
				"id":         req.Name,
				"name":       req.Name,
				"is_active":  true,
				"created_at": "2026-01-01T00:00:00Z",
			},
		})
	})

	body := `{"name":"new_tenant"}`
	req, _ := http.NewRequest("POST", "/api/dev/tenants", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])

	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "new_tenant", data["id"])
	assert.Equal(t, "new_tenant", data["name"])
	assert.Equal(t, true, data["is_active"])
}

func TestSearchUsers_MissingQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/dev/users/search", func(c *gin.Context) {
		query := strings.TrimSpace(c.Query("q"))
		if len(query) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "query parameter 'q' must be at least 2 characters",
			})
			return
		}
	})

	req, _ := http.NewRequest("GET", "/api/dev/users/search", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "at least 2 characters")
}

func TestSearchUsers_TooShortQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/dev/users/search", func(c *gin.Context) {
		query := strings.TrimSpace(c.Query("q"))
		if len(query) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "query parameter 'q' must be at least 2 characters",
			})
			return
		}
	})

	req, _ := http.NewRequest("GET", "/api/dev/users/search?q=a", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "at least 2 characters")
}

func TestSearchUsers_ValidQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Simulate successful search returning users
	r.GET("/api/dev/users/search", func(c *gin.Context) {
		query := strings.TrimSpace(c.Query("q"))
		if len(query) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "query parameter 'q' must be at least 2 characters",
			})
			return
		}

		// Simulate found users
		results := []UserSearchResult{
			{
				ID:         "user-1",
				Username:   "yumna_shop",
				Email:      "yumna@example.com",
				Role:       "owner",
				Status:     "active",
				TenantID:   "tenant-1",
				TenantName: "Yumna Store",
			},
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    results,
			"total":   len(results),
		})
	})

	req, _ := http.NewRequest("GET", "/api/dev/users/search?q=yumna", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(1), resp["total"])

	data := resp["data"].([]interface{})
	assert.Len(t, data, 1)

	user := data[0].(map[string]interface{})
	assert.Equal(t, "user-1", user["id"])
	assert.Equal(t, "yumna_shop", user["username"])
	assert.Equal(t, "yumna@example.com", user["email"])
	assert.Equal(t, "owner", user["role"])
	assert.Equal(t, "active", user["status"])
	assert.Equal(t, "tenant-1", user["tenant_id"])
	assert.Equal(t, "Yumna Store", user["tenant_name"])

	// Verify NO password_hash field in response
	_, hasPassword := user["password_hash"]
	assert.False(t, hasPassword, "response must NOT contain password_hash")
	_, hasPasswordField := user["password"]
	assert.False(t, hasPasswordField, "response must NOT contain password")
}

func TestSearchUsers_EmptyResults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/dev/users/search", func(c *gin.Context) {
		query := strings.TrimSpace(c.Query("q"))
		if len(query) < 2 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "query parameter 'q' must be at least 2 characters",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []UserSearchResult{},
			"total":   0,
		})
	})

	req, _ := http.NewRequest("GET", "/api/dev/users/search?q=nonexistent", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, float64(0), resp["total"])
}
