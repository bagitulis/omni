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
	"strconv"
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

// lazadaBaseResponse captures Lazada business-level error fields returned in 200 OK responses.
type lazadaBaseResponse struct {
	Code      string `json:"code"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// truncateString truncates a string to maxLen characters.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}

// RawRequest executes an arbitrary Lazada REST call with proper signing.
// Rebuilds timestamp and signature per retry attempt.
// Checks both HTTP-level and Lazada business-level errors.
func (c *Client) RawRequest(ctx context.Context, method, apiPath string, params map[string]string, result interface{}) error {
	if ctx == nil {
		return fmt.Errorf("context is required")
	}
	if params == nil {
		params = map[string]string{}
	}

	baseURL := c.baseURLFor(apiPath)

	var resp *http.Response
	var err error
	var lastErr error
	lastStatusCode := 0
	maxRetries := 5
	var retryAfterDuration time.Duration

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			if retryAfterDuration > 0 {
				log.Info().Int("attempt", attempt+1).Dur("retry_after", retryAfterDuration).Msg("[Lazada API] Respecting Retry-After header")
				time.Sleep(retryAfterDuration)
				retryAfterDuration = 0
			} else {
				backoff := time.Duration(1<<uint(attempt)) * 200 * time.Millisecond
				if lastStatusCode == 429 {
					backoff = time.Duration(1<<uint(attempt)) * time.Second
				}
				log.Info().Int("attempt", attempt+1).Dur("backoff", backoff).Msg("[Lazada API] Retrying request")
				time.Sleep(backoff)
			}
		}

		// Rebuild timestamp + signature each attempt so they are fresh
		params["app_key"] = c.appKey
		params["timestamp"] = fmt.Sprintf("%d", time.Now().UnixMilli())
		params["sign_method"] = "sha256"

		if c.accessToken != "" && !strings.HasPrefix(apiPath, "/auth/") && !strings.HasPrefix(apiPath, "/datamoat/") {
			params["access_token"] = c.accessToken
		}

		params["sign"] = c.generateSign(params, apiPath)

		u, parseErr := url.Parse(baseURL + apiPath)
		if parseErr != nil {
			return fmt.Errorf("failed to parse URL %s%s: %w", baseURL, apiPath, parseErr)
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

		if resp.StatusCode == 429 || resp.StatusCode >= 500 {
			if resp.StatusCode == 429 {
				if ra := resp.Header.Get("Retry-After"); ra != "" {
					if seconds, parseErr := strconv.Atoi(ra); parseErr == nil {
						retryAfterDuration = time.Duration(seconds) * time.Second
					}
				}
			}
			resp.Body.Close()
			if attempt < maxRetries {
				log.Warn().Int("status", resp.StatusCode).Msg("[Lazada API] Rate limit or server error, retrying")
				continue
			}
		}

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
		log.Warn().
			Int("http_status", resp.StatusCode).
			Str("api_path", apiPath).
			Str("raw_response", truncateString(string(respBody), 2000)).
			Msg("[Lazada API] HTTP Error")
		return fmt.Errorf("lazada API error [http_status=%d]: %s", resp.StatusCode, string(respBody))
	}

	// Check Lazada business-level error (returned with HTTP 200)
	var baseResp lazadaBaseResponse
	if jsonErr := json.Unmarshal(respBody, &baseResp); jsonErr == nil && baseResp.Code != "" && baseResp.Code != "0" {
		log.Warn().
			Str("api_path", apiPath).
			Str("request_id", baseResp.RequestID).
			Str("code", baseResp.Code).
			Str("type", baseResp.Type).
			Str("message", baseResp.Message).
			Str("raw_response", truncateString(string(respBody), 2000)).
			Msg("[Lazada API] Business Error")
		return fmt.Errorf("lazada API error [request_id=%s]: code=%s, type=%s, message=%s", baseResp.RequestID, baseResp.Code, baseResp.Type, baseResp.Message)
	}

	// Success - log request_id at Debug level
	var successResp lazadaBaseResponse
	if jsonErr := json.Unmarshal(respBody, &successResp); jsonErr == nil {
		log.Debug().
			Str("api_path", apiPath).
			Str("request_id", successResp.RequestID).
			Msg("[Lazada API] Success")
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
