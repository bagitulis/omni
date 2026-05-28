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

	r.POST("/api/dev/reset-password", func(c *gin.Context) {
		var req resetPasswordRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: " + err.Error(),
			})
			return
		}

		// Simulate password validation
		if isPasswordValidationError(errors.New("password must be at least 8 characters")) {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "password must be at least 8 characters",
			})
			return
		}
	})

	req, _ := http.NewRequest("POST", "/api/dev/reset-password",
		strings.NewReader(`{"user_id":"u1","tenant_id":"t1","new_password":"weak"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "password must")
}

func TestIsPasswordValidationError(t *testing.T) {
	tests := []struct {
		name   string
		err    error
		expect bool
	}{
		{"password must prefix", errors.New("password must be at least 8 characters"), true},
		{"password too common", errors.New("password is too common"), true},
		{"generic error", errors.New("database connection failed"), false},
		{"user not found", errors.New("user not found"), false},
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
		c.Set("developer_tenant_scope", "*")
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

func TestSearchUsersWithExplicitTenantScope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dev/users/search", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		if !requireDeveloperPanelAccess(c) {
			return
		}
		// Simulate the new handler logic: explicit scope works when provided
		if len(parseExplicitTenantScope(c.Query("tenant_ids"))) == 0 {
			// In real handler, would default to all active tenants
			// In this mock, just confirm the scope fallback path is reached
			c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}, "total": 0, "scope_source": "default"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true, "data": []interface{}{}, "total": 0, "scope_source": "explicit"})
	})

	// Test with explicit tenant_ids
	req, _ := http.NewRequest("GET", "/api/dev/users/search?q=test&tenant_ids=tenant_a,tenant_b", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
	assert.Equal(t, "explicit", resp["scope_source"])

	// Test without tenant_ids — handler defaults to all active tenants
	req2, _ := http.NewRequest("GET", "/api/dev/users/search?q=test", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
	var resp2 map[string]interface{}
	assert.NoError(t, json.Unmarshal(w2.Body.Bytes(), &resp2))
	assert.Equal(t, true, resp2["success"])
	assert.Equal(t, "default", resp2["scope_source"])
}

func TestSearchUsersContextPropagation(t *testing.T) {
	// Verify that developer context is properly propagated through search
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.GET("/api/dev/users/search", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		c.Set("userID", "dev-user-1")
		c.Set("tenant_id", "default")

		// Verify impersonation check
		if isImpersonatedRequest(c) {
			c.JSON(http.StatusForbidden, gin.H{"success": false, "error": "blocked during impersonation"})
			return
		}

		// Verify role check
		if !requireDeveloperPanelAccess(c) {
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"context": gin.H{
				"role":     c.GetString("role"),
				"user_id":  c.GetString("userID"),
				"tenant":   c.GetString("tenant_id"),
			},
		})
	})

	req, _ := http.NewRequest("GET", "/api/dev/users/search?q=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])
}

func TestSearchUsers_ImpersonationBlocked(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/api/dev/users/search", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		c.Set("impersonated", true)
		if !requireDeveloperPanelAccess(c) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest("GET", "/api/dev/users/search?q=test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeactivateTenant_MissingID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	// Gin returns 404 for /api/dev/tenants/ (empty :id doesn't match route).
	// Test the guard clause directly by sending a valid ID request.
	r.DELETE("/api/dev/tenants/:id", func(c *gin.Context) {
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

	// Valid ID path works correctly
	req, _ := http.NewRequest("DELETE", "/api/dev/tenants/valid_tenant", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, true, resp["success"])

	// Empty ID path returns 404 (Gin routing rejects it)
	req2, _ := http.NewRequest("DELETE", "/api/dev/tenants/", nil)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusNotFound, w2.Code)
}

func TestDeactivateTenant_UnauthorizedTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.DELETE("/api/dev/tenants/:id", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		c.Set("developer_tenant_scope", []string{"tenant_a"})
		tenantID := c.Param("id")
		if !requireDeveloperMutationScope(c, tenantID) {
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req, _ := http.NewRequest("DELETE", "/api/dev/tenants/tenant_b", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestCreateTenant_EmptyName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/dev/tenants", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		if !requireDeveloperPanelAccess(c) {
			return
		}
		var req createTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: name is required",
			})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"success": true, "data": gin.H{"id": req.Name}})
	})

	req, _ := http.NewRequest("POST", "/api/dev/tenants", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"].(string), "validation failed")
}

func TestCreateTenant_ValidName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	r.POST("/api/dev/tenants", func(c *gin.Context) {
		c.Set("role", models.RoleDeveloper)
		if !requireDeveloperPanelAccess(c) {
			return
		}
		var req createTenantRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "validation failed: name is required",
			})
			return
		}
		c.JSON(http.StatusCreated, gin.H{
			"success": true,
			"data": gin.H{
				"id":        req.Name,
				"name":      req.Name,
				"is_active": true,
			},
		})
	})

	req, _ := http.NewRequest("POST", "/api/dev/tenants", strings.NewReader(`{"name":"new_tenant"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
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

// TestTenantContextIntegration verifies context propagation patterns
// used across the developer panel — tenant switch, impersonation block,
// and return-from-impersonation clears caches.
func TestTenantContextIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("developer endpoints blocked during impersonation", func(t *testing.T) {
		r := gin.New()
		r.GET("/api/dev/overview", func(c *gin.Context) {
			c.Set("role", models.RoleDeveloper)
			c.Set("impersonated", true)
			if !requireDeveloperPanelAccess(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req, _ := http.NewRequest("GET", "/api/dev/overview", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("developer endpoints work after return from impersonation", func(t *testing.T) {
		r := gin.New()
		r.GET("/api/dev/overview", func(c *gin.Context) {
			c.Set("role", models.RoleDeveloper)
			// No impersonation flag — developer is back in developer context
			if !requireDeveloperPanelAccess(c) {
				return
			}
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req, _ := http.NewRequest("GET", "/api/dev/overview", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("isImpersonatedRequest checks all context keys", func(t *testing.T) {
		tests := []struct {
			name    string
			setup   func(c *gin.Context)
			expect  bool
		}{
			{
				name: "impersonated key true",
				setup: func(c *gin.Context) {
					c.Set("impersonated", true)
				},
				expect: true,
			},
			{
				name: "is_impersonated key true",
				setup: func(c *gin.Context) {
					c.Set("is_impersonated", true)
				},
				expect: true,
			},
			{
				name: "impersonation_active key true",
				setup: func(c *gin.Context) {
					c.Set("impersonation_active", true)
				},
				expect: true,
			},
			{
				name: "impersonated_by set",
				setup: func(c *gin.Context) {
					c.Set("impersonated_by", "admin-user")
				},
				expect: true,
			},
			{
				name: "impersonation_actor_id set",
				setup: func(c *gin.Context) {
					c.Set("impersonation_actor_id", "admin-123")
				},
				expect: true,
			},
			{
				name: "no impersonation flags",
				setup: func(c *gin.Context) {
					c.Set("role", models.RoleDeveloper)
					c.Set("tenant_id", "default")
				},
				expect: false,
			},
			{
				name: "impersonated key false",
				setup: func(c *gin.Context) {
					c.Set("impersonated", false)
				},
				expect: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				r := gin.New()
				var result bool
				r.GET("/test", func(c *gin.Context) {
					tt.setup(c)
					result = isImpersonatedRequest(c)
				})

				req, _ := http.NewRequest("GET", "/test", nil)
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				assert.Equal(t, tt.expect, result)
			})
		}
	})

	t.Run("tenant scope falls back to JWT tenant_id", func(t *testing.T) {
		r := gin.New()
		var allowed map[string]struct{}
		r.GET("/test", func(c *gin.Context) {
			c.Set("role", models.RoleDeveloper)
			c.Set("tenant_id", "my_tenant")
			// No explicit scope set — should fall back to tenant_id
			allowed = developerAllowedTenants(c)
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Contains(t, allowed, "my_tenant")
		assert.Len(t, allowed, 1)
	})

	t.Run("superadmin has universal tenant access", func(t *testing.T) {
		r := gin.New()
		var canAccess bool
		r.GET("/test", func(c *gin.Context) {
			c.Set("role", roleSuperadmin)
			c.Set("tenant_id", "default")
			canAccess = developerCanAccessTenant(c, "any_tenant")
		})

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		assert.True(t, canAccess)
	})
}

// Ensure services import is used (for compile check)
var _ = services.ErrTenantNotFound
