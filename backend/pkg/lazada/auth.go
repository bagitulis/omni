package lazada

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httputils "github.com/omni/backend/internal/utils/http"
)

// TokenResponse represents Lazada OAuth token response
type TokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	ExpiresIn        int64  `json:"expires_in"`
	Country          string `json:"country"`
	Code             string `json:"code"`
	Message          string `json:"message"`
}

// GetAccessToken exchanges auth code for access token
func (c *Client) GetAccessToken(code string) (*TokenResponse, error) {
	params := map[string]string{
		"code":       code,
		"grant_type": "authorization_code",
	}

	var result TokenResponse
	err := c.doRequest("POST", "/auth/token/create", params, &result)
	return &result, err
}

// RefreshAccessToken refreshes the access token
func (c *Client) RefreshAccessToken(refreshToken string) (*TokenResponse, error) {
	params := map[string]string{
		"refresh_token": refreshToken,
		"grant_type":    "refresh_token",
	}

	var result TokenResponse
	err := c.doRequest("POST", "/auth/token/refresh", params, &result)
	return &result, err
}

// CreateAccessToken creates token using auth code (direct HTTP call)
func CreateAccessToken(ctx context.Context, appKey, appSecret, code, region string) (*TokenResponse, error) {
	baseURL := "https://auth.lazada.com/rest"
	if region == "my" {
		baseURL = "https://auth.lazada.com.my/rest"
	}

	rawURL := fmt.Sprintf("%s/auth/token/create?app_key=%s&app_secret=%s&code=%s",
		baseURL, appKey, appSecret, code)

	body, statusCode, err := httputils.SecureGetAPI(ctx, rawURL)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: status %d, body: %s", statusCode, string(body))
	}

	var result TokenResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse token response: %w", err)
	}
	return &result, nil
}
