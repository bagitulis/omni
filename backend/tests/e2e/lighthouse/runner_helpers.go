package lighthouse

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// extractIssues extracts failed audits from Lighthouse report
func (r *Runner) extractIssues(report LighthouseJSONReport) Issues {
	issues := Issues{
		Performance:   []Issue{},
		Accessibility: []Issue{},
		BestPractices: []Issue{},
		SEO:           []Issue{},
	}

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
		if page.Desktop != nil {
			r.countIssues(issueCount, page.Desktop.Issues, page.Name)
		}
		if page.Mobile != nil {
			r.countIssues(issueCount, page.Mobile.Issues, page.Name)
		}
	}

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

// contains checks if a string slice contains an item
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// getImpact returns impact level based on score
func getImpact(score float64) string {
	if score < 0.5 {
		return "high"
	} else if score < 0.9 {
		return "medium"
	}
	return "low"
}

// getRecommendation returns recommendation for a specific audit
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
