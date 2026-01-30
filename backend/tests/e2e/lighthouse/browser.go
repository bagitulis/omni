// Package lighthouse provides Lighthouse performance testing for the OMNI frontend
package lighthouse

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// Browser handles browser automation using chromedp
type Browser struct {
	config     *Config
	allocCtx   context.Context
	cancelFunc context.CancelFunc
	ctx        context.Context
	cookies    []*network.Cookie
}

// NewBrowser creates a new browser instance
func NewBrowser(config *Config) *Browser {
	return &Browser{
		config: config,
	}
}

// Init initializes the browser
func (b *Browser) Init() error {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(1920, 1080),
	)

	if b.config.ChromePath != "" {
		opts = append(opts, chromedp.ExecPath(b.config.ChromePath))
	}

	b.allocCtx, b.cancelFunc = chromedp.NewExecAllocator(context.Background(), opts...)
	b.ctx, _ = chromedp.NewContext(b.allocCtx, chromedp.WithLogf(func(format string, args ...interface{}) {
		// Silent logging - comment out for debugging
		// fmt.Printf(format+"\n", args...)
	}))

	// Ensure directories exist
	if err := os.MkdirAll(b.config.ResultsDir, 0755); err != nil {
		return fmt.Errorf("failed to create results dir: %w", err)
	}
	if err := os.MkdirAll(b.config.ScreenshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshots dir: %w", err)
	}

	return nil
}

// Close closes the browser
func (b *Browser) Close() {
	if b.cancelFunc != nil {
		b.cancelFunc()
	}
}

// Login performs login and stores cookies
func (b *Browser) Login() (*LoginResult, error) {
	result := &LoginResult{
		Success: false,
	}
	startTime := time.Now()

	loginURL := fmt.Sprintf("%s/login", b.config.FrontendURL)
	fmt.Printf("   Navigating to: %s\n", loginURL)

	// Create a timeout context
	ctx, cancel := context.WithTimeout(b.ctx, 60*time.Second)
	defer cancel()

	// Step 1: Navigate and wait for page load
	fmt.Println("   Waiting for login page...")
	err := chromedp.Run(ctx,
		chromedp.Navigate(loginURL),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
	)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to navigate to login page: %s", err.Error())
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Step 2: Wait for username field
	fmt.Println("   Waiting for username field...")
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(`#username`, chromedp.ByID),
	)
	if err != nil {
		result.Error = fmt.Sprintf("Username field not visible: %s", err.Error())
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Step 3: Fill login form
	fmt.Println("   Filling login form...")
	err = chromedp.Run(ctx,
		// Clear and fill username using JavaScript for reliability
		chromedp.Evaluate(`document.getElementById('username').value = ''`, nil),
		chromedp.Evaluate(fmt.Sprintf(`document.getElementById('username').value = '%s'`, b.config.Username), nil),
		chromedp.Evaluate(`document.getElementById('username').dispatchEvent(new Event('input', { bubbles: true }))`, nil),

		// Clear and fill password
		chromedp.Evaluate(`document.getElementById('password').value = ''`, nil),
		chromedp.Evaluate(fmt.Sprintf(`document.getElementById('password').value = '%s'`, b.config.Password), nil),
		chromedp.Evaluate(`document.getElementById('password').dispatchEvent(new Event('input', { bubbles: true }))`, nil),

		chromedp.Sleep(500*time.Millisecond),
	)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to fill form: %s", err.Error())
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Step 4: Click submit button
	fmt.Println("   Clicking submit button...")
	err = chromedp.Run(ctx,
		chromedp.Click(`button[type="submit"]`, chromedp.NodeVisible),
		chromedp.Sleep(5*time.Second), // Wait for login to complete
	)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to click submit: %s", err.Error())
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	// Step 5: Check current URL
	var currentURL string
	err = chromedp.Run(ctx,
		chromedp.Location(&currentURL),
	)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get URL: %s", err.Error())
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	fmt.Printf("   Current URL after login: %s\n", currentURL)

	// Check if login was successful
	if strings.Contains(currentURL, "/login") {
		// Try to get error message
		var errorMsg string
		chromedp.Run(ctx,
			chromedp.Text(`.error-message`, &errorMsg, chromedp.ByQuery),
		)
		if errorMsg != "" {
			result.Error = fmt.Sprintf("Login failed: %s", errorMsg)
		} else {
			result.Error = "Login failed - still on login page"
		}
		result.Duration = time.Since(startTime).Milliseconds()
		return result, fmt.Errorf("%s", result.Error)
	}

	// Step 6: Store cookies
	fmt.Println("   Extracting cookies...")
	err = chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			cookies, err := network.GetCookies().Do(ctx)
			if err != nil {
				return err
			}
			b.cookies = cookies
			fmt.Printf("   Got %d cookies\n", len(cookies))
			return nil
		}),
	)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to get cookies: %s", err.Error())
		result.Duration = time.Since(startTime).Milliseconds()
		return result, err
	}

	result.Success = true
	result.Duration = time.Since(startTime).Milliseconds()
	return result, nil
}

