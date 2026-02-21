package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/services/sync"
)

func maskToken(token string) string {
	if len(token) <= 8 {
		return "tok_****"
	}
	return "tok_****" + token[len(token)-4:]
}

// executeShopeeTokenRefresh executes Shopee token refresh
func (m *TokenManager) executeShopeeTokenRefresh(ctx context.Context, tenantID, url string, body map[string]interface{}) (*TokenInfo, error) {
	bodyJSON, _ := json.Marshal(body)
	resp, err := m.httpClient.Post(url, "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if errCode, ok := result["error"].(string); ok && errCode != "" {
		errMsg, _ := result["message"].(string)
		return nil, fmt.Errorf("shopee token refresh failed: %s - %s", errCode, errMsg)
	}

	accessToken, _ := result["access_token"].(string)
	newRefreshToken, _ := result["refresh_token"].(string)
	expiresIn := int64(14400) // Default 4 hours
	if val, ok := result["expire_in"].(float64); ok && val > 0 {
		expiresIn = int64(val)
	} else {
		log.Warn().Msg("[SHOPEE REFRESH] expire_in missing or invalid, using default: 4 hours")
	}
	// Shopee refresh token valid for 7 days per Shopee API documentation
	refreshExpiresIn := int64(7 * 24 * 60 * 60) // 7 days in seconds

	log.Info().
		Int64("expires_in", expiresIn).
		Int64("expires_in_hours", expiresIn/3600).
		Int64("refresh_expires_in_days", refreshExpiresIn/86400).
		Msg("[SHOPEE REFRESH] Success! Tokens refreshed")

	return m.saveNewTokens(ctx, tenantID, models.PlatformShopee, accessToken, newRefreshToken, expiresIn, refreshExpiresIn)
}

// executeLazadaTokenRefresh executes Lazada token refresh
func (m *TokenManager) executeLazadaTokenRefresh(ctx context.Context, tenantID, url string, params map[string]string) (*TokenInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	log.Info().Str("url", req.URL.String()).Msg("[LAZADA REFRESH] Requesting refresh")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("[LAZADA REFRESH] HTTP error")
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Info().Int("size", len(respBody)).Msg("[LAZADA REFRESH] Response received")

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Error().Err(err).Msg("[LAZADA REFRESH] JSON parse error")
		return nil, err
	}

	var errCode string
	if codeStr, ok := result["code"].(string); ok {
		errCode = codeStr
	} else if codeNum, ok := result["code"].(float64); ok {
		errCode = fmt.Sprintf("%.0f", codeNum)
	}

	if errCode != "" && errCode != "0" {
		log.Error().Str("code", errCode).Interface("message", result["message"]).Msg("[LAZADA REFRESH] API error")
		return nil, fmt.Errorf("lazada token refresh failed: %s - %v", errCode, result["message"])
	}

	accessToken, _ := result["access_token"].(string)
	newRefreshToken, _ := result["refresh_token"].(string)

	expiresIn := int64(7 * 24 * 60 * 60) // Default 7 days
	if val, ok := result["expires_in"].(float64); ok && val > 0 {
		expiresIn = int64(val)
	} else {
		log.Warn().Msg("[LAZADA REFRESH] expires_in missing or invalid, using default: 7 days")
	}
	// Lazada refresh token - use actual value from API if provided
	refreshExpiresIn := int64(30 * 24 * 60 * 60) // Default 30 days
	if val, ok := result["refresh_expires_in"].(float64); ok && val > 0 {
		refreshExpiresIn = int64(val)
		log.Info().Int64("days", refreshExpiresIn/86400).Msg("[LAZADA REFRESH] refresh_expires_in from API")
	} else {
		log.Info().Msg("[LAZADA REFRESH] refresh_expires_in not provided, using default: 30 days")
	}

	log.Info().
		Int64("expires_in", expiresIn).
		Int64("refresh_expires_in_days", refreshExpiresIn/86400).
		Msg("[LAZADA REFRESH] Success! Tokens refreshed")
	return m.saveNewTokens(ctx, tenantID, models.PlatformLazada, accessToken, newRefreshToken, expiresIn, refreshExpiresIn)
}

