package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildTestState creates a signed state token for testing.
// Sets OAUTH_STATE_SECRET automatically.
func buildTestState(t *testing.T, expiresAt int64) string {
	t.Helper()
	t.Setenv("OAUTH_STATE_SECRET", "test-secret-key-for-unit-tests-32bytes!!")
	claims := oauth.StateClaims{
		TenantID:     "test_tenant",
		Platform:     "shopee",
		AttemptID:    "test-attempt-id",
		CSRFNonce:    "test-nonce-value",
		RedirectPath: "/settings/platforms",
		ExpiresAt:    expiresAt,
	}
	signedState, err := oauth.BuildSignedState(claims)
	require.NoError(t, err)
	return signedState
}

// buildTestStateForPlatform creates a signed state token for a specific platform.
func buildTestStateForPlatform(t *testing.T, platform string, expiresAt int64) string {
	t.Helper()
	t.Setenv("OAUTH_STATE_SECRET", "test-secret-key-for-unit-tests-32bytes!!")
	claims := oauth.StateClaims{
		TenantID:     "test_tenant",
		Platform:     platform,
		AttemptID:    "test-attempt-id",
		CSRFNonce:    "test-nonce-value",
		RedirectPath: "/settings/platforms",
		ExpiresAt:    expiresAt,
	}
	signedState, err := oauth.BuildSignedState(claims)
	require.NoError(t, err)
	return signedState
}

// setupCallbackTest prepares gin test context and service for callback handler tests.
func setupCallbackTest(t *testing.T, queryURL string) (*httptest.ResponseRecorder, *gin.Context, *CredentialApiService) {
	t.Helper()
	t.Setenv("OAUTH_STATE_SECRET", "test-secret-key-for-unit-tests-32bytes!!")
	t.Setenv("FRONTEND_URL", "http://localhost:5173")

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", queryURL, nil)

	svc := NewCredentialApiService("") // empty dbPath — error paths don't touch DB
	return w, c, svc
}

// setupCallbackTestWithPlatform prepares gin test context with route params set.
func setupCallbackTestWithPlatform(t *testing.T, platform, queryURL string) (*httptest.ResponseRecorder, *gin.Context, *CredentialApiService) {
	t.Helper()
	w, c, svc := setupCallbackTest(t, queryURL)
	c.Params = gin.Params{{Key: "platform", Value: platform}}
	return w, c, svc
}

func TestHandleCredentialCallback_MissingState(t *testing.T) {
	w, c, svc := setupCallbackTest(t, "/api/credentials/callback/shopee")

	svc.HandleCredentialCallback(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "error=missing_state")
}

func TestHandleCredentialCallback_InvalidState(t *testing.T) {
	w, c, svc := setupCallbackTest(t, "/api/credentials/callback/shopee?state=bogus.invalid")

	svc.HandleCredentialCallback(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "error=invalid_state")
}

func TestHandleCredentialCallback_ExpiredState(t *testing.T) {
	// Build state with past expiry
	pastExpiry := time.Now().Add(-10 * time.Minute).Unix()
	signedState := buildTestState(t, pastExpiry)

	w, c, svc := setupCallbackTest(t, "/api/credentials/callback/shopee?state="+signedState)

	svc.HandleCredentialCallback(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "error=expired_state")
}

func TestHandleCredentialCallback_AccessDenied(t *testing.T) {
	validState := buildTestState(t, time.Now().Add(10*time.Minute).Unix())
	url := "/api/credentials/callback/shopee?state=" + validState + "&error=access_denied"

	w, c, svc := setupCallbackTest(t, url)

	svc.HandleCredentialCallback(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "error=access_denied")
	// Should redirect to the path from state claims
	assert.Contains(t, location, "/settings/platforms")
}

func TestHandleCredentialCallback_UserCancelled(t *testing.T) {
	validState := buildTestState(t, time.Now().Add(10*time.Minute).Unix())
	// Valid state, no error param, empty code
	url := "/api/credentials/callback/shopee?state=" + validState + "&code="

	w, c, svc := setupCallbackTest(t, url)

	svc.HandleCredentialCallback(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "error=user_cancelled")
}

// NOTE: TestHandleCredentialCallback_MissingShopID was removed because the Phase 2
// platform validation (step 2.5) runs before the shop_id check inside case "shopee":.
// Testing missing_shop_id requires a real DB to pass steps 3-4 (attempt validation
// and app config fetch). This is now covered by integration tests.

// ---------------------------------------------------------------------------
// exchangeShopeeToken logic tests (using mock HTTP server)
//
// Since exchangeShopeeToken calls shopeeService.GetTokenURL() which returns
// the real Shopee URL, we replicate the same HTTP exchange logic here with
// a configurable mock URL to test JSON encoding/decoding and error handling.
// ---------------------------------------------------------------------------

