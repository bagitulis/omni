package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// OmniClient wraps HTTP communication with the omni backend
type OmniClient struct {
	baseURL       string
	defaultTenant string
	token         string
	http          *http.Client
}

// NewOmniClient reads config from environment variables
func NewOmniClient() (*OmniClient, error) {
	baseURL := os.Getenv("OMNI_API_URL")
	if baseURL == "" {
		baseURL = "http://localhost:3000"
	}
	tenant := os.Getenv("OMNI_DEFAULT_TENANT")
	if tenant == "" {
		tenant = "yumna_bertigamart"
	}

	return &OmniClient{
		baseURL:       baseURL,
		defaultTenant: tenant,
		http:          &http.Client{Timeout: 120 * time.Second},
	}, nil
}

// get performs an authenticated GET request
func (c *OmniClient) get(path string) ([]byte, error) {
	return c.doRequest("GET", path, nil)
}

// post performs an authenticated POST request with JSON body
func (c *OmniClient) post(path string, body interface{}) ([]byte, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBytes, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBytes)
	}
	return c.doRequest("POST", path, bodyReader)
}

// put performs an authenticated PUT request with JSON body
func (c *OmniClient) put(path string, body interface{}) ([]byte, error) {
	jsonBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal body: %w", err)
	}
	return c.doRequest("PUT", path, bytes.NewReader(jsonBytes))
}

// getNoTenant performs an authenticated GET without tenant header (for monitoring endpoints)
func (c *OmniClient) getNoTenant(path string) ([]byte, error) {
	return c.doRequestNoTenant("GET", path, nil)
}

// doRequest executes an HTTP request with automatic auth and 401 retry
func (c *OmniClient) doRequest(method, path string, body io.Reader) ([]byte, error) {
	if c.token == "" {
		if err := c.login(); err != nil {
			return nil, fmt.Errorf("login failed: %w", err)
		}
	}

	// Buffer body so we can replay on 401 retry
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
	}

	resp, err := c.executeRequest(method, path, bodyBytes)
	if err != nil {
		return nil, err
	}

	// Retry once on 401 (token expired)
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		if err := c.login(); err != nil {
			return nil, fmt.Errorf("re-login failed: %w", err)
		}
		resp, err = c.executeRequest(method, path, bodyBytes)
		if err != nil {
			return nil, err
		}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// doRequestNoTenant executes an HTTP request without X-Tenant-ID header
func (c *OmniClient) doRequestNoTenant(method, path string, body io.Reader) ([]byte, error) {
	if c.token == "" {
		if err := c.login(); err != nil {
			return nil, fmt.Errorf("login failed: %w", err)
		}
	}

	url := c.baseURL + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	// No X-Tenant-ID header

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}

	// Retry once on 401 (token expired)
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close()
		if err := c.login(); err != nil {
			return nil, fmt.Errorf("re-login failed: %w", err)
		}
		req2, err := http.NewRequest(method, url, body)
		if err != nil {
			return nil, fmt.Errorf("create retry request: %w", err)
		}
		req2.Header.Set("Authorization", "Bearer "+c.token)
		req2.Header.Set("Content-Type", "application/json")
		resp, err = c.http.Do(req2)
		if err != nil {
			return nil, fmt.Errorf("retry request: %w", err)
		}
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(data))
	}

	return data, nil
}

// executeRequest builds and sends a single HTTP request
func (c *OmniClient) executeRequest(method, path string, bodyBytes []byte) (*http.Response, error) {
	url := c.baseURL + path
	var bodyReader io.Reader
	if bodyBytes != nil {
		bodyReader = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", c.defaultTenant)

	return c.http.Do(req)
}
