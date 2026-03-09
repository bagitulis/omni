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
)

// executeShopeeTokenRefresh executes Shopee token refresh
func (m *TokenManager) executeShopeeTokenRefresh(ctx context.Context, tenantID, url string, body map[string]interface{}) (*TokenInfo, error) {
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}
	resp, err := m.httpClient.Post(url, "application/json", bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	accessToken, _ := result["access_token"].(string)
	if accessToken == "" {
		return nil, fmt.Errorf("shopee token refresh failed: no access_token in response")
	}
	newRefreshToken, _ := result["refresh_token"].(string)

	expiresIn := int64(14400)
	if val, ok := result["expire_in"].(float64); ok && val > 0 {
		expiresIn = int64(val)
	}
	// Shopee refresh token valid for 30 days
	refreshExpiresIn := int64(30 * 24 * 60 * 60)

	log.Info().
		Int64("expires_in", expiresIn).
		Int64("refresh_expires_in", refreshExpiresIn).
		Msg("[SHOPEE REFRESH] Success! Tokens refreshed")
	return m.saveNewTokens(ctx, tenantID, models.PlatformShopee, accessToken, newRefreshToken, expiresIn, refreshExpiresIn)
}

// executeLazadaTokenRefresh executes Lazada token refresh
func (m *TokenManager) executeLazadaTokenRefresh(ctx context.Context, tenantID, url string, params map[string]string) (*TokenInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	log.Info().Str("host", req.URL.Host).Str("path", req.URL.Path).Msg("[LAZADA REFRESH] Requesting refresh")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("[LAZADA REFRESH] HTTP error")
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
	log.Info().Int("size", len(respBody)).Msg("[LAZADA REFRESH] Response received")

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		log.Error().Err(err).Msg("[LAZADA REFRESH] JSON parse error")
		return nil, err
	}

	accessToken, _ := result["access_token"].(string)
	if accessToken == "" {
		return nil, fmt.Errorf("lazada token refresh failed: no access_token in response")
	}

	newRefreshToken, _ := result["refresh_token"].(string)
	expiresIn := int64(604800)
	if val, ok := result["expires_in"].(float64); ok && val > 0 {
		expiresIn = int64(val)
	}
	refreshExpiresIn := int64(2592000)
	if val, ok := result["refresh_expires_in"].(float64); ok && val > 0 {
		refreshExpiresIn = int64(val)
	}

	log.Info().
		Int64("expires_in", expiresIn).
		Int64("refresh_expires_in", refreshExpiresIn).
		Msg("[LAZADA REFRESH] Success! Tokens refreshed")
	return m.saveNewTokens(ctx, tenantID, models.PlatformLazada, accessToken, newRefreshToken, expiresIn, refreshExpiresIn)
}

// executeTiktokTokenRefresh executes TikTok token refresh
func (m *TokenManager) executeTiktokTokenRefresh(ctx context.Context, tenantID, url string, params map[string]string) (*TokenInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	q := req.URL.Query()
	for k, v := range params {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	log.Info().Str("host", req.URL.Host).Str("path", req.URL.Path).Msg("[TIKTOK REFRESH] Requesting refresh")

	resp, err := m.httpClient.Do(req)
	if err != nil {
		log.Error().Err(err).Msg("[TIKTOK REFRESH] HTTP error")
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}
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

	expiresIn := normalizeTiktokExpiry(data, "access_token_expire_in", 7*24*60*60, "access")
	refreshExpiresIn := normalizeTiktokExpiry(data, "refresh_token_expire_in", 90*24*60*60, "refresh")

	// Cap at 365 days max — TikTok sometimes returns bogus far-future Unix timestamps
	const maxTiktokRefreshSeconds = int64(365 * 24 * 60 * 60)
	if refreshExpiresIn > maxTiktokRefreshSeconds {
		log.Warn().
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

// normalizeTiktokExpiry handles TikTok's mixed expiry format (Unix timestamp vs relative seconds)
func normalizeTiktokExpiry(data map[string]interface{}, key string, defaultVal int64, label string) int64 {
	rawValue := int64(0)
	if val, ok := data[key].(float64); ok {
		rawValue = int64(val)
	}

	nowSeconds := time.Now().Unix()

	if rawValue > nowSeconds {
		// It's a Unix timestamp — convert to relative seconds
		relative := rawValue - nowSeconds
		log.Info().Int64("expires_in", relative).Msgf("[TIKTOK REFRESH] %s_token_expire_in is Unix timestamp", label)
		return relative
	} else if rawValue > 0 {
		// It's relative seconds — use as-is
		log.Info().Int64("expires_in", rawValue).Msgf("[TIKTOK REFRESH] %s_token_expire_in is relative", label)
		return rawValue
	}

	// Default
	log.Info().Int64("expires_in", defaultVal).Msgf("[TIKTOK REFRESH] %s_token_expire_in not provided, using default", label)
	return defaultVal
}
