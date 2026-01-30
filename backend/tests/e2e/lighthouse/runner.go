// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Runner orchestrates Lighthouse testing
type Runner struct {
	config  *Config
	browser *Browser
	report  *FullReport
}

// NewRunner creates a new test runner
func NewRunner(config *Config) *Runner {
	if config == nil {
		config = DefaultConfig()
	}
	return &Runner{
		config:  config,
		browser: NewBrowser(config),
		report: &FullReport{
			TestInfo: TestInfo{
				TestType: "Lighthouse Performance Test",
				BaseURL:  config.FrontendURL,
			},
			PageTests: []PageResult{},
		},
	}
}

// Run executes the full Lighthouse test suite
func (r *Runner) Run(routes []Route) (*FullReport, error) {
	r.report.TestInfo.StartTime = time.Now()
	r.report.TestInfo.Credentials.Username = r.config.Username

	// Initialize browser
	if err := r.browser.Init(); err != nil {
		return nil, fmt.Errorf("failed to init browser: %w", err)
	}
	defer r.browser.Close()

	// Login
	fmt.Println("🔐 Step 1: Login")
	loginResult, err := r.browser.Login()
	r.report.LoginTest = *loginResult
	if err != nil {
		fmt.Printf("   ❌ Login failed: %s\n", err)
		return r.report, fmt.Errorf("login failed: %w", err)
	}
	fmt.Printf("   ✅ Login successful! Duration: %dms\n\n", loginResult.Duration)

	// Test each route
	fmt.Printf("🧪 Step 2: Running Lighthouse Tests on %d pages\n", len(routes))
	fmt.Printf("   ⚠️  Estimated time: %d - %d minutes\n\n", len(routes)*1, len(routes)*2)

	for i, route := range routes {
		fmt.Printf("\n📄 [%d/%d] Testing: %s\n", i+1, len(routes), route.Name)
		fmt.Printf("   URL: %s%s\n", r.config.FrontendURL, route.Path)
		fmt.Println(strings.Repeat("─", 60))

		result := r.testPage(route)
		r.report.PageTests = append(r.report.PageTests, result)
		r.report.Summary.TotalPages++
	}

	// Finalize report
	r.report.TestInfo.EndTime = time.Now()
	r.report.TestInfo.Duration = r.report.TestInfo.EndTime.Sub(r.report.TestInfo.StartTime).Milliseconds()
	r.report.Summary.TotalDuration = r.report.TestInfo.Duration

	r.calculateAverages()
	r.extractCommonIssues()

	return r.report, nil
}

// testPage runs Lighthouse tests on a single page
func (r *Runner) testPage(route Route) PageResult {
	url := fmt.Sprintf("%s%s", r.config.FrontendURL, route.Path)
	safeName := strings.ToLower(route.Name)
	safeName = strings.ReplaceAll(safeName, " ", "-")
	safeName = strings.ReplaceAll(safeName, "/", "-")

	result := PageResult{
		Name: route.Name,
		Path: route.Path,
		URL:  url,
	}

	// Desktop test
	fmt.Println("   🖥️  Desktop Lighthouse Test...")
	desktopResult := r.runLighthouse(url, safeName, false)
	result.Desktop = desktopResult
	if desktopResult.Success {
		fmt.Printf("      ✅ Performance: %d | A11y: %d | BP: %d | SEO: %d\n",
			desktopResult.Scores.Performance,
			desktopResult.Scores.Accessibility,
			desktopResult.Scores.BestPractices,
			desktopResult.Scores.SEO)
		fmt.Printf("      ⏱️  Duration: %.1fs\n", desktopResult.Duration.Seconds())
	} else {
		fmt.Printf("      ❌ Desktop failed: %s\n", desktopResult.Error)
	}

	// Mobile test
	fmt.Println("   📱 Mobile Lighthouse Test...")
	mobileResult := r.runLighthouse(url, safeName, true)
	result.Mobile = mobileResult
	if mobileResult.Success {
		fmt.Printf("      ✅ Performance: %d | A11y: %d | BP: %d | SEO: %d\n",
			mobileResult.Scores.Performance,
			mobileResult.Scores.Accessibility,
			mobileResult.Scores.BestPractices,
			mobileResult.Scores.SEO)
		fmt.Printf("      ⏱️  Duration: %.1fs\n", mobileResult.Duration.Seconds())
	} else {
		fmt.Printf("      ❌ Mobile failed: %s\n", mobileResult.Error)
	}

	// Take screenshots
	if desktopScreenshot, err := r.browser.TakeScreenshot(url, safeName, false); err == nil {
		if result.Desktop != nil {
			result.Desktop.Screenshot = desktopScreenshot
		}
		r.report.Summary.TotalScreenshots++
	}
	if mobileScreenshot, err := r.browser.TakeScreenshot(url, safeName, true); err == nil {
		if result.Mobile != nil {
			result.Mobile.Screenshot = mobileScreenshot
		}
		r.report.Summary.TotalScreenshots++
	}

	return result
}

