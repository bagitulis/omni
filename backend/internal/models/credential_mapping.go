package models

import "strconv"

// PlatformConfigsRow represents a single row from the platform_configs table.
// Each platform has multiple rows keyed by config_key with string config_value.
type PlatformConfigsRow struct {
	ID          string `json:"id"`
	Platform    string `json:"platform"`
	ConfigKey   string `json:"config_key"`
	ConfigValue string `json:"config_value"`
	DataType    string `json:"data_type"`
	IsEncrypted bool   `json:"is_encrypted"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

// MapToAppConfig converts platform_configs rows into a CredentialAppConfig.
// Platform-specific required keys:
//   - Shopee:  partnerId, partnerKey
//   - Lazada:  appKey, appSecret
//   - TikTok:  appKey, appSecret
//
// Returns nil if required keys are missing or the platform is unsupported.
// Encrypted values are passed through as-is without decryption.
func MapToAppConfig(rows []PlatformConfigsRow, tenantID string) *CredentialAppConfig {
	if len(rows) == 0 {
		return nil
	}

	kv := buildConfigMap(rows)
	platform := rows[0].Platform

	cfg := &CredentialAppConfig{
		TenantID: tenantID,
		Platform: platform,
	}

	switch platform {
	case "shopee":
		pidStr, ok := kv["partnerId"]
		if !ok {
			return nil
		}
		pid, err := strconv.ParseInt(pidStr, 10, 64)
		if err != nil {
			return nil
		}
		cfg.PartnerID = pid
		cfg.PartnerKey = kv["partnerKey"]
		if cfg.PartnerKey == "" {
			return nil
		}

	case "lazada", "tiktok":
		cfg.AppKey = kv["appKey"]
		cfg.AppSecret = kv["appSecret"]
		if cfg.AppKey == "" || cfg.AppSecret == "" {
			return nil
		}

	default:
		return nil
	}

	return cfg
}

// MapToConnection converts platform_configs rows into a CredentialConnection.
// Extracts: shopId, accessToken, refreshToken, tokenExpiry, refreshTokenExpiry,
// and shopCipher (TikTok only). Encrypted values are passed through as-is.
// Numeric timestamps that fail to parse are silently skipped (left at zero value).
func MapToConnection(rows []PlatformConfigsRow, tenantID string, storeName string) *CredentialConnection {
	if len(rows) == 0 {
		return nil
	}

	kv := buildConfigMap(rows)
	platform := rows[0].Platform

	conn := &CredentialConnection{
		TenantID:  tenantID,
		Platform:  platform,
		StoreName: storeName,
		Status:    "disconnected",
	}

	if shopID, ok := kv["shopId"]; ok {
		conn.StoreIdentifier = shopID
	}

	// Token values — preserved as-is, never decrypted.
	conn.AccessToken = kv["accessToken"]
	conn.RefreshToken = kv["refreshToken"]

	// Numeric timestamps (milliseconds since epoch) — optional, best-effort parse.
	if v, ok := kv["tokenExpiry"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			conn.TokenExpiry = n
		}
	}
	if v, ok := kv["refreshTokenExpiry"]; ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			conn.RefreshExpiry = n
		}
	}

	// TikTok-specific: shop cipher for API calls.
	if v, ok := kv["shopCipher"]; ok && platform == "tiktok" {
		conn.ShopCipher = v
	}

	return conn
}

// buildConfigMap converts a slice of PlatformConfigsRow into a flat
// config_key → config_value map for convenient lookup.
func buildConfigMap(rows []PlatformConfigsRow) map[string]string {
	m := make(map[string]string, len(rows))
	for _, r := range rows {
		m[r.ConfigKey] = r.ConfigValue
	}
	return m
}
