package services

type CredentialMigrationMode string

const (
	CredentialMigrationModeDryRun   CredentialMigrationMode = "dry_run"
	CredentialMigrationModeBackfill CredentialMigrationMode = "backfill"
	CredentialMigrationModeCutover  CredentialMigrationMode = "cutover"
	CredentialMigrationModeRollback CredentialMigrationMode = "rollback"
)

type CredentialMigrationOptions struct {
	Mode       CredentialMigrationMode
	BasePath   string
	Actor      string
	ActorRole  string
	AllowWrite bool
}

type CredentialMigrationReport struct {
	Mode                   string                      `json:"mode"`
	CanWrite               bool                        `json:"can_write"`
	CutoverReady           bool                        `json:"cutover_ready"`
	FallbackEnabled        bool                        `json:"fallback_enabled"`
	AbortReasons           []string                    `json:"abort_reasons"`
	Tenants                []CredentialMigrationTenant `json:"tenants"`
	SourceKeyCounts        map[string]int              `json:"source_key_counts"`
	TargetCredentialCounts map[string]int              `json:"target_credential_counts"`
	Backfill               CredentialMigrationBackfill `json:"backfill"`
	Rollback               CredentialMigrationRollback `json:"rollback"`
	Fallback               CredentialMigrationFallback `json:"fallback"`
}

type CredentialMigrationTenant struct {
	TenantID             string   `json:"tenant_id"`
	SchemaName           string   `json:"schema_name"`
	SourceRows           int      `json:"source_rows"`
	TargetConnections    int      `json:"target_connections"`
	TargetAppConfigs     int      `json:"target_app_configs"`
	DuplicateTargets     int      `json:"duplicate_targets"`
	IncompleteSources    int      `json:"incomplete_sources"`
	UnknownShapes        int      `json:"unknown_shapes"`
	UnreadableSecrets    int      `json:"unreadable_secrets"`
	MissingTenantSchema  bool     `json:"missing_tenant_schema"`
	TenantCoverageMatch  bool     `json:"tenant_coverage_match"`
	PlatformCountMatches bool     `json:"platform_count_matches"`
	AbortReasons         []string `json:"abort_reasons,omitempty"`
}

type CredentialMigrationBackfill struct {
	CreatedConnections int `json:"created_connections"`
	CreatedAppConfigs  int `json:"created_app_configs"`
	SkippedExisting    int `json:"skipped_existing"`
}

type CredentialMigrationRollback struct {
	LegacyReadOnly        bool     `json:"legacy_read_only"`
	PreservesMigrated     bool     `json:"preserves_migrated"`
	PreservesNewlyCreated bool     `json:"preserves_newly_created"`
	PreservesRotated      bool     `json:"preserves_rotated"`
	PreservesDisconnected bool     `json:"preserves_disconnected"`
	OperatorSteps         []string `json:"operator_steps"`
}

type CredentialMigrationFallback struct {
	EnvVar         string `json:"env_var"`
	Enabled        bool   `json:"enabled"`
	CanonicalFirst bool   `json:"canonical_first"`
	LegacyReadOnly bool   `json:"legacy_read_only"`
	AuditEvent     string `json:"audit_event"`
}

type legacyCredentialBundle struct {
	TenantID        string
	Platform        string
	StoreIdentifier string
	StoreName       string
	Region          string
	AccessToken     string
	RefreshToken    string
	ShopCipher      string
	TokenExpiry     int64
	AppKey          string
	AppSecret       string
	PartnerID       int64
	PartnerKey      string
	SourceKeys      []string
}
