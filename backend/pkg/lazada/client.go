package lazada

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	BaseURLMY = "https://api.lazada.com.my/rest"
	BaseURLSG = "https://api.lazada.sg/rest"
	BaseURLID = "https://api.lazada.co.id/rest"
	BaseURLTH = "https://api.lazada.co.th/rest"
	BaseURLVN = "https://api.lazada.vn/rest"
	BaseURLPH = "https://api.lazada.com.ph/rest"

	// BaseURLCommon is used for a small subset of APIs that hit the global host,
	// e.g. DataMoat endpoints.
	BaseURLCommon = "https://api.lazada.com/rest"

	// BaseURLAuth is used for token exchange/refresh.
	BaseURLAuth = "https://auth.lazada.com/rest"
)

// Client is the Lazada API client
type Client struct {
	appKey      string
	appSecret   string
	accessToken string
	baseURL     string
	httpClient  *http.Client
}

// NewClient creates a new Lazada API client
func NewClient(appKey, appSecret, region string) *Client {
	baseURL := BaseURLMY
	switch strings.ToLower(region) {
	case "sg":
		baseURL = BaseURLSG
	case "id":
		baseURL = BaseURLID
	case "th":
		baseURL = BaseURLTH
	case "vn":
		baseURL = BaseURLVN
	case "ph":
		baseURL = BaseURLPH
	case "my":
		baseURL = BaseURLMY
	}

	return &Client{
		appKey:     appKey,
		appSecret:  appSecret,
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// SetAccessToken sets the access token
func (c *Client) SetAccessToken(token string) {
	c.accessToken = token
}

func (c *Client) baseURLFor(apiPath string) string {
	switch {
	case strings.HasPrefix(apiPath, "/auth/"):
		return BaseURLAuth
	case strings.HasPrefix(apiPath, "/datamoat/"):
		return BaseURLCommon
	default:
		return c.baseURL
	}
}

// generateSign creates signature for Lazada API
// Per official Lazada SDK:
// 1. Sort parameters alphabetically (excluding 'sign')
// 2. Concatenate params as: key1value1key2value2...
// 3. Prepend API path: /api/path + paramString
// 4. HMAC-SHA256(appSecret, full_string)
// 5. Convert to UPPERCASE hex
func (c *Client) generateSign(params map[string]string, apiPath string) string {
	// Sort keys, excluding 'sign'
	keys := make([]string, 0, len(params))
	for k := range params {
		if k != "sign" { // Exclude sign key
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// Build string to sign: path + key1value1key2value2...
	var sb strings.Builder
	sb.WriteString(apiPath)
	for _, k := range keys {
		sb.WriteString(k)
		sb.WriteString(params[k])
	}

	// HMAC-SHA256
	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(sb.String()))
	return strings.ToUpper(hex.EncodeToString(h.Sum(nil)))
}

// RawRequest executes an arbitrary Lazada REST call with proper signing.
// For GET calls, parameters are sent on the query string. For POST calls, the
// parameters are sent as x-www-form-urlencoded body (matching the official SDK
// behavior and avoiding URL-length issues when payload is large).
func (c *Client) RawRequest(ctx context.Context, method, apiPath string, params map[string]string, result interface{}) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	if params == nil {
		params = map[string]string{}
	}

	// Build request params (mutates the local map).
	timestamp := fmt.Sprintf("%d", time.Now().UnixMilli())
	params["app_key"] = c.appKey
	params["timestamp"] = timestamp
	params["sign_method"] = "sha256"

	// Auth endpoints and DataMoat endpoints do not use access_token.
	if c.accessToken != "" && !strings.HasPrefix(apiPath, "/auth/") && !strings.HasPrefix(apiPath, "/datamoat/") {
		params["access_token"] = c.accessToken
	}

	// Generate signature.
	params["sign"] = c.generateSign(params, apiPath)

	baseURL := c.baseURLFor(apiPath)
	u, _ := url.Parse(baseURL + apiPath)

	var resp *http.Response
	var err error
	var lastErr error
	lastStatusCode := 0
	maxRetries := 3

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff: 200ms, 400ms, 800ms
			backoff := time.Duration(1<<uint(attempt)) * 100 * time.Millisecond
			time.Sleep(backoff)
		}

		var body io.Reader
		if strings.EqualFold(method, http.MethodGet) {
			q := u.Query()
			for k, v := range params {
				q.Set(k, v)
			}
			u.RawQuery = q.Encode()
		} else {
			form := url.Values{}
			for k, v := range params {
				form.Set(k, v)
			}
			body = strings.NewReader(form.Encode())
		}

		req, reqErr := http.NewRequestWithContext(ctx, method, u.String(), body)
		if reqErr != nil {
			return reqErr
		}
		if !strings.EqualFold(method, http.MethodGet) {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8")
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < maxRetries {
				continue
			}
			return err
		}
		lastStatusCode = resp.StatusCode

		// Check for rate limit (429) or server errors (5xx)
		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			resp.Body.Close()
			if attempt < maxRetries {
				continue
			}
		}

		// Success or non-retriable error
		break
	}

	if resp == nil {
		if lastErr != nil {
			return fmt.Errorf("lazada API request failed after %d retries: %w", maxRetries, lastErr)
		}
		return fmt.Errorf("lazada API request failed after %d retries [http_status=%d]", maxRetries, lastStatusCode)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return fmt.Errorf("lazada API error [http_status=%d]: %s", resp.StatusCode, string(respBody))
	}
	if result == nil {
		return nil
	}
	return json.Unmarshal(respBody, result)
}

// RawGet executes a signed GET request.
func (c *Client) RawGet(ctx context.Context, apiPath string, params map[string]string, result interface{}) error {
	return c.RawRequest(ctx, http.MethodGet, apiPath, params, result)
}

// RawPost executes a signed POST request with x-www-form-urlencoded body.
func (c *Client) RawPost(ctx context.Context, apiPath string, params map[string]string, result interface{}) error {
	return c.RawRequest(ctx, http.MethodPost, apiPath, params, result)
}

// doRequest executes HTTP request with structured logging.
func (c *Client) doRequest(method, apiPath string, params map[string]string, result interface{}) error {
	log.Debug().
		Str("method", method).
		Str("api_path", apiPath).
		Msg("[Lazada API] Request")

	return c.RawRequest(context.Background(), method, apiPath, params, result)
}
