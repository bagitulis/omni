package services

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCredentialsInventoryServiceInspectTenantAndRedact(t *testing.T) {
	db := testutils.SetupTestPostgres(t)
	ctx := context.Background()

	err := db.Exec(`
		CREATE TABLE platform_configs (
			id TEXT PRIMARY KEY,
			tenant_id TEXT,
			platform TEXT,
			config_key TEXT,
			config_value TEXT,
			data_type TEXT,
			is_encrypted BOOLEAN,
			metadata JSONB,
			shop_id TEXT,
			shop_name TEXT,
			access_token TEXT,
			refresh_token TEXT,
			token_expires_at TIMESTAMPTZ,
			shop_cipher TEXT,
			is_connected BOOLEAN,
			auth_status TEXT,
			last_sync_at TIMESTAMPTZ,
			settings JSONB,
			created_at TIMESTAMPTZ,
			updated_at TIMESTAMPTZ
		)
	`).Error
	require.NoError(t, err)

	err = db.Exec(`
		INSERT INTO platform_configs (id, tenant_id, platform, config_key, config_value, data_type, is_encrypted, metadata, created_at, updated_at)
		VALUES
		('kv-1', 'yumna_bertigamart', 'shopee', 'accessToken', 'encrypted-token-1', 'string', true, '{}'::jsonb, now(), now()),
		('kv-2', 'yumna_bertigamart', 'shopee', 'refreshToken', 'encrypted-refresh-1', 'string', true, '{}'::jsonb, now(), now()),
		('kv-3', 'yumna_bertigamart', 'shopee', 'shopId', '123', 'string', false, '{}'::jsonb, now(), now()),
		('structured-1', 'yumna_bertigamart', 'shopee', NULL, NULL, NULL, NULL, '{}'::jsonb, 'shop-123', 'Demo Shop', 'structured-token-1', 'structured-refresh-1', now(), NULL, true, 'connected', now(), '{}'::jsonb, now(), now()),
		('structured-2', 'yumna_bertigamart', 'shopee', NULL, NULL, NULL, NULL, '{}'::jsonb, 'shop-123', 'Demo Shop', 'structured-token-2', 'structured-refresh-2', now(), NULL, true, 'connected', now(), '{}'::jsonb, now(), now()),
		('mixed-1', 'tika_nusseyba', 'lazada', 'accessToken', 'mixed-secret', 'string', true, '{}'::jsonb, 'shop-777', 'Mixed Shop', 'mixed-access-token', 'mixed-refresh-token', now(), NULL, true, 'connected', now(), '{}'::jsonb, now(), now())
	`).Error
	require.NoError(t, err)

	svc := NewCredentialsInventoryService("")
	rows, err := svc.inspectTenant(ctx, "yumna_bertigamart", db)
	require.NoError(t, err)
	require.NotEmpty(t, rows)

	var kvRow, structuredRow InventoryRow
	for _, row := range rows {
		if row.RowShape == rowShapeKeyValue {
			kvRow = row
		}
		if row.RowShape == rowShapeStructured {
			structuredRow = row
		}
	}

	assert.Equal(t, 3, kvRow.Count)
	assert.Equal(t, 3, kvRow.TargetBackfillCount)
	assert.Equal(t, 0, kvRow.DuplicateStoreCount)
	assert.Empty(t, kvRow.MissingRequiredFields)
	assert.True(t, kvRow.HasSensitiveValuesSeen)

	assert.Equal(t, 2, structuredRow.Count)
	assert.Equal(t, 2, structuredRow.TargetBackfillCount)
	assert.Equal(t, 1, structuredRow.DuplicateStoreCount)
	assert.Empty(t, structuredRow.MissingRequiredFields)

	mixedRows := make([]InventoryRow, 0)
	for _, row := range rows {
		if row.RowShape == rowShapeMixed {
			mixedRows = append(mixedRows, row)
		}
	}
	require.NotEmpty(t, mixedRows)
	assert.Equal(t, 1, mixedRows[0].Count)
	assert.Equal(t, 1, mixedRows[0].TargetBackfillCount)

	report := &InventoryReport{Rows: rows}
	var buf bytes.Buffer
	require.NoError(t, WriteInventoryReport(&buf, report))
	output := buf.String()
	assert.NotContains(t, output, "encrypted-token-1")
	assert.NotContains(t, output, "structured-token-1")
	assert.NotContains(t, output, "mixed-secret")
	assert.NotContains(t, output, "refresh-token")
	assert.Contains(t, output, "missing_required_fields")

	var decoded InventoryReport
	require.NoError(t, json.Unmarshal(buf.Bytes(), &decoded))
	assert.Equal(t, 3, len(decoded.Rows))
	assert.True(t, strings.Contains(output, "target_backfill_count"))

	buf.Reset()
	require.NoError(t, WriteInventoryTextReport(&buf, report))
	textOutput := buf.String()
	assert.Contains(t, textOutput, "Tenant: yumna_bertigamart")
	assert.Contains(t, textOutput, "Row shape: mixed")
	assert.Contains(t, textOutput, "⚠️ MIXED")
	assert.NotContains(t, textOutput, "encrypted-token-1")
	assert.Contains(t, textOutput, "Secret values: [REDACTED -")
}
