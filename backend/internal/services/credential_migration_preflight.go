package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
)

// PreflightValidationReport holds results of encrypted value validation
type PreflightValidationReport struct {
	Timestamp        time.Time                    `json:"timestamp"`
	TotalEncrypted   int                          `json:"total_encrypted"`
	DecryptSuccess   int                          `json:"decrypt_success"`
	DecryptFailed    int                          `json:"decrypt_failed"`
	FernetKeyChanged bool                         `json:"fernet_key_changed"`
	BlockedTenants   []string                     `json:"blocked_tenants"`
	Failures         []PreflightDecryptFailure    `json:"failures"`
	TenantResults    map[string]TenantPreflight   `json:"tenant_results"`
}

// TenantPreflight holds per-tenant validation results
type TenantPreflight struct {
	TenantID       string   `json:"tenant_id"`
	SchemaName     string   `json:"schema_name"`
	EncryptedCount int      `json:"encrypted_count"`
	FailedCount    int      `json:"failed_count"`
	Blocked        bool     `json:"blocked"`
	FailedKeys     []string `json:"failed_keys,omitempty"`
}

// PreflightDecryptFailure holds a single failure detail (redacted)
type PreflightDecryptFailure struct {
	TenantID   string `json:"tenant_id"`
	Platform   string `json:"platform"`
	ConfigKey  string `json:"config_key"`
	ErrorType  string `json:"error_type"`
	IsRequired bool   `json:"is_required"`
}

// requiredCredentialFields defines fields that block migration if decrypt fails
var requiredCredentialFields = map[string]bool{
	"partnerKey":   true,
	"appKey":       true,
	"appSecret":    true,
	"accessToken":  true,
	"refreshToken": true,
}

// ValidateEncryptedValuesPreflight checks all encrypted platform_configs rows
// across all tenant schemas to verify they can be decrypted.
// Returns error if any required credential field fails to decrypt.
func ValidateEncryptedValuesPreflight(ctx context.Context, tenantService *TenantService, encKey string) (*PreflightValidationReport, error) {
	report := &PreflightValidationReport{
		Timestamp:     time.Now(),
		TenantResults: map[string]TenantPreflight{},
	}

	if encKey == "" {
		return report, fmt.Errorf("ENCRYPTION_KEY is empty — cannot validate encrypted values")
	}

	enc, err := utils.NewEncryptionService(encKey)
	if err != nil {
		return report, fmt.Errorf("failed to create encryption service: %w", err)
	}

	tenants, err := tenantService.GetAvailableTenants(ctx)
	if err != nil {
		return report, fmt.Errorf("list tenants: %w", err)
	}

	totalFailures := 0
	totalEncrypted := 0

	for _, tenant := range tenants {
		tenantResult := validateTenantEncryptedValues(ctx, tenantService, tenant.ID, enc, report)
		report.TenantResults[tenant.ID] = tenantResult
		totalEncrypted += tenantResult.EncryptedCount
		totalFailures += tenantResult.FailedCount
		if tenantResult.Blocked {
			report.BlockedTenants = append(report.BlockedTenants, tenant.ID)
		}
	}

	report.TotalEncrypted = totalEncrypted
	report.DecryptFailed = totalFailures
	report.DecryptSuccess = totalEncrypted - totalFailures

	// Fernet key change detection: if ALL encrypted rows fail, the key likely changed
	if totalEncrypted > 0 && totalFailures == totalEncrypted {
		report.FernetKeyChanged = true
		log.Error().
			Int("total_encrypted", totalEncrypted).
			Msg("[Preflight] CRITICAL: ALL encrypted rows failed decryption — Fernet key may have changed")
	}

	return report, nil
}

