package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// credentialKeys classifies config_key names as credential (app or connection level).
var credentialKeys = map[string]bool{
	// App-level credentials
	"partnerId":  true,
	"partnerKey": true,
	"appKey":     true,
	"appSecret":  true,
	// Connection-level credentials
	"accessToken":  true,
	"refreshToken": true,
	"shopCipher":   true,
}

// appCredentialKeys defines the required APP_CREDENTIAL keys per platform.
var appCredentialKeys = map[string][]string{
	models.PlatformShopee: {"partnerId", "partnerKey"},
	models.PlatformLazada: {"appKey", "appSecret"},
	models.PlatformTiktok: {"appKey", "appSecret"},
}

// BackfillValidationReport tracks decrypt failures during backfill (redacted).
type BackfillValidationReport struct {
	Failures []BackfillDecryptFailure `json:"failures"`
}

// BackfillDecryptFailure records a single decrypt failure (no plaintext secrets).
type BackfillDecryptFailure struct {
	TenantID   string `json:"tenant_id"`
	Platform   string `json:"platform"`
	ConfigKey  string `json:"config_key"`
	IsRequired bool   `json:"is_required"`
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

// BackfillAppConfigs reads platform_configs for the given tenant, identifies
// platforms with app-level credentials, and inserts corresponding
// credential_app_configs rows. Idempotent: skips platforms that already have
// an app config. Returns error if required app credential keys are missing.
func BackfillAppConfigs(ctx context.Context, db *gorm.DB, tenantID string) error {
	if tenantID == "" {
		return nil
	}

	credRows, _, err := readAndClassify(ctx, db)
	if err != nil {
		return err
	}
	if len(credRows) == 0 {
		return nil
	}

	credRepo := repositories.NewCredentialRepository(db)
	platformRows := groupByPlatform(credRows)

	for platform, pRows := range platformRows {
		requiredKeys, ok := appCredentialKeys[platform]
		if !ok {
			log.Warn().Str("platform", platform).Msg("Unknown platform, skipping app config backfill")
			continue
		}

		existing, err := credRepo.GetAppConfig(ctx, tenantID, platform)
		if err != nil {
			return fmt.Errorf("check existing app config for %s: %w", platform, err)
		}
		if existing != nil {
			log.Debug().Str("platform", platform).Msg("App config already exists, skipping")
			continue
		}

		configMap := buildConfigMap(pRows)

		// Validate encrypted credential fields can be decrypted
		valReport := validateCredentialConfigMap(configMap, requiredKeys)
		if hasBlockingFailures(valReport) {
			emitBackfillValidationReport(valReport, tenantID, "app_config_"+platform)
			return fmt.Errorf("corrupted encrypted credentials for platform %s — %d required field(s) failed to decrypt", platform, countBlockingFailures(valReport))
		}

		for _, requiredKey := range requiredKeys {
			if configMap[requiredKey] == "" {
				return fmt.Errorf("missing required app credential key %q for platform %s", requiredKey, platform)
			}
		}

		cfg := buildAppConfig(tenantID, platform, configMap)
		if err := credRepo.UpsertAppConfig(ctx, cfg); err != nil {
			return fmt.Errorf("upsert app config for %s: %w", platform, err)
		}

		log.Info().
			Str("tenant_id", tenantID).
			Str("platform", platform).
			Msg("Backfilled app config")
	}

	return nil
}

// BackfillConnections reads platform_configs for the given tenant and creates
// credential_connections rows for each platform-shop pair with tokens present.
// Idempotent: skips platforms that already have a connection.
// Platforms without a shopId are skipped. Missing tokens trigger a warning log.
func BackfillConnections(ctx context.Context, db *gorm.DB, tenantID string) error {
	if tenantID == "" {
		return nil
	}

	_, allRows, err := readAndClassify(ctx, db)
	if err != nil {
		return err
	}
	if len(allRows) == 0 {
		return nil
	}

	credRepo := repositories.NewCredentialRepository(db)
	platformRows := groupByPlatform(allRows)

	for platform, pRows := range platformRows {
		configMap := buildConfigMap(pRows)

		shopID, ok := configMap["shopId"]
		if !ok || shopID == "" {
			continue
		}

		accessToken := configMap["accessToken"]
		refreshToken := configMap["refreshToken"]
		if accessToken == "" && refreshToken == "" {
			log.Warn().
				Str("tenant_id", tenantID).
				Str("platform", platform).
				Str("shop_id", shopID).
				Msg("Skipping connection backfill — no tokens found")
			continue
		}

		// Validate encrypted credential fields can be decrypted
		connRequiredKeys := []string{"accessToken", "refreshToken"}
		valReport := validateCredentialConfigMap(configMap, connRequiredKeys)
		if hasBlockingFailures(valReport) {
			emitBackfillValidationReport(valReport, tenantID, "connection_"+platform+"_"+shopID)
			log.Error().
				Str("tenant_id", tenantID).
				Str("platform", platform).
				Str("shop_id", shopID).
				Msg("Skipping connection — required credential fields failed to decrypt")
			continue
		}
		// Optional field: shopCipher — silently clear if corrupted
		if platform == "tiktok" {
			if v := configMap["shopCipher"]; v != "" && utils.IsEncrypted(v) {
				if enc, encErr := utils.NewEncryptionService(os.Getenv("ENCRYPTION_KEY")); encErr == nil {
					if _, decErr := enc.Decrypt(v); decErr != nil {
						log.Warn().Str("tenant_id", tenantID).Str("platform", platform).Msg("[Backfill] shopCipher decrypt failed — inserting NULL")
						configMap["shopCipher"] = ""
					}
				}
			}
		}

		existing, err := credRepo.GetConnection(ctx, tenantID, platform, shopID)
		if err != nil {
			return fmt.Errorf("check existing connection for %s/%s: %w", platform, shopID, err)
		}
		if existing != nil {
			continue
		}

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
			StoreIdentifier: shopID,
			StoreName:       storeName,
			Status:          "connected",
			Region:          region,
			AccessToken:     accessToken,
			RefreshToken:    refreshToken,
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

		if err := credRepo.CreateConnection(ctx, conn); err != nil {
			return fmt.Errorf("create connection for %s/%s: %w", platform, shopID, err)
		}

		log.Info().
			Str("tenant_id", tenantID).
			Str("platform", platform).
			Str("shop_id", shopID).
			Msg("Backfilled connection")
	}

	return nil
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

// buildConfigMap creates a key-value map from platform config rows.
func buildConfigMap(rows []inventoryRowData) map[string]string {
	m := make(map[string]string, len(rows))
	for _, row := range rows {
		key := strings.TrimSpace(asString(row["config_key"]))
		value := asString(row["config_value"])
		if key != "" {
			m[key] = value
		}
	}
	return m
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


