package services

const (
	rowShapeKeyValue   = "key_value"
	rowShapeStructured = "structured"
	rowShapeMixed      = "mixed"
	rowShapeUnknown    = "unknown"
)

var requiredSecretColumns = []string{
	"config_value",
	"access_token",
	"refresh_token",
	"app_secret",
	"partner_key",
	"auth_code",
	"ciphertext",
	"decrypted_value",
	"shop_cipher",
}

// InventoryRow summarizes platform_configs rows without exposing secrets.
type InventoryRow struct {
	TenantID               string   `json:"tenant_id"`
	Platform               string   `json:"platform"`
	RowShape               string   `json:"row_shape"`
	Count                  int      `json:"count"`
	MissingRequiredFields  []string `json:"missing_required_fields"`
	DuplicateStoreCount    int      `json:"duplicate_store_count"`
	TargetBackfillCount    int      `json:"target_backfill_count"`
	NeedsBackfill          bool     `json:"needs_backfill"`
	BackfillTarget         string   `json:"backfill_target,omitempty"`
	HasSensitiveValuesSeen bool     `json:"has_sensitive_values_seen"`
	RedactedValueSummary   string   `json:"redacted_value_summary,omitempty"`
}

// InventoryReport is the dry-run output.
type InventoryReport struct {
	Rows []InventoryRow `json:"rows"`
}

type inventoryRowData map[string]any
