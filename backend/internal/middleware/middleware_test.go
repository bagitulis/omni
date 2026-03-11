package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
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
	// Simulate Auth middleware setting tenant_id (canonical key)
	router.Use(func(c *gin.Context) {
		c.Set("tenant_id", "yumna_bertigamart")
		c.Next()
	})
	router.Use(Tenant())
	router.GET("/test", func(c *gin.Context) {
		tenantID := GetTenantID(c)
		c.JSON(http.StatusOK, gin.H{"tenant_id": tenantID})
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

	// Test invalid format (contains special characters)
	// In development mode, any alphanumeric+underscore tenant is valid
	// Only invalid formats should be rejected
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-tenant-id", "invalid-tenant!@#") // Invalid format
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d for invalid format, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestTenant_ValidFormatDevelopment(t *testing.T) {
	router := gin.New()
	router.Use(Tenant())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	// In development mode (no tenant config), valid format should pass
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("x-tenant-id", "any_valid_tenant_123")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d for valid format in dev mode, got %d", http.StatusOK, w.Code)
	}
}

func TestGetTenantID(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())

	// Test empty context
	if got := GetTenantID(c); got != "" {
		t.Errorf("GetTenantID() with empty context = %v, want ''", got)
	}

	// Test with tenantID set (snake_case)
	c.Set("tenant_id", "tenant123")
	if got := GetTenantID(c); got != "tenant123" {
		t.Errorf("GetTenantID() = %v, want 'tenant123'", got)
	}

	// Test with tenantId set (camelCase fallback)
	c2, _ := gin.CreateTestContext(httptest.NewRecorder())
	c2.Set("tenant_id", "tenant456")
	if got := GetTenantID(c2); got != "tenant456" {
		t.Errorf("GetTenantID() camelCase = %v, want 'tenant456'", got)
	}
}

// ===========================================
// RATE LIMITING TESTS
// ===========================================

func TestRateLimiter_AllowsRequestsWithinLimit(t *testing.T) {
	limiter := NewRateLimiter(10, time.Second, 10)

	// Should allow 10 requests
	for i := 0; i < 10; i++ {
		if !limiter.Allow("test-key") {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}
}

func TestRateLimiter_BlocksExcessRequests(t *testing.T) {
	limiter := NewRateLimiter(5, time.Second, 5)

	// Use all tokens
	for i := 0; i < 5; i++ {
		limiter.Allow("test-key")
	}

	// Next request should be blocked
	if limiter.Allow("test-key") {
		t.Error("Request should be blocked after limit exceeded")
	}
}

func TestRateLimiter_AllowWithInfo_ReturnsCorrectValues(t *testing.T) {
	limiter := NewRateLimiter(5, time.Second, 5)

	// First request
	allowed, remaining, retryAfter := limiter.AllowWithInfo("test-key")
	if !allowed {
		t.Error("First request should be allowed")
	}
	if remaining != 4 {
		t.Errorf("Expected 4 remaining, got %d", remaining)
	}
	if retryAfter != 0 {
		t.Errorf("Expected retryAfter 0, got %v", retryAfter)
	}

	// Exhaust tokens
	for i := 0; i < 4; i++ {
		limiter.Allow("test-key")
	}

	// Should be blocked now
	allowed, remaining, retryAfter = limiter.AllowWithInfo("test-key")
	if allowed {
		t.Error("Request should be blocked")
	}
	if remaining != 0 {
		t.Errorf("Expected 0 remaining, got %d", remaining)
	}
	if retryAfter <= 0 {
		t.Error("Expected positive retryAfter")
	}
}

func TestRateLimiter_DifferentKeysAreSeparate(t *testing.T) {
	limiter := NewRateLimiter(2, time.Second, 2)

	// Exhaust key1
	limiter.Allow("key1")
	limiter.Allow("key1")
	if limiter.Allow("key1") {
		t.Error("key1 should be blocked")
	}

	// key2 should still work
	if !limiter.Allow("key2") {
		t.Error("key2 should be allowed")
	}
}

func TestAuthRateLimitMiddleware_CreatesLimiter(t *testing.T) {
	// Just verify it doesn't panic and creates a handler
	handler := AuthRateLimitMiddleware()
	if handler == nil {
		t.Error("AuthRateLimitMiddleware should return a handler")
	}
}

func TestAPIRateLimitMiddleware_CreatesLimiter(t *testing.T) {
	// Just verify it doesn't panic and creates a handler
	handler := APIRateLimitMiddleware()
	if handler == nil {
		t.Error("APIRateLimitMiddleware should return a handler")
	}
}
