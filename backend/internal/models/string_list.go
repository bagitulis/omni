package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

// StringList is a JSONB-backed []string column.
//
// JSONArray exists but is []interface{}, which forces a type assertion at every
// read site. Capabilities (e.g. ["tab_discovery","shopee_scrape"]) are always
// strings, so a typed list keeps call sites honest.
type StringList []string

// Value implements driver.Valuer for database writes.
func (s StringList) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// Scan implements sql.Scanner for database reads.
func (s *StringList) Scan(value any) error {
	if value == nil {
		*s = nil
		return nil
	}

	var data []byte
	switch v := value.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		return errors.New("unsupported type for StringList")
	}

	if len(data) == 0 {
		*s = nil
		return nil
	}

	return json.Unmarshal(data, s)
}

// GormDataType returns the GORM data type for this field.
func (StringList) GormDataType() string {
	return "jsonb"
}

// Has reports whether the list contains the given value.
func (s StringList) Has(v string) bool {
	for _, item := range s {
		if item == v {
			return true
		}
	}
	return false
}
