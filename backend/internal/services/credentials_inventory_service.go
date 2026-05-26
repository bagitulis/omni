package services

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CredentialsInventoryService inspects tenant platform configs in read-only mode.
type CredentialsInventoryService struct {
	tenantService *TenantService
}

// NewCredentialsInventoryService creates the dry-run inventory service.
func NewCredentialsInventoryService(basePath string) *CredentialsInventoryService {
	return &CredentialsInventoryService{
		tenantService: NewTenantService(basePath),
	}
}

// Run executes the dry-run inventory for all active tenants.
func (s *CredentialsInventoryService) Run(ctx context.Context) (*InventoryReport, error) {
	tenants, err := s.tenantService.GetAvailableTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}

	rows := make([]InventoryRow, 0)
	for _, tenant := range tenants {
		db, err := s.tenantService.GetTenantDB(tenant.ID)
		if err != nil {
			log.Warn().Err(err).Str("tenant_id", tenant.ID).Msg("Skipping tenant during inventory")
			continue
		}

		tenantRows, err := s.inspectTenant(ctx, tenant.ID, db)
		if err != nil {
			return nil, fmt.Errorf("inspect tenant %s: %w", tenant.ID, err)
		}
		rows = append(rows, tenantRows...)
	}

	return &InventoryReport{Rows: rows}, nil
}

func (s *CredentialsInventoryService) inspectTenant(ctx context.Context, tenantID string, db *gorm.DB) ([]InventoryRow, error) {
	columns, err := platformConfigColumns(ctx, db)
	if err != nil {
		return nil, err
	}
	if len(columns) == 0 {
		return []InventoryRow{{TenantID: tenantID, RowShape: rowShapeUnknown, MissingRequiredFields: []string{"platform_configs_table_missing"}}}, nil
	}

	rows, err := readPlatformConfigRows(ctx, db, columns)
	if err != nil {
		return nil, err
	}

	storeCounts := duplicateStoreCounts(tenantID, rows)
	result := make([]InventoryRow, 0, len(rows))
	for _, row := range rows {
		summary := summarizeInventoryRow(tenantID, row, storeCounts)
		result = append(result, summary)
	}

	return aggregateInventoryRows(result), nil
}

