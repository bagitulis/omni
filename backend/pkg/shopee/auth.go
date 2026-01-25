package shopee

import (
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// TokenResponse represents Shopee OAuth token response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpireIn     int64  `json:"expire_in"`
	ShopIDList   []int64 `json:"shop_id_list"`
	Error        string `json:"error"`
	Message      string `json:"message"`
}

// GetAccessToken exchanges auth code for access token
func (c *Client) GetAccessToken(code, shopID string) (*TokenResponse, error) {
	timestamp := time.Now().Unix()
	path := "/api/v2/auth/token/get"
	sign := c.generateSign(path, timestamp)

	url := fmt.Sprintf("%s%s?partner_id=%d&timestamp=%d&sign=%s&code=%s&shop_id=%s",
		c.baseURL, path, c.partnerID, timestamp, sign, code, shopID)

	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result TokenResponse
	json.Unmarshal(body, &result)
	return &result, nil
}

// RefreshAccessToken refreshes the access token
func (c *Client) RefreshAccessToken(refreshToken string, shopID int64) (*TokenResponse, error) {
	timestamp := time.Now().Unix()
	path := "/api/v2/auth/access_token/get"
	sign := c.generateSign(path, timestamp)

	url := fmt.Sprintf("%s%s?partner_id=%d&timestamp=%d&sign=%s&refresh_token=%s&shop_id=%d",
		c.baseURL, path, c.partnerID, timestamp, sign, refreshToken, shopID)

	resp, err := c.httpClient.Post(url, "application/json", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result TokenResponse
	json.Unmarshal(body, &result)
	return &result, nil
}
