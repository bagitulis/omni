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

// ProblemSummary represents a score gap and recommendations
type ProblemSummary struct {
	Category        string   `json:"category"`
	Device          string   `json:"device"`
	CurrentScore    float64  `json:"current_score"`
	TargetScore     float64  `json:"target_score"`
	Gap             float64  `json:"gap"`
	Recommendations []string `json:"recommendations"`
}

// GetProblemSummary returns a summary of problems and recommendations
func (g *ReportGenerator) GetProblemSummary() []ProblemSummary {
	var problems []ProblemSummary

	if g.report.Summary.DesktopAvgPerformance < float64(g.thresholds.Performance) {
		problems = append(problems, ProblemSummary{
			Category: "Performance", Device: "Desktop",
			CurrentScore: g.report.Summary.DesktopAvgPerformance,
			TargetScore:  float64(g.thresholds.Performance),
			Gap:          float64(g.thresholds.Performance) - g.report.Summary.DesktopAvgPerformance,
			Recommendations: getPerformanceRecommendations(),
		})
	}

	if g.report.Summary.DesktopAvgAccessibility < float64(g.thresholds.Accessibility) {
		problems = append(problems, ProblemSummary{
			Category: "Accessibility", Device: "Desktop",
			CurrentScore: g.report.Summary.DesktopAvgAccessibility,
			TargetScore:  float64(g.thresholds.Accessibility),
			Gap:          float64(g.thresholds.Accessibility) - g.report.Summary.DesktopAvgAccessibility,
			Recommendations: getAccessibilityRecommendations(),
		})
	}

	if g.report.Summary.MobileAvgPerformance < float64(g.thresholds.Performance) {
		problems = append(problems, ProblemSummary{
			Category: "Performance", Device: "Mobile",
			CurrentScore: g.report.Summary.MobileAvgPerformance,
			TargetScore:  float64(g.thresholds.Performance),
			Gap:          float64(g.thresholds.Performance) - g.report.Summary.MobileAvgPerformance,
			Recommendations: getPerformanceRecommendations(),
		})
	}

	if g.report.Summary.MobileAvgAccessibility < float64(g.thresholds.Accessibility) {
		problems = append(problems, ProblemSummary{
			Category: "Accessibility", Device: "Mobile",
			CurrentScore: g.report.Summary.MobileAvgAccessibility,
			TargetScore:  float64(g.thresholds.Accessibility),
			Gap:          float64(g.thresholds.Accessibility) - g.report.Summary.MobileAvgAccessibility,
			Recommendations: getAccessibilityRecommendations(),
		})
	}

	sort.Slice(problems, func(i, j int) bool {
		return problems[i].Gap > problems[j].Gap
	})

	return problems
}

func getPerformanceRecommendations() []string {
	return []string{
		"Optimize images with WebP format and lazy loading",
		"Implement code splitting and tree shaking",
		"Enable HTTP/2 and compression (gzip/brotli)",
		"Minimize main-thread work and JavaScript execution time",
		"Use font-display: swap for web fonts",
		"Implement resource hints (preconnect, prefetch)",
	}
}

func getAccessibilityRecommendations() []string {
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
