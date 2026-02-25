package inventory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewColumnsService(t *testing.T) {
	service := NewColumnsService(nil, "tenant-a")

	assert.Nil(t, service.db)
	assert.Equal(t, "tenant-a", service.tenantID)
}

func TestGetAvailableColumns(t *testing.T) {
	service := NewColumnsService(nil, "tenant-a")
	columns, err := service.GetAvailableColumns(context.Background())

	assert.NoError(t, err)
	assert.NotEmpty(t, columns)

	first := columns[0]
	assert.Equal(t, "sku", first.Key)
	assert.Equal(t, "SKU", first.Label)
	assert.Equal(t, "string", first.Type)
	assert.True(t, first.Sortable)
	assert.True(t, first.Filterable)

	var hasLastSynced bool
	for _, col := range columns {
		if col.Key == "last_synced" {
			hasLastSynced = true
			break
		}
	}
	assert.True(t, hasLastSynced)
}
