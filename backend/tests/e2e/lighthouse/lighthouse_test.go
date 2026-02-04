// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse_test

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/omni/backend/tests/e2e/lighthouse"
)

var (
	quickMode    = flag.Bool("quick", false, "Run quick test with only 3 sample pages")
	categoryFlag = flag.String("category", "", "Test only specific category (core, product, order, script, report, analytics, settings)")
	verboseFlag  = flag.Bool("verbose", false, "Enable verbose output")
)

func TestMain(m *testing.M) {
	flag.Parse()
	if skip, reason := shouldSkipLighthouse(); skip {
		fmt.Printf("Skipping Lighthouse tests: %s\n", reason)
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func shouldSkipLighthouse() (bool, string) {
	if os.Getenv("LIGHTHOUSE_SKIP") != "" {
		return true, "LIGHTHOUSE_SKIP set"
	}
	if !chromeAvailable() {
		return true, "chrome binary not found"
	}
	config := lighthouse.DefaultConfig()
	if !urlReachable(config.FrontendURL + "/login") {
		return true, "frontend not reachable at " + config.FrontendURL
	}
	return false, ""
}

func chromeAvailable() bool {
	if os.Getenv("CHROME_PATH") != "" {
		if _, err := os.Stat(os.Getenv("CHROME_PATH")); err == nil {
			return true
		}
	}
	binaries := []string{"google-chrome", "chromium", "chromium-browser", "chrome"}
	for _, bin := range binaries {
		if _, err := exec.LookPath(bin); err == nil {
			return true
		}
	}
	return false
}

func urlReachable(url string) bool {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	_ = resp.Body.Close()
	return resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest
}

// TestLighthouseFull runs the full Lighthouse test suite
func TestLighthouseFull(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping full Lighthouse test in short mode")
	}

	config := lighthouse.DefaultConfig()
	runner := lighthouse.NewRunner(config)

	var routes []lighthouse.Route
	if *quickMode {
		routes = lighthouse.GetQuickTestRoutes()
		t.Logf("Running quick test with %d routes", len(routes))
	} else if *categoryFlag != "" {
		routes = lighthouse.GetRoutesByCategory(*categoryFlag)
		t.Logf("Running test for category '%s' with %d routes", *categoryFlag, len(routes))
	} else {
		routes = lighthouse.GetAllRoutes()
		t.Logf("Running full test with %d routes", len(routes))
	}

	report, err := runner.Run(routes)
	if err != nil {
		t.Fatalf("Lighthouse test failed: %v", err)
	}

	// Save reports
	if err := runner.SaveReport(); err != nil {
		t.Errorf("Failed to save JSON report: %v", err)
	}

	generator := lighthouse.NewReportGenerator(report, config)
	if err := generator.GenerateHTMLReport(); err != nil {
		t.Errorf("Failed to generate HTML report: %v", err)
	}
	if err := generator.GenerateMarkdownReport(); err != nil {
		t.Errorf("Failed to generate Markdown report: %v", err)
	}

	runner.PrintSummary()

	// Check thresholds
	thresholds := lighthouse.DefaultThresholds()
	checkThresholds(t, report, thresholds)
}

// TestLighthouseQuick runs a quick Lighthouse test (3 pages)
func TestLighthouseQuick(t *testing.T) {
	config := lighthouse.DefaultConfig()
	runner := lighthouse.NewRunner(config)
	routes := lighthouse.GetQuickTestRoutes()

	t.Logf("Running quick Lighthouse test with %d routes", len(routes))

	report, err := runner.Run(routes)
	if err != nil {
		t.Fatalf("Quick Lighthouse test failed: %v", err)
	}

	if err := runner.SaveReport(); err != nil {
		t.Errorf("Failed to save report: %v", err)
	}

	runner.PrintSummary()

	// Use internal app thresholds (accounts for intentional noindex meta tags)
	thresholds := lighthouse.InternalAppThresholds()
	checkThresholds(t, report, thresholds)
}

// TestLighthouseCategory runs Lighthouse test for a specific category
func TestLighthouseCategory(t *testing.T) {
	categories := []string{"core", "product", "order", "script", "report", "analytics", "settings"}

	for _, category := range categories {
		t.Run(category, func(t *testing.T) {
			if testing.Short() {
				t.Skip("Skipping category test in short mode")
			}

			config := lighthouse.DefaultConfig()
			runner := lighthouse.NewRunner(config)
			routes := lighthouse.GetRoutesByCategory(category)

			if len(routes) == 0 {
				t.Skipf("No routes found for category: %s", category)
			}

			t.Logf("Testing category '%s' with %d routes", category, len(routes))

			report, err := runner.Run(routes)
			if err != nil {
				t.Fatalf("Category test failed: %v", err)
			}

			runner.PrintSummary()

			thresholds := lighthouse.DefaultThresholds()
			checkThresholds(t, report, thresholds)
		})
	}
}

// checkThresholds verifies scores meet minimum thresholds
func checkThresholds(t *testing.T, report *lighthouse.FullReport, thresholds lighthouse.ScoreThresholds) {
	t.Helper()

	// Check desktop scores
	if report.Summary.DesktopAvgPerformance < float64(thresholds.Performance) {
		t.Errorf("Desktop Performance %.0f below threshold %d",
			report.Summary.DesktopAvgPerformance, thresholds.Performance)
	}
	if report.Summary.DesktopAvgAccessibility < float64(thresholds.Accessibility) {
		t.Errorf("Desktop Accessibility %.0f below threshold %d",
			report.Summary.DesktopAvgAccessibility, thresholds.Accessibility)
	}
	if report.Summary.DesktopAvgBestPractices < float64(thresholds.BestPractices) {
		t.Errorf("Desktop Best Practices %.0f below threshold %d",
			report.Summary.DesktopAvgBestPractices, thresholds.BestPractices)
	}
	if report.Summary.DesktopAvgSEO < float64(thresholds.SEO) {
		t.Errorf("Desktop SEO %.0f below threshold %d",
			report.Summary.DesktopAvgSEO, thresholds.SEO)
	}

	// Check mobile scores
	if report.Summary.MobileAvgPerformance < float64(thresholds.Performance) {
		t.Errorf("Mobile Performance %.0f below threshold %d",
			report.Summary.MobileAvgPerformance, thresholds.Performance)
	}
	if report.Summary.MobileAvgAccessibility < float64(thresholds.Accessibility) {
		t.Errorf("Mobile Accessibility %.0f below threshold %d",
			report.Summary.MobileAvgAccessibility, thresholds.Accessibility)
	}
	if report.Summary.MobileAvgBestPractices < float64(thresholds.BestPractices) {
		t.Errorf("Mobile Best Practices %.0f below threshold %d",
			report.Summary.MobileAvgBestPractices, thresholds.BestPractices)
	}
	if report.Summary.MobileAvgSEO < float64(thresholds.SEO) {
		t.Errorf("Mobile SEO %.0f below threshold %d",
			report.Summary.MobileAvgSEO, thresholds.SEO)
	}

	// Log problems summary
	config := lighthouse.DefaultConfig()
	generator := lighthouse.NewReportGenerator(report, config)
	problems := generator.GetProblemSummary()

	if len(problems) > 0 {
		fmt.Println("\n⚠️  Areas needing improvement:")
		for _, p := range problems {
			fmt.Printf("   - %s %s: %.0f (target: %.0f, gap: %.0f)\n",
				p.Device, p.Category, p.CurrentScore, p.TargetScore, p.Gap)
		}
	}
}

// BenchmarkLighthouseSinglePage benchmarks a single page test
func BenchmarkLighthouseSinglePage(b *testing.B) {
	config := lighthouse.DefaultConfig()
	runner := lighthouse.NewRunner(config)

	routes := []lighthouse.Route{
		{Name: "Dashboard", Path: "/"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := runner.Run(routes)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}