// testTokenExchange replicates exchangeShopeeToken with a configurable URL.
func testTokenExchange(ctx context.Context, tokenURL string, reqBody map[string]interface{}) (*shopeeTokenResponse, error) {
	bodyJSON, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal token request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", tokenURL, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("create token request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("token exchange request: %w", err)
	}
	defer resp.Body.Close()

	var tokenResp shopeeTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("shopee token error: %s", tokenResp.Error)
	}
	if tokenResp.AccessToken == "" {
		return nil, fmt.Errorf("empty access_token in response")
	}

	return &tokenResp, nil
}

func TestExchangeShopeeToken_HappyPath(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "test-access-token",
			"refresh_token": "test-refresh-token",
			"expires_in":    14400,
			"shop_id":       12345,
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	reqBody := map[string]interface{}{
		"code":       "test-code",
		"shop_id":    12345,
		"partner_id": 1187586,
	}

	resp, err := testTokenExchange(ctx, mockServer.URL, reqBody)
	require.NoError(t, err)
	require.NotNil(t, resp)

	assert.Equal(t, "test-access-token", resp.AccessToken)
	assert.Equal(t, "test-refresh-token", resp.RefreshToken)
	assert.Equal(t, int64(14400), resp.ExpiresIn)
	assert.Equal(t, int64(12345), resp.ShopID)
}

func TestExchangeShopeeToken_ErrorResponse(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "invalid_code",
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	reqBody := map[string]interface{}{
		"code":       "bad-code",
		"shop_id":    12345,
		"partner_id": 1187586,
	}

	resp, err := testTokenExchange(ctx, mockServer.URL, reqBody)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "shopee token error")
	assert.Contains(t, err.Error(), "invalid_code")
}

func TestExchangeShopeeToken_EmptyAccessToken(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"access_token":  "",
			"refresh_token": "test-refresh",
			"expires_in":    14400,
		})
	}))
	defer mockServer.Close()

	ctx := context.Background()
	reqBody := map[string]interface{}{"code": "test", "shop_id": 1}

	resp, err := testTokenExchange(ctx, mockServer.URL, reqBody)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "empty access_token")
}

// ---------------------------------------------------------------------------
// Platform dispatch tests
// ---------------------------------------------------------------------------

func TestHandleCredentialCallback_PlatformMismatch(t *testing.T) {
	validState := buildTestStateForPlatform(t, "shopee", time.Now().Add(10*time.Minute).Unix())
	url := "/api/credentials/callback/lazada?state=" + validState + "&code=test-code"

	w, c, svc := setupCallbackTestWithPlatform(t, "lazada", url)

	svc.HandleCredentialCallback(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "error=platform_mismatch")
}

func TestHandleCredentialCallback_PlatformMismatch_TiktokVsShopee(t *testing.T) {
	validState := buildTestStateForPlatform(t, "tiktok", time.Now().Add(10*time.Minute).Unix())
	url := "/api/credentials/callback/shopee?state=" + validState + "&code=test-code"

	w, c, svc := setupCallbackTestWithPlatform(t, "shopee", url)

	svc.HandleCredentialCallback(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "error=platform_mismatch")
}

func TestHandleCredentialCallback_PlatformMismatch_LazadaVsTiktok(t *testing.T) {
	validState := buildTestStateForPlatform(t, "lazada", time.Now().Add(10*time.Minute).Unix())
	url := "/api/credentials/callback/tiktok?state=" + validState + "&code=test-code"

	w, c, svc := setupCallbackTestWithPlatform(t, "tiktok", url)

	svc.HandleCredentialCallback(c)

	assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
	location := w.Header().Get("Location")
	assert.Contains(t, location, "error=platform_mismatch")
}

func TestHandleCredentialCallback_AllPlatformsRoutable(t *testing.T) {
	// Verify all platforms pass the mismatch check (step 2.5) and reach the
	// tenant DB step. With an empty dbPath (no running PostgreSQL in unit tests),
	// tenantDB() fails and the handler returns tenant_db_error.
	platforms := []string{"shopee", "lazada", "tiktok"}
	for _, platform := range platforms {
		t.Run(platform, func(t *testing.T) {
			validState := buildTestStateForPlatform(t, platform, time.Now().Add(10*time.Minute).Unix())
			url := "/api/credentials/callback/" + platform + "?state=" + validState + "&code=test-code"

			gin.SetMode(gin.TestMode)
			t.Setenv("FRONTEND_URL", "http://localhost:5173")
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", url, nil)
			c.Params = gin.Params{{Key: "platform", Value: platform}}

			svc := NewCredentialApiService("")
			svc.HandleCredentialCallback(c)

			assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
			location := w.Header().Get("Location")
			assert.NotContains(t, location, "error=platform_mismatch",
				"platform %s should pass mismatch check", platform)
			assert.Contains(t, location, "error=tenant_db_error",
				"platform %s should reach tenant DB validation", platform)
		})
	}
}
