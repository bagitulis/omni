package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	httputils "github.com/omni/backend/internal/utils/http"
)

// TokenResponse represents TikTok OAuth token response
type TokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		AccessToken        string `json:"access_token"`
		RefreshToken       string `json:"refresh_token"`
		AccessTokenExpire  int64  `json:"access_token_expire_in"`
		RefreshTokenExpire int64  `json:"refresh_token_expire_in"`
		OpenID             string `json:"open_id"`
		SellerName         string `json:"seller_name"`
	} `json:"data"`
}

// GetAccessToken exchanges auth code for access token
func (c *Client) GetAccessToken(authCode string) (*TokenResponse, error) {
	params := map[string]string{
		"auth_code":  authCode,
		"grant_type": "authorized_code",
	}

	var result TokenResponse
	err := c.doRequest("POST", "/api/v2/token/get", params, &result)
	return &result, err
}

// RefreshAccessToken refreshes the access token
func (c *Client) RefreshAccessToken(refreshToken string) (*TokenResponse, error) {
	params := map[string]string{
		"refresh_token": refreshToken,
		"grant_type":    "refresh_token",
	}

	var result TokenResponse
	err := c.doRequest("POST", "/api/v2/token/refresh", params, &result)
	return &result, err
}

// CreateAccessToken creates token (direct HTTP call for initial auth)
func CreateAccessToken(ctx context.Context, appKey, appSecret, authCode string) (*TokenResponse, error) {
	rawURL := fmt.Sprintf("%s/api/v2/token/get?app_key=%s&app_secret=%s&auth_code=%s&grant_type=authorized_code",
		BaseURL, appKey, appSecret, authCode)

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
