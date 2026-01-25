// Package parity provides performance comparison tests between Node.js and Go backends
package parity

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"testing"
	"time"
)

// BenchmarkConfig holds configuration for benchmark tests
type BenchmarkConfig struct {
	Iterations int
	Concurrent int
}

// BenchmarkResult holds the result of a benchmark
type BenchmarkResult struct {
	Endpoint     string
	NodeJSTimes  []time.Duration
	GoTimes      []time.Duration
	NodeJSP50    time.Duration
	NodeJSP95    time.Duration
	NodeJSP99    time.Duration
	GoP50        time.Duration
	GoP95        time.Duration
	GoP99        time.Duration
	Speedup      float64
	NodeJSErrors int
	GoErrors     int
}

func percentile(times []time.Duration, p float64) time.Duration {
	if len(times) == 0 {
		return 0
	}
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	idx := int(float64(len(sorted)-1) * p)
	return sorted[idx]
}

func runBenchmark(client *http.Client, url string, headers map[string]string, iterations int) ([]time.Duration, int) {
	times := make([]time.Duration, 0, iterations)
	errors := 0

	for i := 0; i < iterations; i++ {
		_, _, elapsed, err := makeRequest(client, "GET", url, headers)
		if err != nil {
			errors++
			continue
		}
		times = append(times, elapsed)
	}

	return times, errors
}

func calculateResult(endpoint string, nodeTimes, goTimes []time.Duration, nodeErrors, goErrors int) BenchmarkResult {
	result := BenchmarkResult{
		Endpoint:     endpoint,
		NodeJSTimes:  nodeTimes,
		GoTimes:      goTimes,
		NodeJSErrors: nodeErrors,
		GoErrors:     goErrors,
	}

	if len(nodeTimes) > 0 {
		result.NodeJSP50 = percentile(nodeTimes, 0.50)
		result.NodeJSP95 = percentile(nodeTimes, 0.95)
		result.NodeJSP99 = percentile(nodeTimes, 0.99)
	}

	if len(goTimes) > 0 {
		result.GoP50 = percentile(goTimes, 0.50)
		result.GoP95 = percentile(goTimes, 0.95)
		result.GoP99 = percentile(goTimes, 0.99)
	}

	if result.GoP50 > 0 {
		result.Speedup = float64(result.NodeJSP50) / float64(result.GoP50)
	}

	return result
}

// TestPerformanceComparison runs performance tests on key endpoints
func TestPerformanceComparison(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set - skipping performance tests")
	}

	benchCfg := BenchmarkConfig{
		Iterations: 20,
		Concurrent: 1,
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	endpoints := []struct {
		name string
		path string
	}{
		{"Health", "/api/health"},
		{"ShopeeOrders", "/api/shopee/orders?page=1&pageSize=10"},
		{"ShopeeProducts", "/api/shopee/products?page=1&pageSize=10"},
		{"Inventory", "/api/inventory?page=1&pageSize=20"},
		{"LazadaOrders", "/api/lazada/orders?page=1&pageSize=10"},
		{"TiktokOrders", "/api/tiktok/orders?page=1&pageSize=10"},
	}

	t.Logf("Running performance comparison (iterations=%d)", benchCfg.Iterations)
	t.Logf("%-20s | %-12s | %-12s | %-12s | %-12s | Speedup", "Endpoint", "Node P50", "Go P50", "Node P95", "Go P95")
	t.Logf("%s", "--------------------------------------------------------------------------------")

	for _, ep := range endpoints {
		nodeURL := cfg.NodeJSURL + ep.path
		goURL := cfg.GoURL + ep.path

		nodeTimes, nodeErrors := runBenchmark(client, nodeURL, headers, benchCfg.Iterations)
		goTimes, goErrors := runBenchmark(client, goURL, headers, benchCfg.Iterations)

		result := calculateResult(ep.name, nodeTimes, goTimes, nodeErrors, goErrors)

		t.Logf("%-20s | %12v | %12v | %12v | %12v | %.2fx",
			ep.name,
			result.NodeJSP50,
			result.GoP50,
			result.NodeJSP95,
			result.GoP95,
			result.Speedup,
		)

		if result.NodeJSErrors > 0 || result.GoErrors > 0 {
			t.Logf("  Errors: Node.js=%d, Go=%d", result.NodeJSErrors, result.GoErrors)
		}
	}
}

// TestConcurrentPerformance tests performance under concurrent load
func TestConcurrentPerformance(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	concurrencyLevels := []int{5, 10, 20}
	endpoint := "/api/shopee/orders?page=1&pageSize=10"

	for _, concurrency := range concurrencyLevels {
		t.Run(fmt.Sprintf("Concurrency_%d", concurrency), func(t *testing.T) {
			var wg sync.WaitGroup
			nodeResults := make(chan time.Duration, concurrency)
			goResults := make(chan time.Duration, concurrency)

			// Test Node.js
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _, elapsed, err := makeRequest(client, "GET", cfg.NodeJSURL+endpoint, headers)
					if err == nil {
						nodeResults <- elapsed
					}
				}()
			}
			wg.Wait()
			close(nodeResults)

			// Test Go
			for i := 0; i < concurrency; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					_, _, elapsed, err := makeRequest(client, "GET", cfg.GoURL+endpoint, headers)
					if err == nil {
						goResults <- elapsed
					}
				}()
			}
			wg.Wait()
			close(goResults)

			// Collect results
			var nodeTimes, goTimes []time.Duration
			for r := range nodeResults {
				nodeTimes = append(nodeTimes, r)
			}
			for r := range goResults {
				goTimes = append(goTimes, r)
			}

			nodeP50 := percentile(nodeTimes, 0.50)
			goP50 := percentile(goTimes, 0.50)

			speedup := float64(nodeP50) / float64(goP50)
			t.Logf("Concurrency %d: Node.js P50=%v, Go P50=%v (%.2fx faster)",
				concurrency, nodeP50, goP50, speedup)
		})
	}
}

