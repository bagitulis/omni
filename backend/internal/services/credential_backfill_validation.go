package services

import (
	"os"

	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
)

// validateCredentialConfigMap checks that all encrypted credential values in
// the config map can be decrypted. Returns a report of failures.
// Required keys that fail to decrypt are flagged as blocking.
func validateCredentialConfigMap(configMap map[string]string, requiredKeys []string) *BackfillValidationReport {
	report := &BackfillValidationReport{}
	requiredSet := make(map[string]bool, len(requiredKeys))
	for _, k := range requiredKeys {
		requiredSet[k] = true
	}

	enc, err := utils.NewEncryptionService(os.Getenv("ENCRYPTION_KEY"))
	if err != nil {
		log.Warn().Err(err).Msg("[Backfill] Cannot create encryption service — skipping decrypt validation")
		return report
	}

	for key, value := range configMap {
		if !credentialKeys[key] || value == "" || !utils.IsEncrypted(value) {
			continue
		}
		if _, decErr := enc.Decrypt(value); decErr != nil {
			report.Failures = append(report.Failures, BackfillDecryptFailure{
				ConfigKey:  key,
				IsRequired: requiredSet[key],
			})
		}
	}
	return report
}

// emitBackfillValidationReport logs a redacted summary of decrypt failures.
func emitBackfillValidationReport(report *BackfillValidationReport, tenantID, context string) {
	if report == nil || len(report.Failures) == 0 {
		return
	}
	log.Error().
		Str("tenant_id", tenantID).
		Str("context", context).
		Int("total_failures", len(report.Failures)).
		Msg("[Backfill] Encrypted value validation failures detected")
}

// hasBlockingFailures returns true if the report contains any required-field failures.
func hasBlockingFailures(report *BackfillValidationReport) bool {
	if report == nil {
		return false
	}
	for _, f := range report.Failures {
		if f.IsRequired {
			return true
		}
	}
	return false
}

// countBlockingFailures returns the number of required-field failures.
func countBlockingFailures(report *BackfillValidationReport) int {
	if report == nil {
		return 0
	}
	count := 0
	for _, f := range report.Failures {
		if f.IsRequired {
			count++
		}
	}
	return count
}
