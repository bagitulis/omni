// Package e2e contains end-to-end performance tests
package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"
)

// Test configuration
var (
	baseURL  = getEnv("TEST_BASE_URL", "http://localhost:3000")
	apiBase  = baseURL + "/api"
	testUser = getEnv("TEST_USERNAME", "tester")
	testPass = getEnv("TEST_PASSWORD", "tester@123")
)

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

// PageScore represents performance metrics for a page
type PageScore struct {
	Name          string        `json:"name"`
	URL           string        `json:"url"`
	StatusCode    int           `json:"status_code"`
	ResponseTime  time.Duration `json:"response_time_ms"`
	ContentLength int64         `json:"content_length"`
	Success       bool          `json:"success"`
	Error         string        `json:"error,omitempty"`
}

// LighthouseResult represents overall test results
type LighthouseResult struct {
	Timestamp     string      `json:"timestamp"`
	TotalTests    int         `json:"total_tests"`
	Passed        int         `json:"passed"`
	Failed        int         `json:"failed"`
	TotalDuration int64       `json:"total_duration_ms"`
	AvgResponse   int64       `json:"avg_response_ms"`
	Pages         []PageScore `json:"pages"`
}

// httpClient with timeout
var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// getAuthToken authenticates and returns token
func getAuthToken() (string, error) {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	jsonBody, _ := json.Marshal(loginReq)
	resp, err := httpClient.Post(apiBase+"/auth/login", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result struct {
		Token string `json:"token"`
	}
	json.Unmarshal(body, &result)

	return result.Token, nil
}

// measureEndpoint measures response time for an endpoint
func measureEndpoint(name, url, token string) PageScore {
	score := PageScore{
		Name: name,
		URL:  url,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		score.Error = err.Error()
		return score
	}

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	start := time.Now()
	resp, err := httpClient.Do(req)
	score.ResponseTime = time.Since(start)

	if err != nil {
		score.Error = err.Error()
		return score
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	score.StatusCode = resp.StatusCode
	score.ContentLength = int64(len(body))
	score.Success = resp.StatusCode >= 200 && resp.StatusCode < 400

	return score
}

// ===========================================
// LIGHTHOUSE-STYLE PERFORMANCE TESTS
// ===========================================

func TestLighthouse_AllEndpoints(t *testing.T) {
	token, err := getAuthToken()
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	// Define pages to test
	pages := []struct {
		name      string
		url       string
		needsAuth bool
	}{
		{"Health Check", apiBase + "/health", false},
		{"Orders List", apiBase + "/orders", true},
		{"Products List", apiBase + "/products", true},
		{"Tenants List", apiBase + "/auth/tenants", true},
		{"Token Verify", apiBase + "/auth/verify", true},
		{"Current User", apiBase + "/auth/me", true},
		{"Shopee Orders", apiBase + "/orders/shopee?limit=5", true},
		{"TikTok Orders", apiBase + "/orders/tiktok?limit=5", true},
		{"Lazada Orders", apiBase + "/orders/lazada?limit=5", true},
		{"Inventory", apiBase + "/inventory", true},
	}

	result := LighthouseResult{
		Timestamp:  time.Now().Format(time.RFC3339),
		TotalTests: len(pages),
		Pages:      make([]PageScore, 0),
	}

	var totalDuration time.Duration

	for _, page := range pages {
		authToken := ""
		if page.needsAuth {
			authToken = token
		}

		score := measureEndpoint(page.name, page.url, authToken)
		result.Pages = append(result.Pages, score)
		totalDuration += score.ResponseTime

		if score.Success {
			result.Passed++
			t.Logf("✓ %s: %dms (status: %d)", page.name, score.ResponseTime.Milliseconds(), score.StatusCode)
		} else {
			result.Failed++
			if score.Error != "" {
				t.Logf("✗ %s: %s", page.name, score.Error)
			} else {
				t.Logf("✗ %s: status %d", page.name, score.StatusCode)
			}
		}
	}

	result.TotalDuration = totalDuration.Milliseconds()
	if len(result.Pages) > 0 {
		result.AvgResponse = result.TotalDuration / int64(len(result.Pages))
	}

	// Print summary
	t.Logf("\n=== Performance Summary ===")
	t.Logf("Total Tests: %d", result.TotalTests)
	t.Logf("Passed: %d", result.Passed)
	t.Logf("Failed: %d", result.Failed)
	t.Logf("Total Duration: %dms", result.TotalDuration)
	t.Logf("Avg Response: %dms", result.AvgResponse)

	// Performance thresholds
	if result.AvgResponse > 500 {
		t.Errorf("Average response time too high: %dms (threshold: 500ms)", result.AvgResponse)
	}

	if result.Failed > 0 {
		t.Errorf("%d endpoints failed", result.Failed)
	}
}

func TestLighthouse_HealthEndpoint_Performance(t *testing.T) {
	// Health endpoint should be very fast
	score := measureEndpoint("Health", apiBase+"/health", "")

	if !score.Success {
		skipOrFailf(t, "Health endpoint not available: %v", score.Error)
		return
	}

	// Health should respond in under 100ms
	if score.ResponseTime > 100*time.Millisecond {
		t.Errorf("Health endpoint too slow: %dms (threshold: 100ms)", score.ResponseTime.Milliseconds())
	}

	t.Logf("Health response time: %dms", score.ResponseTime.Milliseconds())
}

func TestLighthouse_LoginEndpoint_Performance(t *testing.T) {
	loginReq := map[string]string{
		"username": testUser,
		"password": testPass,
	}

	jsonBody, _ := json.Marshal(loginReq)

	start := time.Now()
	resp, err := httpClient.Post(apiBase+"/auth/login", "application/json", bytes.NewBuffer(jsonBody))
	duration := time.Since(start)

	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}
	defer resp.Body.Close()

	// Login should respond in under 500ms (includes bcrypt)
	if duration > 500*time.Millisecond {
		t.Errorf("Login endpoint too slow: %dms (threshold: 500ms)", duration.Milliseconds())
	}

	t.Logf("Login response time: %dms", duration.Milliseconds())
}

func TestLighthouse_ConcurrentRequests(t *testing.T) {
	token, err := getAuthToken()
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	// Test concurrent requests
	concurrency := 10
	endpoint := apiBase + "/health"

	results := make(chan PageScore, concurrency)

	start := time.Now()

	for i := 0; i < concurrency; i++ {
		go func() {
			score := measureEndpoint("Concurrent", endpoint, token)
			results <- score
		}()
	}

	// Collect results
	var passed, failed int
	var totalResponseTime time.Duration

	for i := 0; i < concurrency; i++ {
		score := <-results
		if score.Success {
			passed++
		} else {
			failed++
		}
		totalResponseTime += score.ResponseTime
	}

	totalDuration := time.Since(start)
	avgResponse := totalResponseTime / time.Duration(concurrency)

	t.Logf("\n=== Concurrent Request Test ===")
	t.Logf("Concurrency: %d", concurrency)
	t.Logf("Total Duration: %dms", totalDuration.Milliseconds())
	t.Logf("Avg Response: %dms", avgResponse.Milliseconds())
	t.Logf("Passed: %d, Failed: %d", passed, failed)

	if failed > 0 {
		t.Errorf("%d concurrent requests failed", failed)
	}

	// All concurrent requests should complete in under 2 seconds
	if totalDuration > 2*time.Second {
		t.Errorf("Concurrent requests too slow: %dms", totalDuration.Milliseconds())
	}
}

func TestLighthouse_ResponseSize(t *testing.T) {
	token, err := getAuthToken()
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	endpoints := []struct {
		name     string
		url      string
		maxBytes int64
	}{
		{"Health", apiBase + "/health", 1024},                        // 1KB max
		{"Tenants", apiBase + "/auth/tenants", 10240},                // 10KB max
		{"Orders (paginated)", apiBase + "/orders?limit=10", 102400}, // 100KB max
	}

	for _, ep := range endpoints {
		score := measureEndpoint(ep.name, ep.url, token)

		if !score.Success {
			t.Logf("⚠ %s: not available", ep.name)
			continue
		}

		if score.ContentLength > ep.maxBytes {
			t.Errorf("%s response too large: %d bytes (max: %d)", ep.name, score.ContentLength, ep.maxBytes)
		} else {
			t.Logf("✓ %s: %d bytes", ep.name, score.ContentLength)
		}
	}
}

// ===========================================
// API AVAILABILITY TESTS
// ===========================================

func TestAPI_AllEndpointsAvailable(t *testing.T) {
	token, err := getAuthToken()
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	// Critical endpoints that must be available
	endpoints := []struct {
		name   string
		method string
		url    string
	}{
		{"Health", "GET", apiBase + "/health"},
		{"Login", "POST", apiBase + "/auth/login"},
		{"Verify Token", "GET", apiBase + "/auth/verify"},
		{"Get Tenants", "GET", apiBase + "/auth/tenants"},
		{"Get Current User", "GET", apiBase + "/auth/me"},
	}

	for _, ep := range endpoints {
		req, _ := http.NewRequest(ep.method, ep.url, nil)
		if ep.method == "POST" && ep.name == "Login" {
			body, _ := json.Marshal(map[string]string{"username": testUser, "password": testPass})
			req, _ = http.NewRequest(ep.method, ep.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
		} else {
			req.Header.Set("Authorization", "Bearer "+token)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			t.Errorf("✗ %s: connection error - %v", ep.name, err)
			continue
		}
		resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 400 {
			t.Logf("✓ %s: available (status %d)", ep.name, resp.StatusCode)
		} else if resp.StatusCode == 401 || resp.StatusCode == 403 {
			t.Logf("✓ %s: available (auth required)", ep.name)
		} else {
			t.Errorf("✗ %s: unexpected status %d", ep.name, resp.StatusCode)
		}
	}
}

// ===========================================
// PRINT RESULTS AS JSON
// ===========================================

func TestLighthouse_GenerateReport(t *testing.T) {
	token, err := getAuthToken()
	if err != nil {
		skipOrFailf(t, "Backend not available: %v", err)
		return
	}

	pages := []struct {
		name string
		url  string
	}{
		{"Health Check", apiBase + "/health"},
		{"Login", apiBase + "/auth/login"},
		{"Tenants", apiBase + "/auth/tenants"},
		{"Orders", apiBase + "/orders"},
		{"Products", apiBase + "/products"},
	}

	result := LighthouseResult{
		Timestamp:  time.Now().Format(time.RFC3339),
		TotalTests: len(pages),
		Pages:      make([]PageScore, 0),
	}

	for _, page := range pages {
		score := measureEndpoint(page.name, page.url, token)
		result.Pages = append(result.Pages, score)

		if score.Success {
			result.Passed++
		} else {
			result.Failed++
		}
		result.TotalDuration += score.ResponseTime.Milliseconds()
	}

	if len(result.Pages) > 0 {
		result.AvgResponse = result.TotalDuration / int64(len(result.Pages))
	}

	// Print JSON report
	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("\n%s\n", string(jsonResult))
}
