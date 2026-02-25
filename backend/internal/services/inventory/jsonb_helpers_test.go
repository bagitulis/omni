package inventory

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetDataMap(t *testing.T) {
	valid := models.InventoryRecord{Data: `{"SKU":"ABC-1","stock":10}`}
	result := GetDataMap(valid)
	assert.Equal(t, "ABC-1", result["SKU"])
	assert.Equal(t, float64(10), result["stock"])

	empty := models.InventoryRecord{Data: ""}
	assert.Empty(t, GetDataMap(empty))

	invalid := models.InventoryRecord{Data: "{"}
	assert.Empty(t, GetDataMap(invalid))
}

func TestGetDataString(t *testing.T) {
	record := models.InventoryRecord{Data: `{"sku":"ABC-1","price":120.5}`}
	assert.Equal(t, "ABC-1", GetDataString(record, "sku"))
	assert.Equal(t, "120.5", GetDataString(record, "price"))
	assert.Equal(t, "", GetDataString(record, "missing"))
}

func TestGetDataFloat(t *testing.T) {
	record := models.InventoryRecord{Data: `{"float_price":120.5,"string_price":"99.9","int_price":7,"invalid":"x"}`}

	assert.Equal(t, 120.5, GetDataFloat(record, "float_price"))
	assert.Equal(t, 99.9, GetDataFloat(record, "string_price"))
	assert.Equal(t, 7.0, GetDataFloat(record, "int_price"))
	assert.Equal(t, 0.0, GetDataFloat(record, "invalid"))
	assert.Equal(t, 0.0, GetDataFloat(record, "missing"))
}

func TestGetDataInt(t *testing.T) {
	record := models.InventoryRecord{Data: `{"float_stock":10.8,"string_stock":"12","int_stock":7,"invalid":"x"}`}

	assert.Equal(t, 10, GetDataInt(record, "float_stock"))
	assert.Equal(t, 12, GetDataInt(record, "string_stock"))
	assert.Equal(t, 7, GetDataInt(record, "int_stock"))
	assert.Equal(t, 0, GetDataInt(record, "invalid"))
	assert.Equal(t, 0, GetDataInt(record, "missing"))
}

func TestSetDataValue(t *testing.T) {
	record := models.InventoryRecord{Data: `{"SKU":"ABC-1"}`}
	err := SetDataValue(&record, "Price", 120.5)
	require.NoError(t, err)

	var data map[string]interface{}
	err = json.Unmarshal([]byte(record.Data), &data)
	require.NoError(t, err)
	assert.Equal(t, "ABC-1", data["SKU"])
	assert.Equal(t, 120.5, data["Price"])

	invalidSource := models.InventoryRecord{Data: "{"}
	err = SetDataValue(&invalidSource, "stock", 5)
	require.NoError(t, err)
	assert.Contains(t, invalidSource.Data, "stock")

	marshalErrorRecord := models.InventoryRecord{}
	err = SetDataValue(&marshalErrorRecord, "bad", func() {})
	assert.Error(t, err)
}

func TestGetSKU(t *testing.T) {
	tests := []struct {
		name     string
		record   models.InventoryRecord
		expected string
	}{
		{
			name:     "from lowercase sku key",
			record:   models.InventoryRecord{Data: `{"sku":"SKU-001"}`, KeyValue: "fallback"},
			expected: "SKU-001",
		},
		{
			name:     "fallback to alternate key",
			record:   models.InventoryRecord{Data: `{"Code":"ALT-123"}`, KeyValue: "fallback"},
			expected: "ALT-123",
		},
		{
			name:     "fallback to key value",
			record:   models.InventoryRecord{Data: `{}`, KeyValue: "FALLBACK-1", KeyColumnName: "name"},
			expected: "FALLBACK-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, GetSKU(tt.record))
		})
	}
}

func TestGetProductName(t *testing.T) {
	recordA := models.InventoryRecord{Data: `{"Nama Barang":"Main Product"}`}
	recordB := models.InventoryRecord{Data: `{"product_name":"Alt Product"}`}
	recordC := models.InventoryRecord{Data: `{}`}

	assert.Equal(t, "Main Product", GetProductName(recordA))
	assert.Equal(t, "Alt Product", GetProductName(recordB))
	assert.Equal(t, "", GetProductName(recordC))
}

func TestGetPrice(t *testing.T) {
	recordA := models.InventoryRecord{Data: `{"HARGA":120.5}`}
	recordB := models.InventoryRecord{Data: `{"price":"99.9"}`}
	recordC := models.InventoryRecord{Data: `{"price":"x"}`}

	assert.Equal(t, 120.5, GetPrice(recordA))
	assert.Equal(t, 99.9, GetPrice(recordB))
	assert.Equal(t, 0.0, GetPrice(recordC))
	assert.Equal(t, 0.0, GetPrice(models.InventoryRecord{Data: `{}`}))
}

func TestGetQuantity(t *testing.T) {
	recordA := models.InventoryRecord{Data: `{"Stock":10}`}
	recordB := models.InventoryRecord{Data: `{"qty":"7"}`}
	recordC := models.InventoryRecord{Data: `{"stok":3}`}
	recordD := models.InventoryRecord{Data: `{"stock":"invalid"}`}

	assert.Equal(t, 10, GetQuantity(recordA))
	assert.Equal(t, 7, GetQuantity(recordB))
	assert.Equal(t, 3, GetQuantity(recordC))
	assert.Equal(t, 0, GetQuantity(recordD))
	assert.Equal(t, 0, GetQuantity(models.InventoryRecord{Data: `{}`}))
}
