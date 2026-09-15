package services

import (
"gorm.io/gorm"

"github.com/omni/backend/internal/config"
"github.com/omni/backend/internal/models"
)
// CredentialApiService handles credential management API operations.
type CredentialApiService struct {
	dbPath string
}

// tenantDB returns a tenant-scoped database connection.
func (s *CredentialApiService) tenantDB(tenantID string) (*gorm.DB, error) {
	return config.GetTenantDB(tenantID, s.dbPath)
}

// CredentialPlatformStatus represents the status of credentials for a platform.
//
// Phase 8 adds the safe (non-secret) fields from CredentialAppConfig's masked
// response so the frontend can render the expiry banner + toggles without a
// second round-trip.
type CredentialPlatformStatus struct {
	Platform      string                                      `json:"platform"`
	Region        string                                      `json:"region,omitempty"`
	Status        string                                      `json:"status"`
	AppConfigured bool                                        `json:"app_configured"`
	Stores        []models.CredentialConnectionMaskedResponse `json:"stores"`
	// Phase 8 — populated when app_config exists for this platform.
	AppConfig *models.CredentialAppConfigMaskedResponse `json:"app_config,omitempty"`
}

// CredentialAppUpsertRequest is the request body for upserting an app credential.
//
// Phase 8 adds Shopee-specific fields (TestPartnerID/TestPartnerKey plus two
// expiry timestamps) and two operational toggles that apply to any platform
// (AppStatus, ActivePartnerEnv). All new fields are optional; when omitted
// the upsert preserves the existing value rather than nulling it.
type CredentialAppUpsertRequest struct {
	Platform        string `json:"-"`
	Region          string `json:"region,omitempty"`
	StoreIdentifier string `json:"store_identifier,omitempty"`
	AppKey          string `json:"app_key,omitempty"`
	AppSecret       string `json:"app_secret,omitempty"`
	PartnerID       int64  `json:"partner_id,omitempty"`
	PartnerKey      string `json:"partner_key,omitempty"`
	Reason          string `json:"reason"`
	// Phase 8 additions (Shopee-focused sandbox + expiry, all optional).
	TestPartnerID           int64  `json:"test_partner_id,omitempty"`
	TestPartnerKey          string `json:"test_partner_key,omitempty"`
	PartnerKeyExpiresAt     string `json:"partner_key_expires_at,omitempty"`      // RFC3339 (e.g. "2026-11-25T22:59:00+07:00")
	TestPartnerKeyExpiresAt string `json:"test_partner_key_expires_at,omitempty"` // RFC3339
	AppStatus               string `json:"app_status,omitempty"`                  // "online" | "offline"
	ActivePartnerEnv        string `json:"active_partner_env,omitempty"`          // "live" | "test" (Shopee only)
}

// CredentialOAuthInitiateRequest is the request body for initiating OAuth.
type CredentialOAuthInitiateRequest struct {
	Platform        string `json:"-"`
	Intent          string `json:"intent"`
	StoreIdentifier string `json:"store_identifier,omitempty"`
	RedirectPath    string `json:"redirect_path"`
}

// CredentialOAuthReconnectRequest is the request body for reconnecting OAuth.
type CredentialOAuthReconnectRequest struct {
	Platform        string `json:"-"`
	StoreIdentifier string `json:"-"`
	Intent          string `json:"intent"`
	RedirectPath    string `json:"redirect_path"`
}

// CredentialConnectionActionRequest is the request body for connection actions.
type CredentialConnectionActionRequest struct {
	Platform        string `json:"-"`
	StoreIdentifier string `json:"-"`
	Reason          string `json:"reason"`
	RevokeRemote    bool   `json:"revoke_remote,omitempty"`
}

// CredentialManualTokenRequest is the request body for manual token entry.
type CredentialManualTokenRequest struct {
	Platform        string `json:"-"`
	StoreIdentifier string `json:"store_identifier"`
	Region          string `json:"region,omitempty"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	ShopCipher      string `json:"shop_cipher,omitempty"`
	Reason          string `json:"reason"`
}

// CredentialOAuthAttemptResponse is the response for OAuth initiation.
type CredentialOAuthAttemptResponse struct {
	AuthURL      string `json:"auth_url"`
	AttemptID    string `json:"attempt_id"`
	ExpiresAt    string `json:"expires_at"`
	AuditEventID string `json:"audit_event_id,omitempty"`
}

// CredentialMutationResponse is the response for credential mutation actions.
type CredentialMutationResponse struct {
	Platform            string `json:"platform"`
	StoreIdentifierMask string `json:"store_identifier_mask,omitempty"`
	Status              string `json:"status"`
	Configured          bool   `json:"configured,omitempty"`
	SecretMask          string `json:"secret_mask,omitempty"`
	AuditEventID        string `json:"audit_event_id,omitempty"`
	RemoteRevokeStatus  string `json:"remote_revoke_status,omitempty"`
	LastRefreshAt       string `json:"last_refresh_at,omitempty"`
	ExpiresAt           string `json:"expires_at,omitempty"`
}

// CredentialAuditListResponse is the response for listing audit events.
type CredentialAuditListResponse struct {
	Events []models.CredentialAuditEvent `json:"events"`
}
