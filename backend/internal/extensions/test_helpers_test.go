package extensions

import "reflect"

// hasField reports whether v (a struct or pointer to struct) has a field with
// the given name.
//
// Used by the isolation tests to assert the *absence* of client-supplied tenant
// fields. Absence is the actual defence: if a TenantID field existed on an
// inbound request, a client could choose which tenant its data lands in. A test
// that asserts "this field does not exist" therefore guards a security property,
// not a style preference.
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
