// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse

import "time"

// Route represents a page route to test
type Route struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

// Scores represents Lighthouse category scores (0-100)
type Scores struct {
	Performance   int `json:"performance"`
	Accessibility int `json:"accessibility"`
	BestPractices int `json:"best_practices"`
	SEO           int `json:"seo"`
}

// Metrics represents Core Web Vitals and other performance metrics
type Metrics struct {
	FirstContentfulPaint   float64 `json:"first_contentful_paint"`
	LargestContentfulPaint float64 `json:"largest_contentful_paint"`
	TotalBlockingTime      float64 `json:"total_blocking_time"`
	CumulativeLayoutShift  float64 `json:"cumulative_layout_shift"`
	SpeedIndex             float64 `json:"speed_index"`
	TimeToInteractive      float64 `json:"time_to_interactive"`
}

// Issue represents a Lighthouse audit issue
type Issue struct {
	ID           string  `json:"id"`
	Title        string  `json:"title"`
	Description  string  `json:"description"`
	Score        float64 `json:"score"`
	DisplayValue string  `json:"display_value,omitempty"`
	Impact       string  `json:"impact,omitempty"`
}

// Issues categorizes issues by Lighthouse category
type Issues struct {
	Performance   []Issue `json:"performance"`
	Accessibility []Issue `json:"accessibility"`
	BestPractices []Issue `json:"best_practices"`
	SEO           []Issue `json:"seo"`
}

// DeviceResult represents test results for a single device type
type DeviceResult struct {
	Success    bool          `json:"success"`
	Duration   time.Duration `json:"duration_ms"`
	Scores     Scores        `json:"scores"`
	Metrics    Metrics       `json:"metrics"`
	Issues     Issues        `json:"issues"`
	Screenshot string        `json:"screenshot,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// PageResult represents test results for a single page
type PageResult struct {
	Name    string        `json:"name"`
	Path    string        `json:"path"`
	URL     string        `json:"url"`
	Desktop *DeviceResult `json:"desktop,omitempty"`
	Mobile  *DeviceResult `json:"mobile,omitempty"`
}

// Summary represents aggregated test statistics
type Summary struct {
	TotalPages              int     `json:"total_pages"`
	DesktopAvgPerformance   float64 `json:"desktop_avg_performance"`
	DesktopAvgAccessibility float64 `json:"desktop_avg_accessibility"`
	DesktopAvgBestPractices float64 `json:"desktop_avg_best_practices"`
	DesktopAvgSEO           float64 `json:"desktop_avg_seo"`
	MobileAvgPerformance    float64 `json:"mobile_avg_performance"`
	MobileAvgAccessibility  float64 `json:"mobile_avg_accessibility"`
	MobileAvgBestPractices  float64 `json:"mobile_avg_best_practices"`
	MobileAvgSEO            float64 `json:"mobile_avg_seo"`
	TotalScreenshots        int     `json:"total_screenshots"`
	TotalDuration           int64   `json:"total_duration_ms"`
}

// TestInfo contains metadata about the test run
type TestInfo struct {
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Duration    int64     `json:"duration_ms"`
	TestType    string    `json:"test_type"`
	BaseURL     string    `json:"base_url"`
	Credentials struct {
		Username string `json:"username"`
	} `json:"credentials"`
}

// FullReport represents the complete Lighthouse test report
type FullReport struct {
	TestInfo     TestInfo      `json:"test_info"`
	LoginTest    LoginResult   `json:"login_test"`
	PageTests    []PageResult  `json:"page_tests"`
	Summary      Summary       `json:"summary"`
	CommonIssues []CommonIssue `json:"common_issues"`
}

// LoginResult represents the result of the login test
type LoginResult struct {
	Success  bool   `json:"success"`
	Duration int64  `json:"duration_ms"`
	Error    string `json:"error,omitempty"`
}

// CommonIssue represents an issue that appears across multiple pages
type CommonIssue struct {
	ID             string   `json:"id"`
	Title          string   `json:"title"`
	Category       string   `json:"category"`
	Impact         string   `json:"impact"`
	Occurrences    int      `json:"occurrences"`
	AffectedPages  []string `json:"affected_pages"`
	Recommendation string   `json:"recommendation"`
}

// LighthouseJSONReport represents the raw Lighthouse JSON output
type LighthouseJSONReport struct {
	Categories map[string]struct {
		Score float64 `json:"score"`
	} `json:"categories"`
	Audits map[string]struct {
		ID           string   `json:"id"`
		Title        string   `json:"title"`
		Description  string   `json:"description"`
		Score        *float64 `json:"score"`
		DisplayValue string   `json:"displayValue"`
		NumericValue float64  `json:"numericValue"`
	} `json:"audits"`
}
