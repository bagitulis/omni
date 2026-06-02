package services

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClassifyDecryptError(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected string
	}{
		{
			name:     "decryption failed",
			errMsg:   "decryption failed",
			expected: "decryption_failed",
		},
		{
			name:     "invalid format",
			errMsg:   "invalid token format",
			expected: "invalid_format",
		},
		{
			name:     "expired token",
			errMsg:   "token expired",
			expected: "token_expired",
		},
		{
			name:     "unknown error",
			errMsg:   "something else entirely",
			expected: "unknown_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := classifyDecryptError(assert.AnError)
			// Override with test-specific error
			result = classifyDecryptError(&testError{tt.errMsg})
			assert.Equal(t, tt.expected, result)
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		target   string
		expected bool
	}{
		{"found", []string{"a", "b", "c"}, "b", true},
		{"not found", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"nil slice", nil, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, containsString(tt.slice, tt.target))
		})
	}
}

func TestFormatPreflightReport(t *testing.T) {
	report := &PreflightValidationReport{
		TotalEncrypted:   10,
		DecryptSuccess:   8,
		DecryptFailed:    2,
		FernetKeyChanged: false,
		BlockedTenants:   []string{"tenant_blocked"},
		TenantResults: map[string]TenantPreflight{
			"tenant_pass": {
				TenantID:       "tenant_pass",
				SchemaName:     "tenant_tenant_pass",
				EncryptedCount: 8,
				FailedCount:    0,
				Blocked:        false,
			},
			"tenant_blocked": {
				TenantID:       "tenant_blocked",
				SchemaName:     "tenant_tenant_blocked",
				EncryptedCount: 2,
				FailedCount:    2,
				Blocked:        true,
				FailedKeys:     []string{"accessToken", "refreshToken"},
			},
		},
		Failures: []PreflightDecryptFailure{
			{TenantID: "tenant_blocked", Platform: "shopee", ConfigKey: "accessToken", ErrorType: "decryption_failed", IsRequired: true},
			{TenantID: "tenant_blocked", Platform: "shopee", ConfigKey: "refreshToken", ErrorType: "decryption_failed", IsRequired: true},
		},
	}

	output := FormatPreflightReport(report)
	assert.Contains(t, output, "PRE-BACKFILL PREFLIGHT VALIDATION REPORT")
	assert.Contains(t, output, "Total Encrypted Rows: 10")
	assert.Contains(t, output, "Decrypt Success: 8")
	assert.Contains(t, output, "Decrypt Failed: 2")
	assert.Contains(t, output, "Fernet Key Changed: false")
	assert.Contains(t, output, "tenant_pass")
	assert.Contains(t, output, "tenant_blocked")
	assert.Contains(t, output, "BLOCKED")
	assert.Contains(t, output, "PASS")
	assert.Contains(t, output, "MIGRATION BLOCKED")
	assert.NotContains(t, output, "CRITICAL") // Not all failed
}

func TestFormatPreflightReportFernetKeyChanged(t *testing.T) {
	report := &PreflightValidationReport{
		TotalEncrypted:   6,
		DecryptSuccess:   0,
		DecryptFailed:    6,
		FernetKeyChanged: true,
		BlockedTenants:   []string{"tenant_a"},
		TenantResults:    map[string]TenantPreflight{},
		Failures:         []PreflightDecryptFailure{},
	}

	output := FormatPreflightReport(report)
	assert.Contains(t, output, "CRITICAL: ALL encrypted rows failed")
	assert.Contains(t, output, "Fernet Key Changed: true")
}

func TestRequiredCredentialFields(t *testing.T) {
	expected := []string{"partnerKey", "appKey", "appSecret", "accessToken", "refreshToken"}
	for _, field := range expected {
		assert.True(t, requiredCredentialFields[field], "field %s should be required", field)
	}
	assert.False(t, requiredCredentialFields["shopId"], "shopId should not be required")
	assert.False(t, requiredCredentialFields["region"], "region should not be required")
}

func TestValidateEncryptedValuesPreflight_EmptyKey(t *testing.T) {
	_, err := ValidateEncryptedValuesPreflight(context.Background(), nil, "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ENCRYPTION_KEY is empty")
}

func TestValidateEncryptedValuesPreflight_InvalidKey(t *testing.T) {
	_, err := ValidateEncryptedValuesPreflight(context.Background(), nil, "not-a-valid-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create encryption service")
}

func TestDecryptRoundTrip_RequiredFields(t *testing.T) {
	key, err := utils.GenerateKey()
	require.NoError(t, err)

	enc, err := utils.NewEncryptionService(key)
	require.NoError(t, err)

	// Encrypt and verify all required fields can be decrypted
	for _, field := range []string{"partnerKey", "appKey", "appSecret", "accessToken", "refreshToken"} {
		t.Run(field, func(t *testing.T) {
			plaintext := "test-value-for-" + field
			encrypted, err := enc.Encrypt(plaintext)
			require.NoError(t, err)

			decrypted, err := enc.Decrypt(encrypted)
			require.NoError(t, err)
			assert.Equal(t, plaintext, decrypted)
		})
	}
}
