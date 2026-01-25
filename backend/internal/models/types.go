package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// JSONMap is a custom type for handling JSONB columns in PostgreSQL
// It properly serializes/deserializes JSON data for GORM
type JSONMap map[string]interface{}

// Value implements the driver.Valuer interface for database writes
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface for database reads
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported type for JSONMap")
	}

	if len(data) == 0 {
		*j = nil
		return nil
	}

	return json.Unmarshal(data, j)
}

// GormDataType returns the GORM data type for this field
func (JSONMap) GormDataType() string {
	return "jsonb"
}

// JSONArray is a custom type for handling JSONB array columns
type JSONArray []interface{}

// Value implements the driver.Valuer interface
func (j JSONArray) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan implements the sql.Scanner interface
func (j *JSONArray) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported type for JSONArray")
	}

	if len(data) == 0 {
		*j = nil
		return nil
	}

	return json.Unmarshal(data, j)
}

// GormDataType returns the GORM data type for this field
func (JSONArray) GormDataType() string {
	return "jsonb"
}
