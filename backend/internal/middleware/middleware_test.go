package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuth_MissingAuthHeader(t *testing.T) {
	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuth_InvalidAuthFormat(t *testing.T) {
	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "invalid-format")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestAuth_InvalidToken(t *testing.T) {
	router := gin.New()
	router.Use(Auth())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetUserID(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test empty context
	if got := GetUserID(c); got != "" {
		t.Errorf("GetUserID() with empty context = %v, want ''", got)
	}

	// Test with userID set
	c.Set("userID", "user123")
	if got := GetUserID(c); got != "user123" {
		t.Errorf("GetUserID() = %v, want 'user123'", got)
	}
}

func TestGetRole(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test empty context
	if got := GetRole(c); got != "" {
		t.Errorf("GetRole() with empty context = %v, want ''", got)
	}

	// Test with role set
	c.Set("role", "admin")
	if got := GetRole(c); got != "admin" {
		t.Errorf("GetRole() = %v, want 'admin'", got)
	}
}

func TestTenant_MissingTenantID(t *testing.T) {
	router := gin.New()
	router.Use(Tenant())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTenant_ValidTenantFromHeader(t *testing.T) {
	router := gin.New()
	router.Use(Tenant())
	router.GET("/test", func(c *gin.Context) {
		tenantID := GetTenantID(c)
		c.JSON(http.StatusOK, gin.H{"tenantID": tenantID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-tenant-id", "yumna_bertigamart")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTenant_ValidTenantFromContext(t *testing.T) {
	router := gin.New()
	// Simulate Auth middleware setting tenantID
	router.Use(func(c *gin.Context) {
		c.Set("tenantID", "yumna_bertigamart")
		c.Next()
	})
	router.Use(Tenant())
	router.GET("/test", func(c *gin.Context) {
		tenantID := GetTenantID(c)
		c.JSON(http.StatusOK, gin.H{"tenantID": tenantID})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTenant_InvalidTenant(t *testing.T) {
	router := gin.New()
	router.Use(Tenant())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-tenant-id", "invalid_tenant")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestGetTenantID(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test empty context
	if got := GetTenantID(c); got != "" {
		t.Errorf("GetTenantID() with empty context = %v, want ''", got)
	}

	// Test with tenantID set (snake_case)
	c.Set("tenantID", "tenant123")
	if got := GetTenantID(c); got != "tenant123" {
		t.Errorf("GetTenantID() = %v, want 'tenant123'", got)
	}

	// Test with tenantId set (camelCase fallback)
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Set("tenantId", "tenant456")
	if got := GetTenantID(c2); got != "tenant456" {
		t.Errorf("GetTenantID() camelCase = %v, want 'tenant456'", got)
	}
}
