package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func resetCSRFTokenStore() {
	tokenStore.mu.Lock()
	tokenStore.tokens = make(map[string]time.Time)
	tokenStore.mu.Unlock()
}

func TestGenerateCSRFToken_GeneratesUniqueValidToken(t *testing.T) {
	resetCSRFTokenStore()

	tokenA, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken() error = %v", err)
	}

	tokenB, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken() second call error = %v", err)
	}

	if tokenA == "" || tokenB == "" {
		t.Fatal("GenerateCSRFToken() returned empty token")
	}

	if tokenA == tokenB {
		t.Fatal("GenerateCSRFToken() should generate unique tokens")
	}

	if !ValidateCSRFToken(tokenA) {
		t.Fatal("generated token should be valid")
	}
}

func TestValidateCSRFToken_ReturnsFalseForUnknownToken(t *testing.T) {
	resetCSRFTokenStore()

	if ValidateCSRFToken("unknown-token") {
		t.Fatal("unknown token should be invalid")
	}
}

func TestValidateCSRFToken_ReturnsFalseForExpiredToken(t *testing.T) {
	resetCSRFTokenStore()

	tokenStore.mu.Lock()
	tokenStore.tokens["expired-token"] = time.Now().Add(-1 * time.Minute)
	tokenStore.mu.Unlock()

	if ValidateCSRFToken("expired-token") {
		t.Fatal("expired token should be invalid")
	}
}

func TestCSRFProtection_AllowsSafeMethodWithoutToken(t *testing.T) {
	resetCSRFTokenStore()

	router := gin.New()
	router.Use(CSRFProtection())
	router.GET("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCSRFProtection_AllowsExemptPathWithoutToken(t *testing.T) {
	resetCSRFTokenStore()

	router := gin.New()
	router.Use(CSRFProtection())
	router.POST("/api/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCSRFProtection_RejectsMutatingMethodWithoutToken(t *testing.T) {
	resetCSRFTokenStore()

	router := gin.New()
	router.Use(CSRFProtection())
	router.POST("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body["error"] != "CSRF token missing" {
		t.Errorf("expected error 'CSRF token missing', got %v", body["error"])
	}
}

func TestCSRFProtection_RejectsInvalidToken(t *testing.T) {
	resetCSRFTokenStore()

	router := gin.New()
	router.Use(CSRFProtection())
	router.POST("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set(csrfHeaderName, "invalid-token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestCSRFProtection_RejectsMismatchedCookieToken(t *testing.T) {
	resetCSRFTokenStore()

	validToken, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken() error = %v", err)
	}

	router := gin.New()
	router.Use(CSRFProtection())
	router.POST("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set(csrfHeaderName, validToken)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: "different-token"})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
	}
}

func TestCSRFProtection_AllowsValidMatchingHeaderAndCookie(t *testing.T) {
	resetCSRFTokenStore()

	validToken, err := GenerateCSRFToken()
	if err != nil {
		t.Fatalf("GenerateCSRFToken() error = %v", err)
	}

	router := gin.New()
	router.Use(CSRFProtection())
	router.POST("/api/orders", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodPost, "/api/orders", nil)
	req.Header.Set(csrfHeaderName, validToken)
	req.AddCookie(&http.Cookie{Name: csrfCookieName, Value: validToken})
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestCSRFCleanupLifecycleStartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartCSRFCleanup(ctx)
	StopCSRFCleanup()
	StopCSRFCleanup()

	StartCSRFCleanup(ctx)
	cancel()
	deadline := time.After(200 * time.Millisecond)
	for {
		csrfLifecycleMu.Lock()
		alive := csrfCleanupAlive
		csrfLifecycleMu.Unlock()
		if !alive {
			return
		}
		select {
		case <-deadline:
			t.Fatal("CSRF cleanup worker did not stop after context cancellation")
		case <-time.After(10 * time.Millisecond):
		}
	}
}
