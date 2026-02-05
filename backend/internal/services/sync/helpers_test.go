package sync

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "simple string",
			m:        map[string]interface{}{"name": "test"},
			key:      "name",
			expected: "test",
		},
		{
			name:     "missing key",
			m:        map[string]interface{}{"name": "test"},
			key:      "missing",
			expected: "",
		},
		{
			name:     "nil map value",
			m:        map[string]interface{}{"name": nil},
			key:      "name",
			expected: "",
		},
		{
			name:     "non-string value",
			m:        map[string]interface{}{"count": 123},
			key:      "count",
			expected: "",
		},
		{
			name: "nested key",
			m: map[string]interface{}{
				"user": map[string]interface{}{
					"name": "nested_value",
				},
			},
			key:      "user.name",
			expected: "nested_value",
		},
		{
			name: "deeply nested key",
			m: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": map[string]interface{}{
						"value": "deep",
					},
				},
			},
			key:      "level1.level2.value",
			expected: "deep",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getString(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetFloat64(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected float64
	}{
		{
			name:     "float64 value",
			m:        map[string]interface{}{"price": 99.99},
			key:      "price",
			expected: 99.99,
		},
		{
			name:     "int value",
			m:        map[string]interface{}{"count": 42},
			key:      "count",
			expected: 42.0,
		},
		{
			name:     "int64 value",
			m:        map[string]interface{}{"big": int64(9999999999)},
			key:      "big",
			expected: 9999999999.0,
		},
		{
			name:     "string value",
			m:        map[string]interface{}{"amount": "123.45"},
			key:      "amount",
			expected: 123.45,
		},
		{
			name:     "missing key",
			m:        map[string]interface{}{},
			key:      "missing",
			expected: 0,
		},
		{
			name:     "invalid string",
			m:        map[string]interface{}{"bad": "not_a_number"},
			key:      "bad",
			expected: 0,
		},
		{
			name: "nested float",
			m: map[string]interface{}{
				"order": map[string]interface{}{
					"total": 199.99,
				},
			},
			key:      "order.total",
			expected: 199.99,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFloat64(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected int
	}{
		{
			name:     "int value",
			m:        map[string]interface{}{"count": 42},
			key:      "count",
			expected: 42,
		},
		{
			name:     "int64 value",
			m:        map[string]interface{}{"big": int64(123)},
			key:      "big",
			expected: 123,
		},
		{
			name:     "float64 value truncates",
			m:        map[string]interface{}{"price": 99.99},
			key:      "price",
			expected: 99,
		},
		{
			name:     "string value",
			m:        map[string]interface{}{"qty": "10"},
			key:      "qty",
			expected: 10,
		},
		{
			name:     "missing key",
			m:        map[string]interface{}{},
			key:      "missing",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInt(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetInt64(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected int64
	}{
		{
			name:     "int64 value",
			m:        map[string]interface{}{"timestamp": int64(1704067200)},
			key:      "timestamp",
			expected: 1704067200,
		},
		{
			name:     "int value",
			m:        map[string]interface{}{"small": 100},
			key:      "small",
			expected: 100,
		},
		{
			name:     "float64 value",
			m:        map[string]interface{}{"decimal": 123.0},
			key:      "decimal",
			expected: 123,
		},
		{
			name:     "string value",
			m:        map[string]interface{}{"str": "9999999999"},
			key:      "str",
			expected: 9999999999,
		},
		{
			name:     "missing key",
			m:        map[string]interface{}{},
			key:      "missing",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInt64(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetNestedValue(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected interface{}
	}{
		{
			name:     "simple key",
			m:        map[string]interface{}{"key": "value"},
			key:      "key",
			expected: "value",
		},
		{
			name: "one level nested",
			m: map[string]interface{}{
				"parent": map[string]interface{}{
					"child": "nested",
				},
			},
			key:      "parent.child",
			expected: "nested",
		},
		{
			name: "two levels nested",
			m: map[string]interface{}{
				"a": map[string]interface{}{
					"b": map[string]interface{}{
						"c": 123,
					},
				},
			},
			key:      "a.b.c",
			expected: 123,
		},
		{
			name: "missing nested key",
			m: map[string]interface{}{
				"a": map[string]interface{}{
					"b": "value",
				},
			},
			key:      "a.c",
			expected: nil,
		},
		{
			name: "non-map intermediate",
			m: map[string]interface{}{
				"a": "string_value",
			},
			key:      "a.b",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNestedValue(tt.m, tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}
