package services

import "github.com/omni/backend/internal/models"

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