// runLighthouse executes Lighthouse CLI for a single test
func (r *Runner) runLighthouse(url, name string, isMobile bool) *DeviceResult {
	startTime := time.Now()
	result := &DeviceResult{
		Success: false,
	}

	// Build Lighthouse CLI command
	args := []string{
		url,
		"--output=json",
		"--output-path=stdout",
		"--chrome-flags=--headless --no-sandbox --disable-gpu",
		"--only-categories=performance,accessibility,best-practices,seo",
		"--quiet",
	}

	if isMobile {
		// For mobile: use form-factor and screen emulation
		args = append(args,
			"--form-factor=mobile",
			"--screenEmulation.mobile=true",
			"--screenEmulation.width=375",
			"--screenEmulation.height=812",
			"--screenEmulation.deviceScaleFactor=2",
			"--throttling.cpuSlowdownMultiplier=4",
		)
	} else {
		args = append(args, "--preset=desktop")
	}

	// Add cookies for authentication
	cookieHeader := r.browser.GetCookiesAsHeader()
	if cookieHeader != "" {
		headersJSON, _ := json.Marshal(map[string]string{
			"Cookie": cookieHeader,
		})
		args = append(args, fmt.Sprintf("--extra-headers=%s", string(headersJSON)))
	}

	// Execute Lighthouse CLI
	cmd := exec.Command("lighthouse", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Error = fmt.Sprintf("lighthouse command failed: %s - %s", err.Error(), stderr.String())
		result.Duration = time.Since(startTime)
		return result
	}

	// Parse Lighthouse JSON output
	var lhReport LighthouseJSONReport
	if err := json.Unmarshal(stdout.Bytes(), &lhReport); err != nil {
		result.Error = fmt.Sprintf("failed to parse lighthouse output: %s", err.Error())
		result.Duration = time.Since(startTime)
		return result
	}

	// Extract scores
	result.Scores = Scores{
		Performance:   int(lhReport.Categories["performance"].Score * 100),
		Accessibility: int(lhReport.Categories["accessibility"].Score * 100),
		BestPractices: int(lhReport.Categories["best-practices"].Score * 100),
		SEO:           int(lhReport.Categories["seo"].Score * 100),
	}

	// Extract metrics
	result.Metrics = r.extractMetrics(lhReport)

	// Extract issues
	result.Issues = r.extractIssues(lhReport)

	result.Success = true
	result.Duration = time.Since(startTime)
	return result
}

// extractMetrics extracts Core Web Vitals from Lighthouse report
func (r *Runner) extractMetrics(report LighthouseJSONReport) Metrics {
	return Metrics{
		FirstContentfulPaint:   report.Audits["first-contentful-paint"].NumericValue,
		LargestContentfulPaint: report.Audits["largest-contentful-paint"].NumericValue,
		TotalBlockingTime:      report.Audits["total-blocking-time"].NumericValue,
		CumulativeLayoutShift:  report.Audits["cumulative-layout-shift"].NumericValue,
		SpeedIndex:             report.Audits["speed-index"].NumericValue,
		TimeToInteractive:      report.Audits["interactive"].NumericValue,
	}
}