// validateTenantEncryptedValues validates encrypted values for a single tenant
func validateTenantEncryptedValues(ctx context.Context, tenantService *TenantService, tenantID string, enc *utils.EncryptionService, report *PreflightValidationReport) TenantPreflight {
	result := TenantPreflight{
		TenantID:   tenantID,
		SchemaName: "tenant_" + tenantID,
	}

	db, err := tenantService.GetTenantDB(tenantID)
	if err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("[Preflight] Cannot access tenant schema")
		result.Blocked = true
		result.FailedKeys = append(result.FailedKeys, "SCHEMA_INACCESSIBLE")
		report.Failures = append(report.Failures, PreflightDecryptFailure{
			TenantID:   tenantID,
			ErrorType:  "schema_inaccessible",
			IsRequired: true,
		})
		return result
	}

	// Query all encrypted rows
	var rows []struct {
		Platform    string
		ConfigKey   string
		ConfigValue string
		IsEncrypted bool
	}

	if err := db.WithContext(ctx).Table("platform_configs").
		Select("platform, config_key, config_value, is_encrypted").
		Where("is_encrypted = ?", true).
		Find(&rows).Error; err != nil {
		log.Warn().Err(err).Str("tenant_id", tenantID).Msg("[Preflight] Failed to query platform_configs")
		result.Blocked = true
		report.Failures = append(report.Failures, PreflightDecryptFailure{
			TenantID:   tenantID,
			ErrorType:  "query_failed",
			IsRequired: true,
		})
		return result
	}

	result.EncryptedCount = len(rows)

	for _, row := range rows {
		if row.ConfigValue == "" {
			continue
		}

		_, err := enc.Decrypt(row.ConfigValue)
		if err != nil {
			result.FailedCount++
			isRequired := requiredCredentialFields[row.ConfigKey]

			// Log redacted — NO plaintext values
			log.Warn().
				Str("tenant_id", tenantID).
				Str("platform", row.Platform).
				Str("config_key", row.ConfigKey).
				Bool("is_required", isRequired).
				Msg("[Preflight] Decrypt failed for encrypted config value")

			report.Failures = append(report.Failures, PreflightDecryptFailure{
				TenantID:   tenantID,
				Platform:   row.Platform,
				ConfigKey:  row.ConfigKey,
				ErrorType:  classifyDecryptError(err),
				IsRequired: isRequired,
			})

			if isRequired {
				result.Blocked = true
				if !containsString(result.FailedKeys, row.ConfigKey) {
					result.FailedKeys = append(result.FailedKeys, row.ConfigKey)
				}
			}
		}
	}

	if result.Blocked {
		log.Error().
			Str("tenant_id", tenantID).
			Int("failed_count", result.FailedCount).
			Strs("failed_keys", result.FailedKeys).
			Msg("[Preflight] Tenant BLOCKED — required credential fields failed to decrypt")
	}

	return result
}

// classifyDecryptError returns a redacted error type for logging
func classifyDecryptError(err error) string {
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "decryption failed"):
		return "decryption_failed"
	case strings.Contains(errMsg, "invalid"):
		return "invalid_format"
	case strings.Contains(errMsg, "expired"):
		return "token_expired"
	default:
		return "unknown_error"
	}
}

func containsString(slice []string, target string) bool {
	for _, s := range slice {
		if s == target {
			return true
		}
	}
	return false
}

// FormatPreflightReport generates a human-readable report string
func FormatPreflightReport(report *PreflightValidationReport) string {
	var sb strings.Builder

	sb.WriteString("=== PRE-BACKFILL PREFLIGHT VALIDATION REPORT ===\n")
	sb.WriteString(fmt.Sprintf("Timestamp: %s\n", report.Timestamp.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Total Encrypted Rows: %d\n", report.TotalEncrypted))
	sb.WriteString(fmt.Sprintf("Decrypt Success: %d\n", report.DecryptSuccess))
	sb.WriteString(fmt.Sprintf("Decrypt Failed: %d\n", report.DecryptFailed))
	sb.WriteString(fmt.Sprintf("Fernet Key Changed: %v\n", report.FernetKeyChanged))
	sb.WriteString(fmt.Sprintf("Blocked Tenants: %v\n", report.BlockedTenants))
	sb.WriteString("\n")

	sb.WriteString("--- Tenant Results ---\n")
	for _, tr := range report.TenantResults {
		status := "PASS"
		if tr.Blocked {
			status = "BLOCKED"
		}
		sb.WriteString(fmt.Sprintf("  %s (%s): %d encrypted, %d failed [%s]\n",
			tr.TenantID, tr.SchemaName, tr.EncryptedCount, tr.FailedCount, status))
		if len(tr.FailedKeys) > 0 {
			sb.WriteString(fmt.Sprintf("    Failed keys: %v\n", tr.FailedKeys))
		}
	}

	if len(report.Failures) > 0 {
		sb.WriteString("\n--- Decrypt Failures (redacted) ---\n")
		for _, f := range report.Failures {
			sb.WriteString(fmt.Sprintf("  tenant=%s platform=%s key=%s error=%s required=%v\n",
				f.TenantID, f.Platform, f.ConfigKey, f.ErrorType, f.IsRequired))
		}
	}

	if report.FernetKeyChanged {
		sb.WriteString("\n*** CRITICAL: ALL encrypted rows failed — Fernet key likely changed! ***\n")
		sb.WriteString("*** Migration BLOCKED until encryption key is restored. ***\n")
	}

	if len(report.BlockedTenants) > 0 {
		sb.WriteString(fmt.Sprintf("\n*** MIGRATION BLOCKED for tenants: %v ***\n", report.BlockedTenants))
	} else {
		sb.WriteString("\n*** ALL TENANTS PASS — migration may proceed. ***\n")
	}

	return sb.String()
}

// SavePreflightReport writes the report to a file
func SavePreflightReport(report *PreflightValidationReport, path string) error {
	content := FormatPreflightReport(report)
	return os.WriteFile(path, []byte(content), 0644)
}