// TestErrorHandlingParity compares error responses between backends
func TestErrorHandlingParity(t *testing.T) {
	cfg := getTestConfig()
	client := &http.Client{Timeout: 10 * time.Second}

	testCases := []struct {
		name           string
		method         string
		path           string
		headers        map[string]string
		expectedStatus int
	}{
		{
			name:           "Unauthorized - No Token",
			method:         "GET",
			path:           "/api/shopee/orders",
			headers:        map[string]string{"x-tenant-id": cfg.TenantID},
			expectedStatus: 401,
		},
		{
			name:           "Missing Tenant ID",
			method:         "GET",
			path:           "/api/shopee/orders",
			headers:        map[string]string{"Authorization": "Bearer " + cfg.AuthToken},
			expectedStatus: 400,
		},
		{
			name:   "Invalid Token",
			method: "GET",
			path:   "/api/shopee/orders",
			headers: map[string]string{
				"Authorization": "Bearer invalid_token_here",
				"x-tenant-id":   cfg.TenantID,
			},
			expectedStatus: 401,
		},
		{
			name:   "Not Found Endpoint",
			method: "GET",
			path:   "/api/nonexistent/endpoint",
			headers: map[string]string{
				"Authorization": "Bearer " + cfg.AuthToken,
				"x-tenant-id":   cfg.TenantID,
			},
			expectedStatus: 404,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			nodeResp, nodeBody, _, nodeErr := makeRequest(client, tc.method, cfg.NodeJSURL+tc.path, tc.headers)
			goResp, goBody, _, goErr := makeRequest(client, tc.method, cfg.GoURL+tc.path, tc.headers)

			if nodeErr != nil {
				t.Skipf("Node.js server error: %v", nodeErr)
			}
			if goErr != nil {
				t.Skipf("Go server error: %v", goErr)
			}

			// Compare status codes
			if nodeResp.StatusCode != goResp.StatusCode {
				t.Errorf("Status code mismatch: Node.js=%d, Go=%d",
					nodeResp.StatusCode, goResp.StatusCode)
			}

			// Compare error response structure
			var nodeError, goError map[string]interface{}
			json.Unmarshal(nodeBody, &nodeError)
			json.Unmarshal(goBody, &goError)

			// Check if error field exists in both
			_, nodeHasError := nodeError["error"]
			_, goHasError := goError["error"]
			_, nodeHasMessage := nodeError["message"]
			_, goHasMessage := goError["message"]

			if nodeHasError != goHasError {
				t.Logf("Error field presence: Node.js=%v, Go=%v", nodeHasError, goHasError)
			}
			if nodeHasMessage != goHasMessage {
				t.Logf("Message field presence: Node.js=%v, Go=%v", nodeHasMessage, goHasMessage)
			}

			t.Logf("Node.js: %d - %s", nodeResp.StatusCode, string(nodeBody))
			t.Logf("Go:      %d - %s", goResp.StatusCode, string(goBody))
		})
	}
}

// TestEdgeCases tests edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	cfg := getTestConfig()
	if cfg.AuthToken == "" {
		t.Skip("AUTH_TOKEN not set")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	headers := getAuthHeaders(cfg)

	edgeCases := []struct {
		name string
		path string
	}{
		{"Large Page Size", "/api/shopee/orders?page=1&pageSize=100"},
		{"Zero Page", "/api/shopee/orders?page=0&pageSize=10"},
		{"Negative Page", "/api/shopee/orders?page=-1&pageSize=10"},
		{"Missing Page Param", "/api/shopee/orders?pageSize=10"},
		{"Empty Search", "/api/shopee/products?search="},
		{"Special Chars Search", "/api/shopee/products?search=%20%26%3D"},
		{"Very Large Page Number", "/api/shopee/orders?page=999999&pageSize=10"},
	}

	for _, ec := range edgeCases {
		t.Run(ec.name, func(t *testing.T) {
			nodeResp, _, _, nodeErr := makeRequest(client, "GET", cfg.NodeJSURL+ec.path, headers)
			goResp, _, _, goErr := makeRequest(client, "GET", cfg.GoURL+ec.path, headers)

			if nodeErr != nil || goErr != nil {
				t.Skipf("Server error: Node=%v, Go=%v", nodeErr, goErr)
			}

			if nodeResp.StatusCode != goResp.StatusCode {
				t.Errorf("Status mismatch for %s: Node.js=%d, Go=%d",
					ec.name, nodeResp.StatusCode, goResp.StatusCode)
			} else {
				t.Logf("✓ %s: Both return %d", ec.name, nodeResp.StatusCode)
			}
		})
	}
}
