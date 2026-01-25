// Package parity provides test utilities for comparing Node.js vs Go API responses
package parity

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"
)

// Default constants
const (
	NodeJSBaseURL = "http://localhost:3000"
	GoBaseURL     = "http://localhost:8080"
	TenantID      = "yumna_bertigamart"
)

// TestConfig holds test configuration
type TestConfig struct {
	NodeJSURL string
	GoURL     string
	TenantID  string
	AuthToken string
}

// getTestConfig returns test configuration from environment
func getTestConfig() *TestConfig {
	nodeURL := os.Getenv("NODEJS_URL")
	if nodeURL == "" {
		nodeURL = NodeJSBaseURL
	}
	goURL := os.Getenv("GO_URL")
	if goURL == "" {
		goURL = GoBaseURL
	}
	tenantID := os.Getenv("TENANT_ID")
	if tenantID == "" {
		tenantID = TenantID
	}
	return &TestConfig{
		NodeJSURL: nodeURL,
		GoURL:     goURL,
		TenantID:  tenantID,
		AuthToken: os.Getenv("AUTH_TOKEN"),
	}
}

// ParityResult represents the result of a parity test
type ParityResult struct {
	Endpoint     string
	NodeJSTime   time.Duration
	GoTime       time.Duration
	Match        bool
	Differences  []string
	NodeJSStatus int
	GoStatus     int
}

// makeRequest makes an HTTP request and returns the response
func makeRequest(client *http.Client, method, url string, headers map[string]string) (*http.Response, []byte, time.Duration, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, 0, err
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	start := time.Now()
	resp, err := client.Do(req)
	elapsed := time.Since(start)
	if err != nil {
		return nil, nil, elapsed, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	return resp, body, elapsed, err
}

// compareJSON compares two JSON responses and returns differences
func compareJSON(nodeJS, golang []byte) (bool, []string) {
	var nodeData, goData map[string]interface{}

	if err := json.Unmarshal(nodeJS, &nodeData); err != nil {
		return false, []string{"Failed to parse Node.js response as JSON"}
	}
	if err := json.Unmarshal(golang, &goData); err != nil {
		return false, []string{"Failed to parse Go response as JSON"}
	}

	differences := compareObjects("", nodeData, goData)
	return len(differences) == 0, differences
}

// compareObjects recursively compares two objects
func compareObjects(path string, node, golang interface{}) []string {
	var diffs []string

	if node == nil && golang == nil {
		return diffs
	}

	nodeType := reflect.TypeOf(node)
	goType := reflect.TypeOf(golang)

	if nodeType != goType {
		if isNumeric(node) && isNumeric(golang) {
			if toFloat(node) != toFloat(golang) {
				diffs = append(diffs, fmt.Sprintf("%s: value mismatch %v vs %v", path, node, golang))
			}
			return diffs
		}
		diffs = append(diffs, fmt.Sprintf("%s: type mismatch %T vs %T", path, node, golang))
		return diffs
	}

	switch nodeVal := node.(type) {
	case map[string]interface{}:
		goVal := golang.(map[string]interface{})
		allKeys := make(map[string]bool)
		for k := range nodeVal {
			allKeys[k] = true
		}
		for k := range goVal {
			allKeys[k] = true
		}

		for k := range allKeys {
			newPath := k
			if path != "" {
				newPath = path + "." + k
			}

			nodeV, nodeHas := nodeVal[k]
			goV, goHas := goVal[k]

			if !nodeHas {
				diffs = append(diffs, fmt.Sprintf("%s: only in Go", newPath))
				continue
			}
			if !goHas {
				diffs = append(diffs, fmt.Sprintf("%s: only in Node.js", newPath))
				continue
			}

			if isTimestampField(k) {
				continue
			}

			diffs = append(diffs, compareObjects(newPath, nodeV, goV)...)
		}

	case []interface{}:
		goVal := golang.([]interface{})
		if len(nodeVal) != len(goVal) {
			diffs = append(diffs, fmt.Sprintf("%s: array length mismatch %d vs %d", path, len(nodeVal), len(goVal)))
			return diffs
		}
		for i := range nodeVal {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			diffs = append(diffs, compareObjects(newPath, nodeVal[i], goVal[i])...)
		}

	default:
		if !reflect.DeepEqual(node, golang) {
			diffs = append(diffs, fmt.Sprintf("%s: value mismatch %v vs %v", path, node, golang))
		}
	}

	return diffs
}

// isTimestampField checks if a field name is a timestamp field
func isTimestampField(field string) bool {
	timestampFields := []string{
		"createdAt", "updatedAt", "created_at", "updated_at",
		"timestamp", "time", "date", "expiresAt", "lastSync",
		"syncedAt", "processedAt", "lastUpdated",
	}
	fieldLower := strings.ToLower(field)
	for _, tf := range timestampFields {
		if strings.ToLower(tf) == fieldLower || strings.HasSuffix(fieldLower, "at") || strings.HasSuffix(fieldLower, "time") {
			return true
		}
	}
	return false
}

// isNumeric checks if value is numeric
func isNumeric(v interface{}) bool {
	switch v.(type) {
	case float64, float32, int, int64, int32:
		return true
	}
	return false
}

// toFloat converts numeric value to float64
func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	}
	return 0
}

// getAuthHeaders returns common authentication headers
func getAuthHeaders(cfg *TestConfig) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + cfg.AuthToken,
		"x-tenant-id":   cfg.TenantID,
	}
}

// logTimings logs response times comparison
func logTimings(t *testing.T, nodeResp, goResp *http.Response, nodeTime, goTime time.Duration) {
	t.Logf("Node.js: %d in %v", nodeResp.StatusCode, nodeTime)
	speedup := float64(nodeTime) / float64(goTime)
	t.Logf("Go: %d in %v (%.1fx)", goResp.StatusCode, goTime, speedup)
}

// logDiffs logs JSON differences with limit
func logDiffs(t *testing.T, diffs []string, maxDiffs int) {
	t.Logf("Response differences (first %d):", maxDiffs)
	for i, d := range diffs {
		if i >= maxDiffs {
			break
		}
		t.Logf("  - %s", d)
	}
}
