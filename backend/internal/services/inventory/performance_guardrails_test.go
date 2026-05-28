package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestPerformanceGuardrails_MaxExportRows verifies the export row cap
// is set to 10000 to prevent unbounded memory usage during CSV/sheet exports.
func TestPerformanceGuardrails_MaxExportRows(t *testing.T) {
	assert.Equal(t, 10000, MaxExportRows,
		"MaxExportRows must be 10000 to prevent unbounded memory usage")
	assert.Greater(t, MaxExportRows, 0,
		"MaxExportRows must be positive")
}

// TestPerformanceGuardrails_ExportToCSV_EmptyRecords verifies that
// the CSV export gracefully handles empty record sets without panicking.
func TestPerformanceGuardrails_ExportToCSV_EmptyRecords(t *testing.T) {
	data, contentType, filename, err := exportToCSV(nil)
	assert.Error(t, err, "exportToCSV should return error for nil records")
	assert.Nil(t, data)
	assert.Empty(t, contentType)
	assert.Empty(t, filename)
}
