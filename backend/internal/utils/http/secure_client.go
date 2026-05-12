package httputils

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	DefaultTimeout     = 30 * time.Second
	MaxResponseBody    = 50 * 1024 * 1024 // 50MB for file downloads
	MaxAPIResponseBody = 10 * 1024 * 1024 // 10MB for API responses
)

// AllowedHosts defines permitted external hosts for HTTP requests.
var AllowedHosts = map[string]bool{
	"partner.shopeemobile.com":             true,
	"partner.test-stable.shopeemobile.com": true,
	"api.lazada.com":                       true,
	"auth.lazada.com":                      true,
	"auth.lazada.com.my":                   true,
	"open-api.tiktokglobalshop.com":        true,
	"auth.tiktok-shops.com":                true,
	"accounts.google.com":                  true,
	"oauth2.googleapis.com":                true,
	"www.googleapis.com":                   true,
	"sheets.googleapis.com":                true,
}

// SecureGet performs a GET request with timeout, host validation, and body size limit.
func SecureGet(ctx context.Context, rawURL string, maxBody int64) ([]byte, int, error) {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, 0, fmt.Errorf("invalid URL: %w", err)
	}

	host := strings.ToLower(parsed.Hostname())
	if !AllowedHosts[host] {
		return nil, 0, fmt.Errorf("host not allowed: %s", host)
	}

	if parsed.Scheme != "https" {
		return nil, 0, fmt.Errorf("only HTTPS allowed, got: %s", parsed.Scheme)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	client := &http.Client{Timeout: DefaultTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxBody))
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read body: %w", err)
	}

	return body, resp.StatusCode, nil
}

// SecureGetAPI is a convenience wrapper for API calls (10MB limit).
func SecureGetAPI(ctx context.Context, rawURL string) ([]byte, int, error) {
	return SecureGet(ctx, rawURL, MaxAPIResponseBody)
}
