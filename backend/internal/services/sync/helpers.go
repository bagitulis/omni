package sync

import (
	"strconv"
	"strings"
)

// getString gets a string value from a map, supporting nested keys with dot notation
func getString(m map[string]interface{}, key string) string {
	val := getNestedValue(m, key)
	if val == nil {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	// Fallback for non-string types
	return ""
}

// getFloat64 gets a float64 value from a map, supporting nested keys with dot notation
func getFloat64(m map[string]interface{}, key string) float64 {
	val := getNestedValue(m, key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case float64:
		return v
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 0
}

// getInt gets an int value from a map, supporting nested keys with dot notation
func getInt(m map[string]interface{}, key string) int {
	val := getNestedValue(m, key)
	if val == nil {
		return 0
	}
	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, _ := strconv.Atoi(v)
		return i
	}
	return 0
}

// getNestedValue retrieves a value from a nested map using dot notation
func getNestedValue(m map[string]interface{}, key string) interface{} {
	if !strings.Contains(key, ".") {
		return m[key]
	}

	parts := strings.Split(key, ".")
	var current interface{} = m

	for _, part := range parts {
		if currMap, ok := current.(map[string]interface{}); ok {
			current = currMap[part]
		} else {
			return nil
		}
	}

	return current
}