// extractIssues extracts failed audits from Lighthouse report
func (r *Runner) extractIssues(report LighthouseJSONReport) Issues {
	issues := Issues{
		Performance:   []Issue{},
		Accessibility: []Issue{},
		BestPractices: []Issue{},
		SEO:           []Issue{},
	}

	// Map audits to categories (simplified - in real implementation, use category refs)
	performanceAudits := []string{"first-contentful-paint", "largest-contentful-paint", "total-blocking-time", "cumulative-layout-shift", "speed-index", "interactive"}
	accessibilityAudits := []string{"color-contrast", "button-name", "image-alt", "label", "link-name", "aria-allowed-attr"}
	bestPracticesAudits := []string{"uses-http2", "uses-passive-event-listeners", "no-document-write", "js-libraries"}
	seoAudits := []string{"meta-description", "document-title", "link-text", "robots-txt", "hreflang"}

	for auditID, audit := range report.Audits {
		if audit.Score == nil || *audit.Score >= 0.9 {
			continue
		}

		issue := Issue{
			ID:           auditID,
			Title:        audit.Title,
			Description:  audit.Description,
			Score:        *audit.Score,
			DisplayValue: audit.DisplayValue,
		}

		// Determine category
		if contains(performanceAudits, auditID) {
			issues.Performance = append(issues.Performance, issue)
		} else if contains(accessibilityAudits, auditID) {
			issues.Accessibility = append(issues.Accessibility, issue)
		} else if contains(bestPracticesAudits, auditID) {
			issues.BestPractices = append(issues.BestPractices, issue)
		} else if contains(seoAudits, auditID) {
			issues.SEO = append(issues.SEO, issue)
		}
	}

	return issues
}

// calculateAverages calculates average scores across all pages
func (r *Runner) calculateAverages() {
	var desktopPerf, desktopA11y, desktopBP, desktopSEO float64
	var mobilePerf, mobileA11y, mobileBP, mobileSEO float64
	var desktopCount, mobileCount float64

	for _, page := range r.report.PageTests {
		if page.Desktop != nil && page.Desktop.Success {
			desktopPerf += float64(page.Desktop.Scores.Performance)
			desktopA11y += float64(page.Desktop.Scores.Accessibility)
			desktopBP += float64(page.Desktop.Scores.BestPractices)
			desktopSEO += float64(page.Desktop.Scores.SEO)
			desktopCount++
		}
		if page.Mobile != nil && page.Mobile.Success {
			mobilePerf += float64(page.Mobile.Scores.Performance)
			mobileA11y += float64(page.Mobile.Scores.Accessibility)
			mobileBP += float64(page.Mobile.Scores.BestPractices)
			mobileSEO += float64(page.Mobile.Scores.SEO)
			mobileCount++
		}
	}

	if desktopCount > 0 {
		r.report.Summary.DesktopAvgPerformance = desktopPerf / desktopCount
		r.report.Summary.DesktopAvgAccessibility = desktopA11y / desktopCount
		r.report.Summary.DesktopAvgBestPractices = desktopBP / desktopCount
		r.report.Summary.DesktopAvgSEO = desktopSEO / desktopCount
	}

	if mobileCount > 0 {
		r.report.Summary.MobileAvgPerformance = mobilePerf / mobileCount
		r.report.Summary.MobileAvgAccessibility = mobileA11y / mobileCount
		r.report.Summary.MobileAvgBestPractices = mobileBP / mobileCount
		r.report.Summary.MobileAvgSEO = mobileSEO / mobileCount
	}
}

// extractCommonIssues identifies issues that appear across multiple pages
func (r *Runner) extractCommonIssues() {
	issueCount := make(map[string]*CommonIssue)

	for _, page := range r.report.PageTests {
		// Process desktop issues
		if page.Desktop != nil {
			r.countIssues(issueCount, page.Desktop.Issues, page.Name)
		}
		// Process mobile issues
		if page.Mobile != nil {
			r.countIssues(issueCount, page.Mobile.Issues, page.Name)
		}
	}

	// Filter to issues appearing on 2+ pages
	for _, issue := range issueCount {
		if issue.Occurrences >= 2 {
			r.report.CommonIssues = append(r.report.CommonIssues, *issue)
		}
	}
}

