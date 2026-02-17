package route

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildCoverageReport(t *testing.T) {
	tmpDir := t.TempDir()
	apiDir := filepath.Join(tmpDir, "frontend", "src", "api")
	require.NoError(t, os.MkdirAll(apiDir, 0o755))

	content := `
import api from "./client"

export async function callRoutes() {
  await api.get("/orders")
  await api.post("/orders")
  await api.get("/missing")
  await api.delete("/orders")
}
`

	require.NoError(t, os.WriteFile(filepath.Join(apiDir, "orders.ts"), []byte(content), 0o644))

	scanner := NewScannerService(tmpDir)
	backendRoutes := []RouteMapping{
		{Method: "GET", Path: "/api/orders", Tags: []string{"orders"}},
		{Method: "POST", Path: "/api/orders", Tags: []string{"orders"}},
		{Method: "GET", Path: "/api/backend-only", Tags: []string{"backend-only"}},
	}

	report := scanner.BuildCoverageReport(backendRoutes)

	assert.Equal(t, 3, report.TotalRoutes)
	assert.Equal(t, 4, report.TotalCalledRoutes)
	assert.Equal(t, 2, len(report.Categories.Connected))
	assert.Equal(t, 2, len(report.Categories.FrontendOnly))
	assert.Equal(t, 1, len(report.Categories.BackendOnly))
	assert.Equal(t, 1, len(report.Categories.Unused))

	frontendOnlyStatus := map[string]string{}
	for _, item := range report.Categories.FrontendOnly {
		frontendOnlyStatus[item.Method+" "+item.Endpoint] = item.Status
	}

	assert.Equal(t, "missing_backend_route", frontendOnlyStatus["GET /api/missing"])
	assert.Equal(t, "method_mismatch", frontendOnlyStatus["DELETE /api/orders"])
	assert.Equal(t, 1, report.Statistics["method_mismatch_total"])

	component, exists := report.Components["orders"]
	if assert.True(t, exists) {
		assert.Contains(t, component.RoutesCalled, "GET /api/orders")
		assert.Contains(t, component.RoutesCalled, "POST /api/orders")
	}
}

func TestBuildCoverageReport_NormalizationCases(t *testing.T) {
	testCases := []struct {
		name                string
		frontendContent     string
		backendRoutes       []RouteMapping
		expectedConnected   int
		expectedFrontOnly   int
		expectedStatusByKey map[string]string
	}{
		{
			name: "trailing slash and query normalized to connected",
			frontendContent: `
import api from "./client"
export async function run() {
  await api.get("/orders/")
  await api.get("/orders?id=1")
}
`,
			backendRoutes: []RouteMapping{
				{Method: "GET", Path: "/api/orders", Tags: []string{"orders"}},
			},
			expectedConnected: 1,
			expectedFrontOnly: 0,
		},
		{
			name: "dynamic segment parity frontend template vs backend param",
			frontendContent: `
import api from "./client"
export async function run(orderSn: string) {
  await api.get("/orders/${orderSn}")
}
`,
			backendRoutes: []RouteMapping{
				{Method: "GET", Path: "/api/orders/:id", Tags: []string{"orders"}},
			},
			expectedConnected: 1,
			expectedFrontOnly: 0,
		},
		{
			name: "same endpoint different method flagged method_mismatch",
			frontendContent: `
import api from "./client"
export async function run() {
  await api.delete("/orders")
}
`,
			backendRoutes: []RouteMapping{
				{Method: "GET", Path: "/api/orders", Tags: []string{"orders"}},
			},
			expectedConnected: 0,
			expectedFrontOnly: 1,
			expectedStatusByKey: map[string]string{
				"DELETE /api/orders": "method_mismatch",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			apiDir := filepath.Join(tmpDir, "frontend", "src", "api")
			require.NoError(t, os.MkdirAll(apiDir, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(apiDir, "orders.ts"), []byte(tc.frontendContent), 0o644))

			scanner := NewScannerService(tmpDir)
			report := scanner.BuildCoverageReport(tc.backendRoutes)

			assert.Equal(t, tc.expectedConnected, len(report.Categories.Connected))
			assert.Equal(t, tc.expectedFrontOnly, len(report.Categories.FrontendOnly))

			if len(tc.expectedStatusByKey) > 0 {
				actualStatus := make(map[string]string)
				for _, item := range report.Categories.FrontendOnly {
					actualStatus[item.Method+" "+item.Endpoint] = item.Status
				}

				for key, expectedStatus := range tc.expectedStatusByKey {
					assert.Equal(t, expectedStatus, actualStatus[key])
				}
			}
		})
	}
}
