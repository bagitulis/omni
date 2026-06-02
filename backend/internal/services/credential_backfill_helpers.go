package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// configEntry tracks a single config_key row with its timestamp for duplicate resolution.
type configEntry struct {
	value     string
	updatedAt time.Time
}

// buildConfigMap creates a key-value map from platform config rows, resolving
// duplicate keys using a precedence policy:
//   - Duplicate keys with one empty and one non-empty → use non-empty
//   - Duplicate keys with conflicting non-empty values → log warning, use latest updated_at
func buildConfigMap(rows []inventoryRowData) map[string]string {
	keyEntries := make(map[string][]configEntry)

	for _, row := range rows {
		key := strings.TrimSpace(asString(row["config_key"]))
		value := asString(row["config_value"])
		if key == "" {
			continue
		}
		var ts time.Time
		if v := row["updated_at"]; v != nil {
			switch t := v.(type) {
			case time.Time:
				ts = t
			case string:
				ts, _ = time.Parse(time.RFC3339, t)
			}
		}
		keyEntries[key] = append(keyEntries[key], configEntry{value: value, updatedAt: ts})
	}

	m := make(map[string]string, len(keyEntries))
	for key, entries := range keyEntries {
		if len(entries) == 1 {
			m[key] = entries[0].value
			continue
		}
		m[key] = resolveDuplicateKey(key, entries)
	}
	return m
}

// resolveDuplicateKey applies the duplicate key precedence policy.
// Returns the winning value for a config_key with multiple rows.
func resolveDuplicateKey(key string, entries []configEntry) string {
	// Find indices of all non-empty values
	nonEmpty := make([]int, 0, len(entries))
	for i, e := range entries {
		if e.value != "" {
			nonEmpty = append(nonEmpty, i)
		}
	}

	// All empty → return empty
	if len(nonEmpty) == 0 {
		return ""
	}

	// Exactly one non-empty → use it (regardless of timestamp)
	if len(nonEmpty) == 1 {
		return entries[nonEmpty[0]].value
	}

	// Multiple non-empty values → find latest updated_at among non-empty
	bestIdx := nonEmpty[0]
	for _, idx := range nonEmpty[1:] {
		if entries[idx].updatedAt.After(entries[bestIdx].updatedAt) {
			bestIdx = idx
		}
	}

	// Log warning about conflicting values
	log.Warn().
		Str("config_key", key).
		Int("duplicate_count", len(entries)).
		Time("selected_updated_at", entries[bestIdx].updatedAt).
		Msg("Duplicate config_key resolved — latest updated_at wins")

	return entries[bestIdx].value
}

// ClassifyRows splits inventory rows into credential and non-credential categories.
func ClassifyRows(rows []inventoryRowData) (credRows, nonCredRows []inventoryRowData) {
	for _, row := range rows {
		key := asString(row["config_key"])
		if credentialKeys[key] {
			credRows = append(credRows, row)
		} else {
			nonCredRows = append(nonCredRows, row)
		}
	}
	return credRows, nonCredRows
}

// groupByPlatform groups classified rows by their platform field.
func groupByPlatform(rows []inventoryRowData) map[string][]inventoryRowData {
	result := make(map[string][]inventoryRowData)
	for _, row := range rows {
		platform := strings.TrimSpace(asString(row["platform"]))
		if platform != "" {
			result[platform] = append(result[platform], row)
		}
	}
	return result
}

// buildAppConfig constructs a CredentialAppConfig from the platform config values.
func buildAppConfig(tenantID, platform string, configMap map[string]string) *models.CredentialAppConfig {
	region := configMap["region"]
	if region == "" {
		region = configMap["country"]
	}
	if region == "" {
		region = "id"
	}

	cfg := &models.CredentialAppConfig{
		TenantID: tenantID,
		Platform: platform,
		Region:   region,
	}

	switch platform {
	case models.PlatformShopee:
		cfg.PartnerID, _ = strconv.ParseInt(configMap["partnerId"], 10, 64)
		cfg.PartnerKey = configMap["partnerKey"]
	case models.PlatformLazada, models.PlatformTiktok:
		cfg.AppKey = configMap["appKey"]
		cfg.AppSecret = configMap["appSecret"]
	}

	cfg.Configured = true
	return cfg
}

// readAndClassify reads all platform_configs rows and splits them into
// credential and non-credential categories. Shared helper for backfill functions.
func readAndClassify(ctx context.Context, db *gorm.DB) (credRows, allRows []inventoryRowData, err error) {
	columns, err := platformConfigColumns(ctx, db)
	if err != nil {
		return nil, nil, fmt.Errorf("discover platform_configs columns: %w", err)
	}
	if len(columns) == 0 {
		return nil, nil, nil
	}

	allRows, err = readPlatformConfigRows(ctx, db, columns)
	if err != nil {
		return nil, nil, fmt.Errorf("read platform_configs: %w", err)
	}

	credRows, _ = ClassifyRows(allRows)
	return credRows, allRows, nil
}

// buildConnectionFromConfig assembles a CredentialConnection from the resolved configMap.
// Handles shopName→sellerName fallback, region→country fallback, TikTok shopCipher,
// and token expiry parsing.
func buildConnectionFromConfig(tenantID, platform string, configMap map[string]string) *models.CredentialConnection {
	storeName := configMap["shopName"]
	if storeName == "" {
		storeName = configMap["sellerName"]
	}

	region := configMap["region"]
	if region == "" {
		region = configMap["country"]
	}

	conn := &models.CredentialConnection{
		TenantID:        tenantID,
		Platform:        platform,
		StoreIdentifier: configMap["shopId"],
		StoreName:       storeName,
		Status:          "connected",
		Region:          region,
		AccessToken:     configMap["accessToken"],
		RefreshToken:    configMap["refreshToken"],
		CreatedBy:       "backfill",
	}

	if platform == "tiktok" {
		conn.ShopCipher = configMap["shopCipher"]
	}

	if v := configMap["tokenExpiry"]; v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			conn.TokenExpiry = n
		}
	}
	if v := configMap["refreshTokenExpiry"]; v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			conn.RefreshExpiry = n
		}
	}

	return conn
}