// GetCookies returns the stored cookies
func (b *Browser) GetCookies() []*network.Cookie {
	return b.cookies
}

// GetCookiesAsHeader returns cookies formatted for HTTP header
func (b *Browser) GetCookiesAsHeader() string {
	if len(b.cookies) == 0 {
		return ""
	}

	var parts []string
	for _, c := range b.cookies {
		parts = append(parts, fmt.Sprintf("%s=%s", c.Name, c.Value))
	}
	return strings.Join(parts, "; ")
}

// TakeScreenshot takes a screenshot of the current page
func (b *Browser) TakeScreenshot(url string, name string, isMobile bool) (string, error) {
	viewport := ViewportDesktop()
	prefix := "desktop"
	if isMobile {
		viewport = ViewportMobile()
		prefix = "mobile"
	}

	safeName := strings.ToLower(name)
	safeName = strings.ReplaceAll(safeName, " ", "-")
	safeName = strings.ReplaceAll(safeName, "/", "-")
	filename := fmt.Sprintf("%s-%s.png", prefix, safeName)
	screenshotPath := filepath.Join(b.config.ScreenshotsDir, filename)

	ctx, cancel := context.WithTimeout(b.ctx, 30*time.Second)
	defer cancel()

	var buf []byte
	err := chromedp.Run(ctx,
		chromedp.EmulateViewport(int64(viewport.Width), int64(viewport.Height)),
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
		chromedp.ActionFunc(func(ctx context.Context) error {
			if isMobile {
				return b.closeMobileSidebar(ctx)
			}
			return nil
		}),
		chromedp.Sleep(500*time.Millisecond),
		chromedp.FullScreenshot(&buf, 90),
	)

	if err != nil {
		return "", fmt.Errorf("failed to take screenshot: %w", err)
	}

	if err := os.WriteFile(screenshotPath, buf, 0644); err != nil {
		return "", fmt.Errorf("failed to save screenshot: %w", err)
	}

	return filename, nil
}

// closeMobileSidebar closes the sidebar on mobile view
func (b *Browser) closeMobileSidebar(ctx context.Context) error {
	script := `
		// Handle LeftSidebar component
		const sidebar = document.querySelector('.left-sidebar');
		if (sidebar) {
			sidebar.classList.add('collapsed');
			sidebar.style.width = '0px';
			sidebar.style.transform = 'translateX(-100%)';
			sidebar.style.opacity = '0';
			sidebar.style.pointerEvents = 'none';
		}

		// Hide overlay
		const overlay = document.querySelector('.sidebar-overlay');
		if (overlay) {
			overlay.style.display = 'none';
			overlay.style.opacity = '0';
		}

		// Handle sidebar wrapper
		const wrapper = document.querySelector('.left-sidebar-wrapper');
		if (wrapper) {
			wrapper.style.width = '0px';
			wrapper.style.minWidth = '0px';
		}

		// Ensure main content takes full width
		const mainContent = document.querySelector('.main-content, main, [class*="main-content"]');
		if (mainContent) {
			mainContent.style.marginLeft = '0';
			mainContent.style.width = '100%';
		}
	`

	return chromedp.Evaluate(script, nil).Do(ctx)
}

// NavigateTo navigates to a specific URL
func (b *Browser) NavigateTo(url string) error {
	ctx, cancel := context.WithTimeout(b.ctx, 30*time.Second)
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.WaitReady("body"),
		chromedp.Sleep(2*time.Second),
	)
}

// WaitForLoad waits for page to load
func (b *Browser) WaitForLoad() error {
	return chromedp.Run(b.ctx,
		chromedp.WaitReady("body"),
	)
}
