package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCORS_AllowsDefaultWhitelistedOrigin(t *testing.T) {
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	origin := "http://localhost:5173"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("expected Access-Control-Allow-Origin %q, got %q", origin, got)
	}

	if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("expected Access-Control-Allow-Credentials true, got %q", got)
	}

	if got := w.Header().Get("X-Content-Type-Options"); got != "nosniff" {
		t.Errorf("expected X-Content-Type-Options nosniff, got %q", got)
	}

	if got := w.Header().Get("X-Frame-Options"); got != "SAMEORIGIN" {
		t.Errorf("expected X-Frame-Options SAMEORIGIN, got %q", got)
	}

	if got := w.Header().Get("Referrer-Policy"); got != "strict-origin-when-cross-origin" {
		t.Errorf("expected Referrer-Policy strict-origin-when-cross-origin, got %q", got)
	}
}

func TestCORS_DoesNotAllowUnlistedOrigin(t *testing.T) {
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://attacker.example")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin for unlisted origin, got %q", got)
	}
}

func TestCORS_AllowsOriginFromEnvironmentVariable(t *testing.T) {
	t.Setenv("CORS_ORIGINS", "https://app.example.com, https://portal.example.com")

	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	origin := "https://app.example.com"
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", origin)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if got := w.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("expected Access-Control-Allow-Origin %q, got %q", origin, got)
	}
}

func TestCORS_PreflightOptionsReturnsNoContent(t *testing.T) {
	router := gin.New()
	router.Use(CORS())
	router.OPTIONS("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}

func TestCORS_NeverUsesWildcardOrigin(t *testing.T) {
	router := gin.New()
	router.Use(CORS())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if got := w.Header().Get("Access-Control-Allow-Origin"); got == "*" {
		t.Errorf("expected explicit origin, got wildcard %q", got)
	}
}
