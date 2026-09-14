package models

import "reflect"

// hasField reports whether v (a struct or pointer to struct) has a field with
// the given name. Test helper used to assert schema-isolation conventions:
// tenant-scoped tables must NOT carry a TenantID column because isolation is
// handled by the PostgreSQL schema (tenant_{id}), matching models.Job.
func hasField(v any, name string) bool {
	t := reflect.TypeOf(v)
	if t == nil {
		return false
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}
	_, ok := t.FieldByName(name)
	return ok
}
