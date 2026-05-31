package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/omni/backend/internal/services/oauth"
)

// NOTE: TikTok's RefreshToken API discards the new refresh_token (only returns access_token).
// The original refresh_token has a ~365 day expiry, which is the effective connection lifetime.
// After 365 days, the user must re-authorize via full OAuth flow.

// tiktokTokenResponse represents the TikTok token exchange response.
type tiktokTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	ErrorCode        int    `json:"code"`
	ErrorMsg         string `json:"message"`
}

// exchangeTiktokToken performs the HTTP GET to exchange an auth code for tokens.
// TikTok uses GET with query parameters (not POST like Shopee).
func exchangeTiktokToken(ctx context.Context, tiktokService *oauth.TiktokOAuthService, code string) (*tiktokTokenResponse, error) {
	tokenURL, params := tiktokService.BuildTokenRequest(code)

	// Encode params as URL query string (TikTok uses GET, not POST)
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	fullURL := fmt.Sprintf("%s?%s", tokenURL, query.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("exchangeTiktokToken: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tiktokTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.ErrorCode != 0 {
		return nil, fmt.Errorf("exchangeTiktokToken: API error code=%d: %s", tokenResp.ErrorCode, tokenResp.ErrorMsg)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token in response")
	}

	return &tokenResp, nil
}

// tiktokShopsResponse represents the TikTok authorized shops API response.
type tiktokShopsResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		ShopID     string `json:"shop_id"`
		ShopCipher string `json:"shop_cipher"`
		ShopName   string `json:"shop_name"`
	} `json:"data"`
}

// fetchTiktokStoreIdentifier retrieves the first authorized shop's identifier from TikTok.
// Returns shopID, shopCipher, shopName, and any error.
func fetchTiktokStoreIdentifier(ctx context.Context, tiktokService *oauth.TiktokOAuthService, accessToken string) (storeIdentifier, shopCipher, storeName string, err error) {
	shopsURL, params := tiktokService.GetAuthorizedShops(accessToken)

	// Append params as query string to the signed URL
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	fullURL := fmt.Sprintf("%s?%s", shopsURL, query.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("create shops request: %w", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", "", "", fmt.Errorf("shops request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", "", fmt.Errorf("fetchTiktokStoreIdentifier: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var shopsResp tiktokShopsResponse
	if err := json.NewDecoder(resp.Body).Decode(&shopsResp); err != nil {
		return "", "", "", fmt.Errorf("decode shops response: %w", err)
	}

	if shopsResp.Code != 0 {
		return "", "", "", fmt.Errorf("fetchTiktokStoreIdentifier: API error code=%d: %s", shopsResp.Code, shopsResp.Message)
	}
	if len(shopsResp.Data) == 0 {
		return "", "", "", fmt.Errorf("no authorized shops returned")
	}

	shop := shopsResp.Data[0]
	return shop.ShopID, shop.ShopCipher, shop.ShopName, nil
}
