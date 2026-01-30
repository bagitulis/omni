// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ReportGenerator handles report generation and analysis
type ReportGenerator struct {
	report     *FullReport
	config     *Config
	thresholds ScoreThresholds
}

// NewReportGenerator creates a new report generator
func NewReportGenerator(report *FullReport, config *Config) *ReportGenerator {
	return &ReportGenerator{
		report:     report,
		config:     config,
		thresholds: DefaultThresholds(),
	}
}

// SetThresholds sets score thresholds for pass/fail
func (g *ReportGenerator) SetThresholds(thresholds ScoreThresholds) {
	g.thresholds = thresholds
}

// GenerateHTMLReport generates an HTML report
func (g *ReportGenerator) GenerateHTMLReport() error {
	html := g.buildHTMLReport()
	reportPath := filepath.Join(g.config.ResultsDir, "lighthouse-report.html")
	return os.WriteFile(reportPath, []byte(html), 0644)
}

// GenerateMarkdownReport generates a Markdown report
func (g *ReportGenerator) GenerateMarkdownReport() error {
	md := g.buildMarkdownReport()
	reportPath := filepath.Join(g.config.ResultsDir, "lighthouse-report.md")
	return os.WriteFile(reportPath, []byte(md), 0644)
}

// buildHTMLReport builds HTML report content
func (g *ReportGenerator) buildHTMLReport() string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
	<title>Lighthouse Test Report - OMNI</title>
	<style>
		body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; margin: 40px; }
		h1 { color: #1a73e8; }
		.summary { background: #f8f9fa; padding: 20px; border-radius: 8px; margin: 20px 0; }
		.scores { display: flex; gap: 20px; flex-wrap: wrap; }
		.score-card { background: white; padding: 15px; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); min-width: 150px; }
		.score-value { font-size: 32px; font-weight: bold; }
		.score-label { color: #666; font-size: 14px; }
		.pass { color: #0d9488; }
		.fail { color: #dc2626; }
		.warn { color: #d97706; }
		table { width: 100%%; border-collapse: collapse; margin: 20px 0; }
		th, td { padding: 12px; text-align: left; border-bottom: 1px solid #ddd; }
		th { background: #f1f3f5; }
		.issue-list { margin: 20px 0; }
		.issue { background: #fff3cd; padding: 10px; margin: 5px 0; border-radius: 4px; border-left: 4px solid #ffc107; }
	</style>
</head>
<body>
	<h1>🔦 Lighthouse Performance Report</h1>
	<p>Generated: %s</p>
	
	<div class="summary">
		<h2>📊 Summary</h2>
		<p>Total Pages: %d | Duration: %.1f minutes</p>
		
		<h3>🖥️ Desktop Averages</h3>
		<div class="scores">
			<div class="score-card">
				<div class="score-value %s">%.0f</div>
				<div class="score-label">Performance</div>
			</div>
			<div class="score-card">
				<div class="score-value %s">%.0f</div>
				<div class="score-label">Accessibility</div>
			</div>
			<div class="score-card">
				<div class="score-value %s">%.0f</div>
				<div class="score-label">Best Practices</div>
			</div>
			<div class="score-card">
				<div class="score-value %s">%.0f</div>
				<div class="score-label">SEO</div>
			</div>
		</div>
		
		<h3>📱 Mobile Averages</h3>
		<div class="scores">
			<div class="score-card">
				<div class="score-value %s">%.0f</div>
				<div class="score-label">Performance</div>
			</div>
			<div class="score-card">
				<div class="score-value %s">%.0f</div>
				<div class="score-label">Accessibility</div>
			</div>
			<div class="score-card">
				<div class="score-value %s">%.0f</div>
				<div class="score-label">Best Practices</div>
			</div>
			<div class="score-card">
				<div class="score-value %s">%.0f</div>
				<div class="score-label">SEO</div>
			</div>
		</div>
	</div>

	%s

	%s
</body>
</html>`,
		g.report.TestInfo.EndTime.Format("2006-01-02 15:04:05"),
		g.report.Summary.TotalPages,
		float64(g.report.TestInfo.Duration)/1000/60,
		g.scoreClass(g.report.Summary.DesktopAvgPerformance),
		g.report.Summary.DesktopAvgPerformance,
		g.scoreClass(g.report.Summary.DesktopAvgAccessibility),
		g.report.Summary.DesktopAvgAccessibility,
		g.scoreClass(g.report.Summary.DesktopAvgBestPractices),
		g.report.Summary.DesktopAvgBestPractices,
		g.scoreClass(g.report.Summary.DesktopAvgSEO),
		g.report.Summary.DesktopAvgSEO,
		g.scoreClass(g.report.Summary.MobileAvgPerformance),
		g.report.Summary.MobileAvgPerformance,
		g.scoreClass(g.report.Summary.MobileAvgAccessibility),
		g.report.Summary.MobileAvgAccessibility,
		g.scoreClass(g.report.Summary.MobileAvgBestPractices),
		g.report.Summary.MobileAvgBestPractices,
		g.scoreClass(g.report.Summary.MobileAvgSEO),
		g.report.Summary.MobileAvgSEO,
		g.buildPagesTable(),
		g.buildIssuesSection(),
	)
}

// buildPagesTable builds HTML table for page results
func (g *ReportGenerator) buildPagesTable() string {
	html := `<h2>📄 Page Results</h2>
	<table>
		<tr>
			<th>Page</th>
			<th>Desktop Perf</th>
			<th>Desktop A11y</th>
			<th>Mobile Perf</th>
			<th>Mobile A11y</th>
		</tr>`

	for _, page := range g.report.PageTests {
		desktopPerf, desktopA11y := "-", "-"
		mobilePerf, mobileA11y := "-", "-"

		if page.Desktop != nil && page.Desktop.Success {
			desktopPerf = fmt.Sprintf("%d", page.Desktop.Scores.Performance)
			desktopA11y = fmt.Sprintf("%d", page.Desktop.Scores.Accessibility)
		}
		if page.Mobile != nil && page.Mobile.Success {
			mobilePerf = fmt.Sprintf("%d", page.Mobile.Scores.Performance)
			mobileA11y = fmt.Sprintf("%d", page.Mobile.Scores.Accessibility)
		}

		html += fmt.Sprintf(`<tr>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
			<td>%s</td>
		</tr>`, page.Name, desktopPerf, desktopA11y, mobilePerf, mobileA11y)
	}

	html += "</table>"
	return html
}

// buildIssuesSection builds HTML for common issues
func (g *ReportGenerator) buildIssuesSection() string {
	if len(g.report.CommonIssues) == 0 {
		return ""
	}

	html := `<h2>⚠️ Common Issues</h2><div class="issue-list">`

	for _, issue := range g.report.CommonIssues {
		html += fmt.Sprintf(`<div class="issue">
			<strong>[%s] %s</strong> - %d pages affected<br>
			<em>%s</em>
		</div>`, issue.Category, issue.Title, issue.Occurrences, issue.Recommendation)
	}

	html += "</div>"
	return html
}

// buildMarkdownReport builds Markdown report content
func (g *ReportGenerator) buildMarkdownReport() string {
	md := fmt.Sprintf(`# 🔦 Lighthouse Performance Report

**Generated:** %s
**Total Pages:** %d
**Duration:** %.1f minutes

## 📊 Summary

### 🖥️ Desktop Averages
| Metric | Score | Status |
|--------|-------|--------|
| Performance | %.0f | %s |
| Accessibility | %.0f | %s |
| Best Practices | %.0f | %s |
| SEO | %.0f | %s |

### 📱 Mobile Averages
| Metric | Score | Status |
|--------|-------|--------|
| Performance | %.0f | %s |
| Accessibility | %.0f | %s |
| Best Practices | %.0f | %s |
| SEO | %.0f | %s |

## 📄 Page Results

| Page | Desktop Perf | Desktop A11y | Mobile Perf | Mobile A11y |
|------|-------------|--------------|-------------|-------------|
`,
		g.report.TestInfo.EndTime.Format("2006-01-02 15:04:05"),
		g.report.Summary.TotalPages,
		float64(g.report.TestInfo.Duration)/1000/60,
		g.report.Summary.DesktopAvgPerformance, g.statusEmoji(g.report.Summary.DesktopAvgPerformance),
		g.report.Summary.DesktopAvgAccessibility, g.statusEmoji(g.report.Summary.DesktopAvgAccessibility),
		g.report.Summary.DesktopAvgBestPractices, g.statusEmoji(g.report.Summary.DesktopAvgBestPractices),
		g.report.Summary.DesktopAvgSEO, g.statusEmoji(g.report.Summary.DesktopAvgSEO),
		g.report.Summary.MobileAvgPerformance, g.statusEmoji(g.report.Summary.MobileAvgPerformance),
		g.report.Summary.MobileAvgAccessibility, g.statusEmoji(g.report.Summary.MobileAvgAccessibility),
		g.report.Summary.MobileAvgBestPractices, g.statusEmoji(g.report.Summary.MobileAvgBestPractices),
		g.report.Summary.MobileAvgSEO, g.statusEmoji(g.report.Summary.MobileAvgSEO),
	)

	for _, page := range g.report.PageTests {
		desktopPerf, desktopA11y := "-", "-"
		mobilePerf, mobileA11y := "-", "-"

		if page.Desktop != nil && page.Desktop.Success {
			desktopPerf = fmt.Sprintf("%d", page.Desktop.Scores.Performance)
			desktopA11y = fmt.Sprintf("%d", page.Desktop.Scores.Accessibility)
		}
		if page.Mobile != nil && page.Mobile.Success {
			mobilePerf = fmt.Sprintf("%d", page.Mobile.Scores.Performance)
			mobileA11y = fmt.Sprintf("%d", page.Mobile.Scores.Accessibility)
		}

		md += fmt.Sprintf("| %s | %s | %s | %s | %s |\n",
			page.Name, desktopPerf, desktopA11y, mobilePerf, mobileA11y)
	}

	if len(g.report.CommonIssues) > 0 {
		md += "\n## ⚠️ Common Issues\n\n"
		for _, issue := range g.report.CommonIssues {
			md += fmt.Sprintf("### %s\n- **Category:** %s\n- **Affected Pages:** %d\n- **Recommendation:** %s\n\n",
				issue.Title, issue.Category, issue.Occurrences, issue.Recommendation)
		}
	}

	return md
}

// scoreClass returns CSS class based on score
func (g *ReportGenerator) scoreClass(score float64) string {
	if score >= 90 {
		return "pass"
	} else if score >= 50 {
		return "warn"
	}
	return "fail"
}

// statusEmoji returns emoji based on score
func (g *ReportGenerator) statusEmoji(score float64) string {
	if score >= float64(g.thresholds.Performance) {
		return "✅ Pass"
	} else if score >= 50 {
		return "⚠️ Warning"
	}
	return "❌ Fail"
}

// GetProblemSummary returns a summary of problems and recommendations
func (g *ReportGenerator) GetProblemSummary() []ProblemSummary {
	var problems []ProblemSummary

	// Check desktop scores
	if g.report.Summary.DesktopAvgPerformance < float64(g.thresholds.Performance) {
		problems = append(problems, ProblemSummary{
			Category:        "Performance",
			Device:          "Desktop",
			CurrentScore:    g.report.Summary.DesktopAvgPerformance,
			TargetScore:     float64(g.thresholds.Performance),
			Gap:             float64(g.thresholds.Performance) - g.report.Summary.DesktopAvgPerformance,
			Recommendations: g.getPerformanceRecommendations(),
		})
	}

	if g.report.Summary.DesktopAvgAccessibility < float64(g.thresholds.Accessibility) {
		problems = append(problems, ProblemSummary{
			Category:        "Accessibility",
			Device:          "Desktop",
			CurrentScore:    g.report.Summary.DesktopAvgAccessibility,
			TargetScore:     float64(g.thresholds.Accessibility),
			Gap:             float64(g.thresholds.Accessibility) - g.report.Summary.DesktopAvgAccessibility,
			Recommendations: g.getAccessibilityRecommendations(),
		})
	}

	// Check mobile scores
	if g.report.Summary.MobileAvgPerformance < float64(g.thresholds.Performance) {
		problems = append(problems, ProblemSummary{
			Category:        "Performance",
			Device:          "Mobile",
			CurrentScore:    g.report.Summary.MobileAvgPerformance,
			TargetScore:     float64(g.thresholds.Performance),
			Gap:             float64(g.thresholds.Performance) - g.report.Summary.MobileAvgPerformance,
			Recommendations: g.getPerformanceRecommendations(),
		})
	}

	if g.report.Summary.MobileAvgAccessibility < float64(g.thresholds.Accessibility) {
		problems = append(problems, ProblemSummary{
			Category:        "Accessibility",
			Device:          "Mobile",
			CurrentScore:    g.report.Summary.MobileAvgAccessibility,
			TargetScore:     float64(g.thresholds.Accessibility),
			Gap:             float64(g.thresholds.Accessibility) - g.report.Summary.MobileAvgAccessibility,
			Recommendations: g.getAccessibilityRecommendations(),
		})
	}

	// Sort by gap (biggest problems first)
	sort.Slice(problems, func(i, j int) bool {
		return problems[i].Gap > problems[j].Gap
	})

	return problems
}

// ProblemSummary represents a score gap and recommendations
type ProblemSummary struct {
	Category        string   `json:"category"`
	Device          string   `json:"device"`
	CurrentScore    float64  `json:"current_score"`
	TargetScore     float64  `json:"target_score"`
	Gap             float64  `json:"gap"`
	Recommendations []string `json:"recommendations"`
}

// getPerformanceRecommendations returns performance improvement suggestions
func (g *ReportGenerator) getPerformanceRecommendations() []string {
	return []string{
		"Optimize images with WebP format and lazy loading",
		"Implement code splitting and tree shaking",
		"Enable HTTP/2 and compression (gzip/brotli)",
		"Minimize main-thread work and JavaScript execution time",
		"Use font-display: swap for web fonts",
		"Implement resource hints (preconnect, prefetch)",
	}
}

// getAccessibilityRecommendations returns accessibility improvement suggestions
func (g *ReportGenerator) getAccessibilityRecommendations() []string {
	return []string{
		"Ensure sufficient color contrast (4.5:1 for normal text)",
		"Add accessible names to all interactive elements",
		"Associate labels with form inputs",
		"Add alt text to all images",
		"Ensure proper heading hierarchy",
		"Add ARIA attributes where needed",
	}
}

// SaveJSONReport saves the full report as JSON
func (g *ReportGenerator) SaveJSONReport() error {
	reportPath := filepath.Join(g.config.ResultsDir, "lighthouse-full-report.json")
	data, err := json.MarshalIndent(g.report, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}
	return os.WriteFile(reportPath, data, 0644)
}
