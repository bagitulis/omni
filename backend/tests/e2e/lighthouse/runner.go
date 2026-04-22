// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
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

	if err := r.browser.Init(); err != nil {
		return nil, fmt.Errorf("failed to init browser: %w", err)
	}
	defer r.browser.Close()

	fmt.Println("🔐 Step 1: Login")
	loginResult, err := r.browser.Login()
	r.report.LoginTest = *loginResult
	if err != nil {
		fmt.Printf("   ❌ Login failed: %s\n", err)
		return r.report, fmt.Errorf("login failed: %w", err)
	}
	fmt.Printf("   ✅ Login successful! Duration: %dms\n\n", loginResult.Duration)

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

	fmt.Println("   🖥️  Desktop Lighthouse Test...")
	desktopResult := r.runLighthouse(url, safeName, false)
	result.Desktop = desktopResult
	if desktopResult.Success {
		fmt.Printf("      ✅ Performance: %d | A11y: %d | BP: %d | SEO: %d\n",
			desktopResult.Scores.Performance, desktopResult.Scores.Accessibility,
			desktopResult.Scores.BestPractices, desktopResult.Scores.SEO)
		fmt.Printf("      ⏱️  Duration: %.1fs\n", desktopResult.Duration.Seconds())
	} else {
		fmt.Printf("      ❌ Desktop failed: %s\n", desktopResult.Error)
	}

	fmt.Println("   📱 Mobile Lighthouse Test...")
	mobileResult := r.runLighthouse(url, safeName, true)
	result.Mobile = mobileResult
	if mobileResult.Success {
		fmt.Printf("      ✅ Performance: %d | A11y: %d | BP: %d | SEO: %d\n",
			mobileResult.Scores.Performance, mobileResult.Scores.Accessibility,
			mobileResult.Scores.BestPractices, mobileResult.Scores.SEO)
		fmt.Printf("      ⏱️  Duration: %.1fs\n", mobileResult.Duration.Seconds())
	} else {
		fmt.Printf("      ❌ Mobile failed: %s\n", mobileResult.Error)
	}

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
	result := &DeviceResult{Success: false}

	args := []string{
		url,
		"--output=json",
		"--output-path=stdout",
		"--chrome-flags=--headless --no-sandbox --disable-gpu",
		"--only-categories=performance,accessibility,best-practices,seo",
		"--quiet",
	}

	if isMobile {
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

	cookieHeader := r.browser.GetCookiesAsHeader()
	if cookieHeader != "" {
		headersJSON, _ := json.Marshal(map[string]string{"Cookie": cookieHeader})
		args = append(args, fmt.Sprintf("--extra-headers=%s", string(headersJSON)))
	}

	cmd := exec.Command("lighthouse", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.Error = fmt.Sprintf("lighthouse command failed: %s - %s", err.Error(), stderr.String())
		result.Duration = time.Since(startTime)
		return result
	}

	var lhReport LighthouseJSONReport
	if err := json.Unmarshal(stdout.Bytes(), &lhReport); err != nil {
		result.Error = fmt.Sprintf("failed to parse lighthouse output: %s", err.Error())
		result.Duration = time.Since(startTime)
		return result
	}

	result.Scores = Scores{
		Performance:   int(lhReport.Categories["performance"].Score * 100),
		Accessibility: int(lhReport.Categories["accessibility"].Score * 100),
		BestPractices: int(lhReport.Categories["best-practices"].Score * 100),
		SEO:           int(lhReport.Categories["seo"].Score * 100),
	}

	result.Metrics = r.extractMetrics(lhReport)
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
