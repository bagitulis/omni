package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// JSONMap Tests

func TestJSONMapValue(t *testing.T) {
	tests := []struct {
		name    string
		input   JSONMap
		want    string
		wantErr bool
	}{
		{
			name:    "Nil JSONMap",
			input:   nil,
			want:    "",
			wantErr: false,
		},
		{
			name:    "Empty JSONMap",
			input:   JSONMap{},
			want:    "{}",
			wantErr: false,
		},
		{
			name: "JSONMap with data",
			input: JSONMap{
				"key1": "value1",
				"key2": 123,
				"key3": true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.input.Value()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.input == nil {
				assert.Nil(t, val)
				return
			}

			// Verify it's valid JSON
			data, ok := val.([]byte)
			require.True(t, ok, "Value should return []byte")

			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			assert.NoError(t, err)
		})
	}
}

func TestJSONMapScan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		expect  JSONMap
		wantErr bool
	}{
		{
			name:    "Nil value",
			input:   nil,
			expect:  nil,
			wantErr: false,
		},
		{
			name:    "Bytes input",
			input:   []byte(`{"key":"value"}`),
			expect:  JSONMap{"key": "value"},
			wantErr: false,
		},
		{
			name:    "String input",
			input:   `{"name":"test"}`,
			expect:  JSONMap{"name": "test"},
			wantErr: false,
		},
		{
			name:    "Empty bytes",
			input:   []byte{},
			expect:  nil,
			wantErr: false,
		},
		{
			name:    "Invalid type",
			input:   123,
			wantErr: true,
		},
		{
			name:    "Invalid JSON",
			input:   []byte(`{invalid}`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result JSONMap
			err := result.Scan(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expect, result)
		})
	}
}

func TestJSONMapGormDataType(t *testing.T) {
	jm := JSONMap{}
	assert.Equal(t, "jsonb", jm.GormDataType())
}

func TestJSONMapRoundTrip(t *testing.T) {
	// Note: JSON unmarshaling converts all numbers to float64
	original := JSONMap{
		"string": "value",
		"number": float64(42),
		"float":  3.14,
		"bool":   true,
		"null":   nil,
		"nested": map[string]interface{}{"inner": "value"},
		"array":  []interface{}{float64(1), float64(2), float64(3)},
	}

	// Encode to database value
	dbValue, err := original.Value()
	require.NoError(t, err)

	// Decode from database value
	var decoded JSONMap
	err = decoded.Scan(dbValue)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, original, decoded)
}

// JSONArray Tests

func TestJSONArrayValue(t *testing.T) {
	tests := []struct {
		name    string
		input   JSONArray
		want    string
		wantErr bool
	}{
		{
			name:    "Nil JSONArray",
			input:   nil,
			want:    "",
			wantErr: false,
		},
		{
			name:    "Empty JSONArray",
			input:   JSONArray{},
			want:    "[]",
			wantErr: false,
		},
		{
			name: "JSONArray with data",
			input: JSONArray{
				"string",
				123,
				3.14,
				true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := tt.input.Value()

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			if tt.input == nil {
				assert.Nil(t, val)
				return
			}

			// Verify it's valid JSON
			data, ok := val.([]byte)
			require.True(t, ok, "Value should return []byte")

			var result []interface{}
			err = json.Unmarshal(data, &result)
			assert.NoError(t, err)
		})
	}
}

func TestJSONArrayScan(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		check   func(JSONArray) bool
		wantErr bool
	}{
		{
			name:    "Nil value",
			input:   nil,
			check:   func(j JSONArray) bool { return j == nil },
			wantErr: false,
		},
		{
			name:  "Bytes input",
			input: []byte(`["a","b","c"]`),
			check: func(j JSONArray) bool {
				return len(j) == 3 && j[0] == "a"
			},
			wantErr: false,
		},
		{
			name:  "String input",
			input: `[1,2,3]`,
			check: func(j JSONArray) bool {
				return len(j) == 3
			},
			wantErr: false,
		},
		{
			name:    "Empty bytes",
			input:   []byte{},
			check:   func(j JSONArray) bool { return j == nil },
			wantErr: false,
		},
		{
			name:    "Invalid type",
			input:   123,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result JSONArray
			err := result.Scan(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.True(t, tt.check(result))
		})
	}
}

func TestJSONArrayGormDataType(t *testing.T) {
	ja := JSONArray{}
	assert.Equal(t, "jsonb", ja.GormDataType())
}

func TestJSONArrayRoundTrip(t *testing.T) {
	// Note: JSON unmarshaling converts all numbers to float64
	original := JSONArray{
		"string",
		float64(42),
		3.14,
		true,
		nil,
		[]interface{}{float64(1), float64(2)},
		map[string]interface{}{"key": "value"},
	}

	// Encode to database value
	dbValue, err := original.Value()
	require.NoError(t, err)

	// Decode from database value
	var decoded JSONArray
	err = decoded.Scan(dbValue)
	require.NoError(t, err)

	// Verify round-trip
	assert.Equal(t, original, decoded)
}