// countIssues counts occurrences of each issue
func (r *Runner) countIssues(counts map[string]*CommonIssue, issues Issues, pageName string) {
	processCategory := func(category string, issueList []Issue) {
		for _, issue := range issueList {
			if existing, ok := counts[issue.ID]; ok {
				existing.Occurrences++
				if !contains(existing.AffectedPages, pageName) {
					existing.AffectedPages = append(existing.AffectedPages, pageName)
				}
			} else {
				counts[issue.ID] = &CommonIssue{
					ID:             issue.ID,
					Title:          issue.Title,
					Category:       category,
					Impact:         getImpact(issue.Score),
					Occurrences:    1,
					AffectedPages:  []string{pageName},
					Recommendation: getRecommendation(issue.ID),
				}
			}
		}
	}

	processCategory("performance", issues.Performance)
	processCategory("accessibility", issues.Accessibility)
	processCategory("best-practices", issues.BestPractices)
	processCategory("seo", issues.SEO)
}

// SaveReport saves the report to a JSON file
func (r *Runner) SaveReport() error {
	reportPath := filepath.Join(r.config.ResultsDir, "lighthouse-full-report.json")
	data, err := json.MarshalIndent(r.report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	return os.WriteFile(reportPath, data, 0644)
}

// PrintSummary prints a summary of the test results
func (r *Runner) PrintSummary() {
	fmt.Println("\n\n╔════════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    📊 LIGHTHOUSE TEST SUMMARY 📊                    ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	fmt.Println("📋 Overall Statistics:")
	fmt.Printf("   Total Pages Tested: %d\n", r.report.Summary.TotalPages)
	fmt.Printf("   Total Screenshots: %d\n", r.report.Summary.TotalScreenshots)
	fmt.Printf("   Total Duration: %.1f minutes\n\n", float64(r.report.TestInfo.Duration)/1000/60)

	fmt.Println("🖥️  Desktop Averages:")
	fmt.Printf("   Performance:    %.0f/100\n", r.report.Summary.DesktopAvgPerformance)
	fmt.Printf("   Accessibility:  %.0f/100\n", r.report.Summary.DesktopAvgAccessibility)
	fmt.Printf("   Best Practices: %.0f/100\n", r.report.Summary.DesktopAvgBestPractices)
	fmt.Printf("   SEO:            %.0f/100\n\n", r.report.Summary.DesktopAvgSEO)

	fmt.Println("📱 Mobile Averages:")
	fmt.Printf("   Performance:    %.0f/100\n", r.report.Summary.MobileAvgPerformance)
	fmt.Printf("   Accessibility:  %.0f/100\n", r.report.Summary.MobileAvgAccessibility)
	fmt.Printf("   Best Practices: %.0f/100\n", r.report.Summary.MobileAvgBestPractices)
	fmt.Printf("   SEO:            %.0f/100\n\n", r.report.Summary.MobileAvgSEO)

	if len(r.report.CommonIssues) > 0 {
		fmt.Println("⚠️  Common Issues (appearing on 2+ pages):")
		for _, issue := range r.report.CommonIssues {
			fmt.Printf("   - [%s] %s (%d pages)\n", issue.Category, issue.Title, issue.Occurrences)
		}
		fmt.Println()
	}

	fmt.Printf("✅ Report saved to: %s/lighthouse-full-report.json\n", r.config.ResultsDir)
	fmt.Printf("📁 Screenshots saved to: %s\n", r.config.ScreenshotsDir)
}

// Helper functions
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func getImpact(score float64) string {
	if score < 0.5 {
		return "high"
	} else if score < 0.9 {
		return "medium"
	}
	return "low"
}

func getRecommendation(auditID string) string {
	recommendations := map[string]string{
		"color-contrast": "Ensure text has sufficient contrast ratio (4.5:1 for normal text, 3:1 for large text)",
		"button-name":    "Add accessible names to buttons using aria-label or visible text",
		"image-alt":      "Add alt text to all images",
		"label":          "Associate labels with form inputs using 'for' attribute",
		"link-name":      "Ensure links have descriptive text",
	}
	if rec, ok := recommendations[auditID]; ok {
		return rec
	}
	return "Refer to Lighthouse documentation for details"
}
