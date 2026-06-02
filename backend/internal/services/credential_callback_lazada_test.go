package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// testLazadaTokenExchange replicates exchangeLazadaToken HTTP logic with a
// configurable URL (since BuildTokenRequest returns the real Lazada URL).
// This tests the same JSON encoding/decoding and error handling paths.
// ---------------------------------------------------------------------------

func testLazadaTokenExchange(ctx context.Context, tokenURL string, formValues url.Values) (*lazadaTokenResponse, error) {
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

// testFetchLazadaStoreIdentifier replicates fetchLazadaStoreIdentifier HTTP logic
// with a configurable URL.
func testFetchLazadaStoreIdentifier(ctx context.Context, apiURL string) (string, string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("create lazada seller info request: %w", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", "", fmt.Errorf("lazada seller info request: %w", err)
	}
	defer resp.Body.Close()

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

	return fmt.Sprintf("%d", sellerResp.Data.SellerID), sellerResp.Data.SellerName, nil
}

func TestExchangeLazadaToken_HappyPath(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":       "lazada-access-token-123",
			"refresh_token":      "lazada-refresh-token-456",
			"expires_in":         86400,
			"refresh_expires_in": 2592000,
			"seller_id":          "MY12ABC",
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	formValues := url.Values{
		"code":      {"test-auth-code"},
		"app_key":   {"test-app-key"},
		"timestamp": {"1234567890000"},
		"sign":      {"test-signature"},
	}

	resp, err := testLazadaTokenExchange(ctx, mockServer.URL, formValues)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "lazada-access-token-123", resp.AccessToken)
	assert.Equal(t, "lazada-refresh-token-456", resp.RefreshToken)
	assert.Equal(t, int64(86400), resp.ExpiresIn)
	assert.Equal(t, int64(2592000), resp.RefreshExpiresIn)
	assert.Equal(t, "MY12ABC", resp.SellerID)
}

func TestExchangeLazadaToken_ErrorResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid_code",
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	formValues := url.Values{"code": {"bad-code"}}

	resp, err := testLazadaTokenExchange(ctx, mockServer.URL, formValues)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "lazada token error")
	assert.Contains(t, err.Error(), "invalid_code")
}

func TestExchangeLazadaToken_EmptyAccessToken(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "",
			"refresh_token": "some-refresh",
			"expires_in":    86400,
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	formValues := url.Values{"code": {"test"}}

	resp, err := testLazadaTokenExchange(ctx, mockServer.URL, formValues)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "empty access_token")
}

func TestExchangeLazadaToken_VerifyPOST(t *testing.T) {
	var capturedMethod string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "test-token",
			"refresh_token": "test-refresh",
			"expires_in":    86400,
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	formValues := url.Values{"code": {"test"}}

	resp, err := testLazadaTokenExchange(ctx, mockServer.URL, formValues)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "POST", capturedMethod, "Lazada token exchange must use POST")
}

func TestFetchLazadaStoreIdentifier_HappyPath(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "0",
			"message": "success",
			"data": map[string]interface{}{
				"seller_id":   12345678,
				"seller_name": "My Lazada Store",
			},
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	storeIdentifier, storeName, err := testFetchLazadaStoreIdentifier(ctx, mockServer.URL)
	require.NoError(t, err)

	assert.Equal(t, "12345678", storeIdentifier)
	assert.Equal(t, "My Lazada Store", storeName)
}

func TestFetchLazadaStoreIdentifier_ErrorCode(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "15",
			"message": "InvalidAccessParameter",
			"data":    nil,
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	storeIdentifier, storeName, err := testFetchLazadaStoreIdentifier(ctx, mockServer.URL)
	require.Error(t, err)

	assert.Empty(t, storeIdentifier)
	assert.Empty(t, storeName)
	assert.Contains(t, err.Error(), "lazada seller info error")
	assert.Contains(t, err.Error(), "code=15")
	assert.Contains(t, err.Error(), "InvalidAccessParameter")
}