// executeTiktokTokenRefresh executes TikTok token refresh
func (m *TokenManager) executeTiktokTokenRefresh(ctx context.Context, tenantID, url string, params map[string]string) (*TokenInfo, error) {
	req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	log.Info().Str("url", req.URL.String()).Msg("[TIKTOK REFRESH] Requesting refresh")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("[TIKTOK REFRESH] HTTP error")
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	log.Info().Int("size", len(respBody)).Msg("[TIKTOK REFRESH] Response received")

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Error().Err(err).Msg("[TIKTOK REFRESH] JSON parse error")
		return nil, err
	}

	// Check for error in response
	if code, ok := result["code"].(float64); ok && code != 0 {
		log.Error().Interface("code", result["code"]).Interface("message", result["message"]).Msg("[TIKTOK REFRESH] API error")
		return nil, fmt.Errorf("tiktok token refresh failed: %v", result["message"])
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok || data == nil {
		log.Error().Msg("[TIKTOK REFRESH] Invalid response structure: data is missing or nil")
		return nil, fmt.Errorf("tiktok token refresh failed: invalid response structure")
	}

	accessToken, _ := data["access_token"].(string)
	if accessToken == "" {
		log.Error().Msg("[TIKTOK REFRESH] No access_token in response data")
		return nil, fmt.Errorf("tiktok token refresh failed: no access_token in response")
	}

	newRefreshToken, _ := data["refresh_token"].(string)

	// TikTok API returns access_token_expire_in in two different formats:
	// 1. Unix timestamp (seconds since epoch) - if value > current time in seconds
	// 2. Relative seconds - if value <= current time in seconds
	// We need to handle both cases like Node.js backend does
	accessTokenExpireValue := int64(0)
	if val, ok := data["access_token_expire_in"].(float64); ok {
		accessTokenExpireValue = int64(val)
	}

	nowSeconds := time.Now().Unix()
	var expiresIn int64

	if accessTokenExpireValue > nowSeconds {
		// It's a Unix timestamp (seconds) - convert to relative seconds
		expiresIn = accessTokenExpireValue - nowSeconds
		log.Info().Int64("expires_in", expiresIn).Msg("[TIKTOK REFRESH] access_token_expire_in is Unix timestamp")
	} else if accessTokenExpireValue > 0 {
		// It's relative seconds - use as-is
		expiresIn = accessTokenExpireValue
		log.Info().Int64("expires_in", expiresIn).Msg("[TIKTOK REFRESH] access_token_expire_in is relative")
	} else {
		// Default to 7 days if not provided
		expiresIn = 7 * 24 * 60 * 60
		log.Info().Int64("expires_in", expiresIn).Msg("[TIKTOK REFRESH] access_token_expire_in not provided, using default")
	}

	// Handle refresh token expiry - normalize like access token
	refreshTokenExpireValue := int64(0)
	if val, ok := data["refresh_token_expire_in"].(float64); ok {
		refreshTokenExpireValue = int64(val)
	}

	var refreshExpiresIn int64
	if refreshTokenExpireValue > nowSeconds {
		// It's a Unix timestamp - convert to relative seconds
		refreshExpiresIn = refreshTokenExpireValue - nowSeconds
		log.Info().Int64("refresh_expires_in", refreshExpiresIn).Int64("days", refreshExpiresIn/86400).Msg("[TIKTOK REFRESH] refresh_token_expire_in is Unix timestamp")
	} else if refreshTokenExpireValue > 0 {
		// It's relative seconds - use as-is
		refreshExpiresIn = refreshTokenExpireValue
		log.Info().Int64("refresh_expires_in", refreshExpiresIn).Int64("days", refreshExpiresIn/86400).Msg("[TIKTOK REFRESH] refresh_token_expire_in is relative")
	} else {
		// Default to 90 days if not provided
		refreshExpiresIn = 90 * 24 * 60 * 60
		log.Info().Int64("refresh_expires_in", refreshExpiresIn).Msg("[TIKTOK REFRESH] refresh_token_expire_in not provided, using default")
	}

	// Cap at 365 days max — TikTok sometimes returns bogus far-future Unix timestamps
	// (e.g. 4875922303 = year 2124, a sentinel "never expires" value that must be capped)
	const maxTiktokRefreshSeconds = int64(365 * 24 * 60 * 60)
	if refreshExpiresIn > maxTiktokRefreshSeconds {
		log.Warn().
			Int64("raw_value", refreshTokenExpireValue).
			Int64("computed_seconds", refreshExpiresIn).
			Int64("computed_days", refreshExpiresIn/86400).
			Int64("capped_days", maxTiktokRefreshSeconds/86400).
			Msg("[TIKTOK REFRESH] refresh_token_expire_in exceeds 365 days, capping (TikTok sentinel value)")
		refreshExpiresIn = maxTiktokRefreshSeconds
	}

	log.Info().
		Int64("expires_in", expiresIn).
		Int64("refresh_expires_in", refreshExpiresIn).
		Msg("[TIKTOK REFRESH] Success! Tokens refreshed")
	return m.saveNewTokens(ctx, tenantID, models.PlatformTiktok, accessToken, newRefreshToken, expiresIn, refreshExpiresIn)
}

