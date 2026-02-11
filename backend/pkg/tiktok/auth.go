package tiktok

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
func CreateAccessToken(appKey, appSecret, authCode string) (*TokenResponse, error) {
	url := fmt.Sprintf("%s/api/v2/token/get?app_key=%s&app_secret=%s&auth_code=%s&grant_type=authorized_code",
		BaseURL, appKey, appSecret, authCode)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result TokenResponse
	json.Unmarshal(body, &result)
	return &result, nil
}
