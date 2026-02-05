package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractHeaders(t *testing.T) {
	t.Run("extracts keys from map", func(t *testing.T) {
		data := map[string]interface{}{
			"name":  "Product A",
			"price": 100,
			"stock": 50,
		}

		headers := extractHeaders(data)

		assert.Len(t, headers, 3)
		assert.Contains(t, headers, "name")
		assert.Contains(t, headers, "price")
		assert.Contains(t, headers, "stock")
	})

	t.Run("returns empty slice for empty map", func(t *testing.T) {
		data := map[string]interface{}{}

		headers := extractHeaders(data)

		assert.Len(t, headers, 0)
		assert.NotNil(t, headers)
	})
}

func TestBuildSheetValues(t *testing.T) {
	t.Run("builds 2D array with headers and data", func(t *testing.T) {
		headers := []string{"name", "price", "stock"}
		data := []map[string]interface{}{
			{"name": "Product A", "price": 100, "stock": 50},
			{"name": "Product B", "price": 200, "stock": 30},
		}

		values := buildSheetValues(data, headers)

		assert.Len(t, values, 3) // 1 header + 2 data rows
		assert.Equal(t, headers[0], values[0][0])
		assert.Equal(t, headers[1], values[0][1])
		assert.Equal(t, headers[2], values[0][2])
		assert.Equal(t, "Product A", values[1][0])
		assert.Equal(t, 100, values[1][1])
		assert.Equal(t, "Product B", values[2][0])
	})

	t.Run("handles missing keys with empty string", func(t *testing.T) {
		headers := []string{"name", "price", "description"}
		data := []map[string]interface{}{
			{"name": "Product A", "price": 100},
		}

		values := buildSheetValues(data, headers)

		assert.Len(t, values, 2)
		assert.Equal(t, "", values[1][2]) // description is missing
	})

	t.Run("returns header only for empty data", func(t *testing.T) {
		headers := []string{"name", "price"}
		data := []map[string]interface{}{}

		values := buildSheetValues(data, headers)

		assert.Len(t, values, 1) // Only header row
	})
}

func TestParseSheetHeaders(t *testing.T) {
	t.Run("parses string headers", func(t *testing.T) {
		row := []interface{}{"Name", "Price", "Stock"}

		headers := parseSheetHeaders(row)

		assert.Len(t, headers, 3)
		assert.Equal(t, "Name", headers[0])
		assert.Equal(t, "Price", headers[1])
		assert.Equal(t, "Stock", headers[2])
	})

	t.Run("handles non-string values with empty string", func(t *testing.T) {
		row := []interface{}{"Name", 123, nil}

		headers := parseSheetHeaders(row)

		assert.Len(t, headers, 3)
		assert.Equal(t, "Name", headers[0])
		assert.Equal(t, "", headers[1])
		assert.Equal(t, "", headers[2])
	})

	t.Run("returns empty slice for empty row", func(t *testing.T) {
		row := []interface{}{}

		headers := parseSheetHeaders(row)

		assert.Len(t, headers, 0)
	})
}

func TestConvertRowsToMaps(t *testing.T) {
	t.Run("converts rows to maps using headers", func(t *testing.T) {
		headers := []string{"name", "price", "stock"}
		rows := [][]interface{}{
			{"Product A", 100, 50},
			{"Product B", 200, 30},
		}

		result := convertRowsToMaps(rows, headers)

		assert.Len(t, result, 2)
		assert.Equal(t, "Product A", result[0]["name"])
		assert.Equal(t, 100, result[0]["price"])
		assert.Equal(t, 50, result[0]["stock"])
		assert.Equal(t, "Product B", result[1]["name"])
	})

	t.Run("handles rows shorter than headers", func(t *testing.T) {
		headers := []string{"name", "price", "stock"}
		rows := [][]interface{}{
			{"Product A", 100}, // Missing stock
		}

		result := convertRowsToMaps(rows, headers)

		assert.Len(t, result, 1)
		assert.Equal(t, "Product A", result[0]["name"])
		assert.Equal(t, 100, result[0]["price"])
		assert.Equal(t, "", result[0]["stock"])
	})

	t.Run("skips empty header columns", func(t *testing.T) {
		headers := []string{"name", "", "stock"}
		rows := [][]interface{}{
			{"Product A", "ignored", 50},
		}

		result := convertRowsToMaps(rows, headers)

		assert.Len(t, result, 1)
		assert.Equal(t, "Product A", result[0]["name"])
		assert.Equal(t, 50, result[0]["stock"])
		_, hasEmpty := result[0][""]
		assert.False(t, hasEmpty)
	})

	t.Run("returns empty slice for empty rows", func(t *testing.T) {
		headers := []string{"name", "price"}
		rows := [][]interface{}{}

		result := convertRowsToMaps(rows, headers)

		assert.Len(t, result, 0)
		assert.NotNil(t, result)
	})
}
