package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimitMiddlewareWithHeaders_SetsHeadersOnAllowedRequest(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute, 2)
	router := gin.New()
	router.Use(RateLimitMiddlewareWithHeaders(limiter, func(_ *gin.Context) string {
		return "test-key"
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if got := w.Header().Get("X-RateLimit-Limit"); got != "2" {
		t.Errorf("expected X-RateLimit-Limit 2, got %q", got)
	}

	if got := w.Header().Get("X-RateLimit-Remaining"); got != "1" {
		t.Errorf("expected X-RateLimit-Remaining 1, got %q", got)
	}
}

func TestRateLimitMiddlewareWithHeaders_SetsRetryAfterOnBlockedRequest(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute, 1)
	router := gin.New()
	router.Use(RateLimitMiddlewareWithHeaders(limiter, func(_ *gin.Context) string {
		return "test-key"
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	firstReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	firstResp := httptest.NewRecorder()
	router.ServeHTTP(firstResp, firstReq)

	if firstResp.Code != http.StatusOK {
		t.Fatalf("expected first request status %d, got %d", http.StatusOK, firstResp.Code)
	}

	secondReq := httptest.NewRequest(http.MethodGet, "/test", nil)
	secondResp := httptest.NewRecorder()
	router.ServeHTTP(secondResp, secondReq)

	if secondResp.Code != http.StatusTooManyRequests {
		t.Fatalf("expected second request status %d, got %d", http.StatusTooManyRequests, secondResp.Code)
	}

	retryAfterRaw := secondResp.Header().Get("Retry-After")
	if retryAfterRaw == "" {
		t.Fatal("expected Retry-After header to be set")
	}

	retryAfterSeconds, err := strconv.Atoi(retryAfterRaw)
	if err != nil {
		t.Fatalf("expected Retry-After to be an integer, got %q", retryAfterRaw)
	}

	if retryAfterSeconds <= 0 {
		t.Errorf("expected Retry-After > 0, got %d", retryAfterSeconds)
	}

	var body map[string]any
	if err := json.Unmarshal(secondResp.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	retryAfterBody, ok := body["retry_after"].(float64)
	if !ok {
		t.Fatalf("expected retry_after number in response body, got %v", body["retry_after"])
	}

	if int(retryAfterBody) != retryAfterSeconds {
		t.Errorf("expected retry_after body value %d, got %d", retryAfterSeconds, int(retryAfterBody))
	}
}

func TestRateLimitMiddlewareWithHeaders_UsesDifferentKeysIndependently(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute, 1)
	router := gin.New()
	router.Use(RateLimitMiddlewareWithHeaders(limiter, func(c *gin.Context) string {
		return c.GetHeader("X-Test-Key")
	}))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})

	requestA1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	requestA1.Header.Set("X-Test-Key", "A")
	responseA1 := httptest.NewRecorder()
	router.ServeHTTP(responseA1, requestA1)

	if responseA1.Code != http.StatusOK {
		t.Fatalf("expected key A first request status %d, got %d", http.StatusOK, responseA1.Code)
	}

	requestA2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	requestA2.Header.Set("X-Test-Key", "A")
	responseA2 := httptest.NewRecorder()
	router.ServeHTTP(responseA2, requestA2)

	if responseA2.Code != http.StatusTooManyRequests {
		t.Fatalf("expected key A second request status %d, got %d", http.StatusTooManyRequests, responseA2.Code)
	}

	requestB1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	requestB1.Header.Set("X-Test-Key", "B")
	responseB1 := httptest.NewRecorder()
	router.ServeHTTP(responseB1, requestB1)

	if responseB1.Code != http.StatusOK {
		t.Fatalf("expected key B first request status %d, got %d", http.StatusOK, responseB1.Code)
	}
}
