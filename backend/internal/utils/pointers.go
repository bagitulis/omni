package utils

// Pointer helper functions for converting values to pointers
// Used throughout the codebase when APIs require pointer types

// StringPtr returns a pointer to the given string value
func StringPtr(s string) *string {
	return &s
}

// Int64Ptr returns a pointer to the given int64 value
func Int64Ptr(i int64) *int64 {
	return &i
}

// IntPtr returns a pointer to the given int value
func IntPtr(i int) *int {
	return &i
}

// Float64Ptr returns a pointer to the given float64 value
func Float64Ptr(f float64) *float64 {
	return &f
}

// BoolPtr returns a pointer to the given bool value
func BoolPtr(b bool) *bool {
	return &b
}
