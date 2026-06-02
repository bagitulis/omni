package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/omni/backend/internal/services/oauth"
	"github.com/rs/zerolog/log"
)

// lazadaTokenResponse represents the Lazada token exchange response.
type lazadaTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	SellerID         string `json:"seller_id"`
	ErrorMsg         string `json:"error,omitempty"`
}

// lazadaSellerInfoResponse represents the Lazada seller info API response.
type lazadaSellerInfoResponse struct {
	Code    string                `json:"code"`
	Message string                `json:"message"`
	Data    *lazadaSellerInfoData `json:"data"`
}

// lazadaSellerInfoData holds seller info from Lazada /seller/info endpoint.
type lazadaSellerInfoData struct {
	SellerID   int64  `json:"seller_id"`
	SellerName string `json:"seller_name"`
}

// exchangeLazadaToken performs an HTTP POST to exchange a Lazada auth code for tokens.
// Lazada uses form-encoded POST body (not JSON).
func exchangeLazadaToken(ctx context.Context, lazadaService *oauth.LazadaOAuthService, code string) (*lazadaTokenResponse, error) {
	tokenURL, params := lazadaService.BuildTokenRequest(code)

	// Encode params as form body
	formValues := url.Values{}
	for k, v := range params {
		formValues.Set(k, v)
	}
	formBody := formValues.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(formBody))
	if err != nil {
		return nil, fmt.Errorf("create lazada token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("lazada token exchange request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("exchangeLazadaToken: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp lazadaTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode lazada token response: %w", err)
	}

	if tokenResp.ErrorMsg != "" {
		return nil, fmt.Errorf("lazada token error: %s", tokenResp.ErrorMsg)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token in lazada token response")
	}

	return &tokenResp, nil
}

// fetchLazadaStoreIdentifier calls the Lazada /seller/info API to retrieve the
// seller ID and store name for a connected Lazada account.
func fetchLazadaStoreIdentifier(ctx context.Context, lazadaService *oauth.LazadaOAuthService, accessToken string) (storeIdentifier, storeName string, err error) {
	apiURL, params := lazadaService.BuildAPIRequest("/seller/info", accessToken, nil)

	// Append params as query string
	query := url.Values{}
	for k, v := range params {
		query.Set(k, v)
	}
	signedURL := fmt.Sprintf("%s?%s", apiURL, query.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", signedURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("create lazada seller info request: %w", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("lazada seller info request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("fetchLazadaStoreIdentifier: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var sellerResp lazadaSellerInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&sellerResp); err != nil {
		return "", "", fmt.Errorf("decode lazada seller info response: %w", err)
	}

	if sellerResp.Code != "0" {
		return "", "", fmt.Errorf("lazada seller info error: code=%s message=%s", sellerResp.Code, sellerResp.Message)
	}
	if sellerResp.Data == nil {
		return "", "", fmt.Errorf("lazada seller info returned empty data")
	}

	storeIdentifier = fmt.Sprintf("%d", sellerResp.Data.SellerID)
	storeName = sellerResp.Data.SellerName

	log.Info().
		Str("store_identifier", storeIdentifier).
		Str("store_name", storeName).
		Msg("Fetched Lazada store identifier")

	return storeIdentifier, storeName, nil
}