func platformConfigColumns(ctx context.Context, db *gorm.DB) ([]string, error) {
	var cols []string
	if err := db.WithContext(ctx).Raw(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_name = 'platform_configs'
		ORDER BY ordinal_position
	`).Scan(&cols).Error; err != nil {
		return nil, fmt.Errorf("list platform_configs columns: %w", err)
	}
	return cols, nil
}

func readPlatformConfigRows(ctx context.Context, db *gorm.DB, columns []string) ([]inventoryRowData, error) {
	rows, err := db.WithContext(ctx).Table("platform_configs").Rows()
	if err != nil {
		return nil, fmt.Errorf("scan platform_configs rows: %w", err)
	}
	defer rows.Close()

	result := make([]inventoryRowData, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("scan platform_configs row: %w", err)
		}
		row := make(inventoryRowData, len(columns))
		for i, column := range columns {
			if b, ok := values[i].([]byte); ok {
				row[column] = string(b)
				continue
			}
			if values[i] == nil {
				row[column] = nil
				continue
			}
			if nv, ok := values[i].(sql.NullString); ok {
				if nv.Valid {
					row[column] = nv.String
				}
				continue
			}
			row[column] = values[i]
		}
		result = append(result, row)
	}
	return result, nil
}

func summarizeInventoryRow(tenantID string, row inventoryRowData, storeCounts map[string]int) InventoryRow {
	platform := strings.TrimSpace(asString(row["platform"]))
	shape := detectRowShape(row)
	missing := missingRequiredFields(shape, row)
	storeKey := storeIdentity(tenantID, row, shape)
	duplicateCount := 0
	if storeKey != "" {
		duplicateCount = max(0, storeCounts[storeKey]-1)
	}
	needsBackfill := len(missing) > 0 || shape == rowShapeKeyValue || shape == rowShapeMixed || shape == rowShapeUnknown
	backfillTarget := ""
	if needsBackfill {
		backfillTarget = "structured"
		if shape == rowShapeStructured {
			backfillTarget = "complete"
		}
	}

	return InventoryRow{
		TenantID:               tenantID,
		Platform:               platform,
		RowShape:               shape,
		Count:                  1,
		MissingRequiredFields:  missing,
		DuplicateStoreCount:    duplicateCount,
		TargetBackfillCount:    boolToInt(needsBackfill),
		NeedsBackfill:          needsBackfill,
		BackfillTarget:         backfillTarget,
		HasSensitiveValuesSeen: hasSensitiveValue(row),
		RedactedValueSummary:   redactedSummary(row),
	}
}

func detectRowShape(row inventoryRowData) string {
	keyValueFields := []string{"config_key", "config_value"}
	structuredFields := []string{"tenant_id", "shop_id", "shop_name", "access_token", "refresh_token"}
	keyValuePresent := 0
	for _, key := range keyValueFields {
		if presentAndNonEmpty(row[key]) {
			keyValuePresent++
		}
	}
	structuredPresent := 0
	for _, key := range structuredFields {
		if presentAndNonEmpty(row[key]) {
			structuredPresent++
		}
	}
	if keyValuePresent > 0 && structuredPresent > 0 {
		return rowShapeMixed
	}
	if keyValuePresent > 0 {
		return rowShapeKeyValue
	}
	if structuredPresent > 0 {
		return rowShapeStructured
	}
	return rowShapeUnknown
}

func missingRequiredFields(shape string, row inventoryRowData) []string {
	required := map[string][]string{
		rowShapeKeyValue:   {"platform", "config_key", "config_value"},
		rowShapeStructured: {"tenant_id", "platform", "shop_id", "shop_name", "access_token", "refresh_token"},
		rowShapeMixed:      {"platform", "config_key", "tenant_id", "shop_id", "access_token", "refresh_token"},
	}
	fields := required[shape]
	missing := make([]string, 0)
	for _, field := range fields {
		if !presentAndNonEmpty(row[field]) {
			missing = append(missing, field)
		}
	}
	sort.Strings(missing)
	return missing
}

func storeIdentity(tenantID string, row inventoryRowData, shape string) string {
	platform := strings.TrimSpace(asString(row["platform"]))
	switch shape {
	case rowShapeKeyValue:
		return strings.Join([]string{tenantID, platform, strings.TrimSpace(asString(row["config_key"]))}, "|")
	case rowShapeStructured, rowShapeMixed:
		shopID := strings.TrimSpace(asString(row["shop_id"]))
		if shopID == "" {
			shopID = strings.TrimSpace(asString(row["shop_name"]))
		}
		return strings.Join([]string{tenantID, platform, shopID}, "|")
	default:
		return strings.Join([]string{tenantID, platform, strings.TrimSpace(asString(row["id"]))}, "|")
	}
}

func duplicateStoreCounts(tenantID string, rows []inventoryRowData) map[string]int {
	counts := make(map[string]int)
	for _, row := range rows {
		rowTenantID := strings.TrimSpace(asString(row["tenant_id"]))
		if rowTenantID == "" {
			rowTenantID = tenantID
		}
		counts[storeIdentity(rowTenantID, row, detectRowShape(row))]++
	}
	return counts
}

func hasSensitiveValue(row inventoryRowData) bool {
	for _, key := range requiredSecretColumns {
		if presentAndNonEmpty(row[key]) {
			return true
		}
	}
	return false
}

func redactedSummary(row inventoryRowData) string {
	maxBytes := 0
	for _, key := range requiredSecretColumns {
		if value := strings.TrimSpace(asString(row[key])); value != "" {
			if len(value) > maxBytes {
				maxBytes = len(value)
			}
		}
	}
	if maxBytes == 0 {
		return ""
	}
	return fmt.Sprintf("[REDACTED - %d bytes]", maxBytes)
}
