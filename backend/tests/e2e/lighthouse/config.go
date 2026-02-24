// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse

import "os"

// getEnvOrDefault returns the value of the environment variable named by the key,
// or defaultVal if the variable is not set or empty.
func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// Config holds configuration for Lighthouse testing
type Config struct {
	BaseURL        string
	FrontendURL    string
	Username       string
	Password       string
	ResultsDir     string
	ScreenshotsDir string
	ChromePath     string
	Timeout        int // seconds
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		BaseURL:        "http://localhost:3000",
		FrontendURL:    "http://localhost:5174",
		Username:       getEnvOrDefault("LIGHTHOUSE_USERNAME", "admin"),
		Password:       getEnvOrDefault("LIGHTHOUSE_PASSWORD", "password"),
		ResultsDir:     "test-results",
		ScreenshotsDir: "test-results/lighthouse-screenshots",
		ChromePath:     "", // Use system default
		Timeout:        60,
	}
}

// Viewport represents screen dimensions
type Viewport struct {
	Width  int
	Height int
}

// ViewportDesktop returns desktop viewport dimensions
func ViewportDesktop() Viewport {
	return Viewport{Width: 1920, Height: 1080}
}

// ViewportMobile returns mobile viewport dimensions
func ViewportMobile() Viewport {
	return Viewport{Width: 375, Height: 812}
}

// LighthouseConfig represents Lighthouse CLI configuration
type LighthouseConfig struct {
	FormFactor      string
	ScreenEmulation ScreenEmulation
	Throttling      Throttling
	Categories      []string
}

// ScreenEmulation represents screen emulation settings
type ScreenEmulation struct {
	Mobile            bool
	Width             int
	Height            int
	DeviceScaleFactor float64
	Disabled          bool
}

// Throttling represents network/CPU throttling settings
type Throttling struct {
	RTTMs               int
	ThroughputKbps      float64
	CPUSlowdownMultiple float64
}

// DesktopLighthouseConfig returns Lighthouse config for desktop
func DesktopLighthouseConfig() LighthouseConfig {
	return LighthouseConfig{
		FormFactor: "desktop",
		ScreenEmulation: ScreenEmulation{
			Mobile:            false,
			Width:             1350,
			Height:            940,
			DeviceScaleFactor: 1,
			Disabled:          false,
		},
		Throttling: Throttling{
			RTTMs:               40,
			ThroughputKbps:      10240,
			CPUSlowdownMultiple: 1,
		},
		Categories: []string{"performance", "accessibility", "best-practices", "seo"},
	}
}

// MobileLighthouseConfig returns Lighthouse config for mobile
func MobileLighthouseConfig() LighthouseConfig {
	return LighthouseConfig{
		FormFactor: "mobile",
		ScreenEmulation: ScreenEmulation{
			Mobile:            true,
			Width:             375,
			Height:            812,
			DeviceScaleFactor: 2,
			Disabled:          false,
		},
		Throttling: Throttling{
			RTTMs:               150,
			ThroughputKbps:      1638.4,
			CPUSlowdownMultiple: 4,
		},
		Categories: []string{"performance", "accessibility", "best-practices", "seo"},
	}
}

// ScoreThresholds defines minimum acceptable scores
type ScoreThresholds struct {
	Performance   int
	Accessibility int
	BestPractices int
	SEO           int
}

// DefaultThresholds returns default score thresholds (95% target)
func DefaultThresholds() ScoreThresholds {
	return ScoreThresholds{
		Performance:   95,
		Accessibility: 95,
		BestPractices: 95,
		SEO:           95,
	}
}

// StrictThresholds returns strict score thresholds (100% target)
func StrictThresholds() ScoreThresholds {
	return ScoreThresholds{
		Performance:   100,
		Accessibility: 100,
		BestPractices: 100,
		SEO:           100,
	}
}

// InternalAppThresholds returns thresholds for internal applications
// SEO is lower because internal apps intentionally use noindex meta tags
// to prevent search engine indexing (security best practice)
func InternalAppThresholds() ScoreThresholds {
	return ScoreThresholds{
		Performance:   80,
		Accessibility: 90,
		BestPractices: 90,
		SEO:           60, // Expected: noindex causes ~30-40 point penalty
	}
}
