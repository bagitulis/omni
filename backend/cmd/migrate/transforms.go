package main

import (
	"encoding/json"
	"strings"
	"time"
)

// transformDateTime converts date/time values from SQLite to PostgreSQL format
func transformDateTime(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case string:
		if val == "" {
			return nil
		}
		formats := []string{
			"2006-01-02T15:04:05.000Z",
			"2006-01-02T15:04:05Z",
			"2006-01-02 15:04:05",
			"2006-01-02",
		}
		for _, format := range formats {
			if t, err := time.Parse(format, val); err == nil {
				return t
			}
		}
		return val
	case time.Time:
		return val
	default:
		return v
	}
}

// transformBoolean converts boolean values from SQLite to PostgreSQL format
func transformBoolean(v interface{}) interface{} {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case int64:
		return val == 1
	case int:
		return val == 1
	case bool:
		return val
	case string:
		return val == "1" || strings.ToLower(val) == "true"
	default:
		return false
	}
}

// transformJSON validates and passes through JSON values
func transformJSON(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.(type) {
	case string:
		if val == "" || val == "null" {
			return nil
		}
		var js interface{}
		if err := json.Unmarshal([]byte(val), &js); err != nil {
			return val
		}
		return val
	default:
		return v
	}
}
