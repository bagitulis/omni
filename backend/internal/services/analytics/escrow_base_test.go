package analytics

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBaseEscrowServiceAndTable(t *testing.T) {
	service := NewBaseEscrowService(nil, "tenant_123", "/tmp/test.db")

	assert.Nil(t, service.DB)
	assert.Equal(t, "tenant_123", service.TenantID)
	assert.Equal(t, "tenant_tenant_123", service.Schema)
	assert.Equal(t, "/tmp/test.db", service.DBPath)
	assert.Equal(t, "tenant_tenant_123.orders", service.Table("orders"))
}

func TestParseFloat(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{name: "empty string", input: "", expected: 0},
		{name: "valid decimal", input: "123.45", expected: 123.45},
		{name: "invalid value", input: "abc", expected: 0},
		{name: "scientific notation", input: "12e2", expected: 1200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, ParseFloat(tt.input))
		})
	}
}

func TestPointerHelpers(t *testing.T) {
	str := StringPtr("value")
	int64Val := Int64Ptr(123)
	intVal := IntPtr(456)

	if assert.NotNil(t, str) {
		assert.Equal(t, "value", *str)
	}
	if assert.NotNil(t, int64Val) {
		assert.Equal(t, int64(123), *int64Val)
	}
	if assert.NotNil(t, intVal) {
		assert.Equal(t, 456, *intVal)
	}
}

func TestEscrowSyncTables(t *testing.T) {
	shopee := ShopeeEscrowTables()
	assert.Equal(t, EscrowSyncTables{
		OrderTable: "shopee_escrow_orders",
		ItemTable:  "shopee_escrow_items",
		SyncTable:  "shopee_escrow_sync",
	}, shopee)

	tiktok := TiktokEscrowTables()
	assert.Equal(t, EscrowSyncTables{
		OrderTable: "tiktok_escrow_orders",
		ItemTable:  "tiktok_escrow_items",
		SyncTable:  "tiktok_escrow_sync",
	}, tiktok)
}

func TestSyncResultWithProgressJSONTags(t *testing.T) {
	input := SyncResultWithProgress{
		TotalOrders:     100,
		ProcessedOrders: 90,
		FailedOrders:    10,
		TotalItems:      250,
		Message:         "completed",
		Cancelled:       false,
	}

	data, err := json.Marshal(input)
	assert.NoError(t, err)

	var out map[string]interface{}
	err = json.Unmarshal(data, &out)
	assert.NoError(t, err)

	assert.Contains(t, out, "total_orders")
	assert.Contains(t, out, "processed_orders")
	assert.Contains(t, out, "failed_orders")
	assert.Contains(t, out, "total_items")
	assert.Contains(t, out, "message")
	assert.Contains(t, out, "cancelled")
	assert.NotContains(t, out, "TotalOrders")
	assert.NotContains(t, out, "ProcessedOrders")
}
