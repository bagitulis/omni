package inventory

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseHeaders(t *testing.T) {
	row := []interface{}{"SKU", 123, nil, true, 45.6}
	headers := parseHeaders(row)

	assert.Equal(t, []string{"SKU", "123", "", "true", "45.6"}, headers)
}

func TestFindColumnIndex(t *testing.T) {
	headers := []string{" SKU ", "Product Name", "Stock"}

	assert.Equal(t, 0, findColumnIndex(headers, "sku"))
	assert.Equal(t, 1, findColumnIndex(headers, " product name "))
	assert.Equal(t, 2, findColumnIndex(headers, "STOCK"))
	assert.Equal(t, -1, findColumnIndex(headers, "price"))
}

func TestHashStrings(t *testing.T) {
	inputA := []string{"a", "b", "c"}
	inputB := []string{"a", "b", "c"}
	inputC := []string{"c", "b", "a"}

	hashA := hashStrings(inputA)
	hashB := hashStrings(inputB)
	hashC := hashStrings(inputC)

	assert.Equal(t, hashA, hashB)
	assert.Len(t, hashA, 32)
	assert.NotEqual(t, hashA, hashC)
}