// saveNewTokens saves refreshed tokens to tenant database using key-value format
// This matches Node.js behavior: saves to PlatformConfig table with configKey/configValue
// IMPORTANT: Also invalidates cached platform clients so they will reload with new tokens
func (m *TokenManager) saveNewTokens(ctx context.Context, tenantID, platformName, accessToken, refreshToken string, expiresIn, refreshExpiresIn int64) (*TokenInfo, error) {
	tenantRepo, err := m.getTenantConfigRepo(tenantID)
	if err != nil {
		return nil, err
	}

	// Save tokens using the key-value format (matches Node.js)
	err = tenantRepo.UpdateTokens(ctx, platformName, accessToken, refreshToken, expiresIn, refreshExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("failed to save tokens: %w", err)
	}

	// CRITICAL: Invalidate cached platform clients so they reload with new tokens
	// Without this, the old token would still be used until server restart
	platform.InvalidateTenantPlatformService(tenantID)
	sync.InvalidateTenantInstance(tenantID)

	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Second)
	refreshExpiresAt := time.Now().Add(time.Duration(refreshExpiresIn) * time.Second)

	return &TokenInfo{
		Platform:            platformName,
		AccessToken:         accessToken,
		RefreshToken:        refreshToken,
		ExpiresAt:           expiresAt,
		RefreshTokenExpires: refreshExpiresAt,
		IsValid:             true,
		NeedsRefresh:        false,
	}, nil
}

// GetAllTokenStatus returns token status for all platforms for a tenant
func (m *TokenManager) GetAllTokenStatus(ctx context.Context, tenantID string) (map[string]*TokenInfo, error) {
	platforms := []string{models.PlatformShopee, models.PlatformLazada, models.PlatformTiktok}
	result := make(map[string]*TokenInfo)

	for _, p := range platforms {
		info, err := m.GetTokenStatus(ctx, tenantID, p)
		if err != nil {
			// Log but continue with other platforms
			result[p] = &TokenInfo{Platform: p, IsValid: false}
			continue
		}
		result[p] = info
	}

	return result, nil
}

// RefreshExpiredTokens refreshes all expired tokens for a tenant
func (m *TokenManager) RefreshExpiredTokens(ctx context.Context, tenantID string) (map[string]bool, error) {
	results := make(map[string]bool)

	// Check and refresh Shopee
	if status, _ := m.GetTokenStatus(ctx, tenantID, models.PlatformShopee); status != nil && status.NeedsRefresh {
		_, err := m.RefreshShopeeToken(ctx, tenantID)
		results[models.PlatformShopee] = err == nil
	}

	// Check and refresh Lazada
	if status, _ := m.GetTokenStatus(ctx, tenantID, models.PlatformLazada); status != nil && status.NeedsRefresh {
		_, err := m.RefreshLazadaToken(ctx, tenantID)
		results[models.PlatformLazada] = err == nil
	}

	// Check and refresh TikTok
	if status, _ := m.GetTokenStatus(ctx, tenantID, models.PlatformTiktok); status != nil && status.NeedsRefresh {
		_, err := m.RefreshTiktokToken(ctx, tenantID)
		results[models.PlatformTiktok] = err == nil
	}

	return results, nil
}

// GetShopID returns the shop ID for a platform from tenant config
func (m *TokenManager) GetShopID(ctx context.Context, tenantID, platformName string) (int64, error) {
	tenantRepo, err := m.getTenantConfigRepo(tenantID)
	if err != nil {
		return 0, err
	}

	tokenInfo, err := tenantRepo.GetTokenInfo(ctx, platformName)
	if err != nil {
		return 0, err
	}
	if tokenInfo == nil {
		return 0, fmt.Errorf("no config found for platform %s", platformName)
	}

	return tokenInfo.ShopID, nil
}

// GetAccessToken returns decrypted access token for a platform
func (m *TokenManager) GetAccessToken(ctx context.Context, tenantID, platformName string) (string, error) {
	status, err := m.GetTokenStatus(ctx, tenantID, platformName)
	if err != nil {
		return "", err
	}
	if !status.IsValid {
		return "", fmt.Errorf("token is invalid or expired for platform %s", platformName)
	}
	return status.AccessToken, nil
}
