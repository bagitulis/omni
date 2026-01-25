package lazada

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
func CreateAccessToken(appKey, appSecret, code, region string) (*TokenResponse, error) {
	baseURL := "https://auth.lazada.com/rest"
	if region == "my" {
		baseURL = "https://auth.lazada.com.my/rest"
	}

	url := fmt.Sprintf("%s/auth/token/create?app_key=%s&app_secret=%s&code=%s",
		baseURL, appKey, appSecret, code)

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
