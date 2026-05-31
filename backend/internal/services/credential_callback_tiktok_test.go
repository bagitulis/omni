package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// testTiktokTokenExchange replicates exchangeTiktokToken HTTP logic with a
// configurable URL (since BuildTokenRequest returns the real TikTok URL).
// TikTok uses GET with query parameters for token exchange.
// ---------------------------------------------------------------------------

func testTiktokTokenExchange(ctx context.Context, tokenURL string) (*tiktokTokenResponse, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp tiktokTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token in response")
	}

	return &tokenResp, nil
}

// testFetchTiktokStoreIdentifier replicates fetchTiktokStoreIdentifier HTTP logic
// with a configurable URL.
func testFetchTiktokStoreIdentifier(ctx context.Context, shopsURL string) (string, string, string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", shopsURL, nil)
	if err != nil {
		return "", "", "", fmt.Errorf("create shops request: %w", err)
	}

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return "", "", "", fmt.Errorf("shops request: %w", err)
	}
	defer resp.Body.Close()

	var shopsResp tiktokShopsResponse
	if err := json.NewDecoder(resp.Body).Decode(&shopsResp); err != nil {
		return "", "", "", fmt.Errorf("decode shops response: %w", err)
	}

	if len(shopsResp.Data) == 0 {
		return "", "", "", fmt.Errorf("no authorized shops returned")
	}

	shop := shopsResp.Data[0]
	return shop.ShopID, shop.ShopCipher, shop.ShopName, nil
}

func TestExchangeTiktokToken_HappyPath(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":       "tiktok-access-token-789",
			"refresh_token":      "tiktok-refresh-token-012",
			"expires_in":         86400,
			"refresh_expires_in": 31536000,
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	tokenURL := fmt.Sprintf("%s?auth_code=test-code&app_key=test-key&timestamp=1234567890&sign=test-sign", mockServer.URL)

	resp, err := testTiktokTokenExchange(ctx, tokenURL)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "tiktok-access-token-789", resp.AccessToken)
	assert.Equal(t, "tiktok-refresh-token-012", resp.RefreshToken)
	assert.Equal(t, int64(86400), resp.ExpiresIn)
	assert.Equal(t, int64(31536000), resp.RefreshExpiresIn)
}

func TestExchangeTiktokToken_ErrorResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token": "",
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	resp, err := testTiktokTokenExchange(ctx, mockServer.URL)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "empty access_token")
}

func TestExchangeTiktokToken_EmptyAccessToken(t *testing.T) {
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
	resp, err := testTiktokTokenExchange(ctx, mockServer.URL)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "empty access_token")
}

func TestExchangeTiktokToken_VerifyGET(t *testing.T) {
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
	resp, err := testTiktokTokenExchange(ctx, mockServer.URL)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "GET", capturedMethod, "TikTok token exchange must use GET")
}

func TestFetchTiktokStoreIdentifier_HappyPath(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data": []map[string]interface{}{
				{
					"shop_id":     "789012",
					"shop_cipher": "aB3cD4eF5gH6iJ7kL8mN",
					"shop_name":   "My TikTok Shop",
				},
			},
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	storeIdentifier, shopCipher, storeName, err := testFetchTiktokStoreIdentifier(ctx, mockServer.URL)
	require.NoError(t, err)

	assert.Equal(t, "789012", storeIdentifier)
	assert.Equal(t, "aB3cD4eF5gH6iJ7kL8mN", shopCipher)
	assert.Equal(t, "My TikTok Shop", storeName)
}

func TestFetchTiktokStoreIdentifier_EmptyShops(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    0,
			"message": "success",
			"data":    []interface{}{},
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	storeIdentifier, shopCipher, storeName, err := testFetchTiktokStoreIdentifier(ctx, mockServer.URL)
	require.Error(t, err)

	assert.Empty(t, storeIdentifier)
	assert.Empty(t, shopCipher)
	assert.Empty(t, storeName)
	assert.Contains(t, err.Error(), "no authorized shops returned")
}
