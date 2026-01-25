package utils

import (
	"testing"
)

func TestStringPtr(t *testing.T) {
	input := "test"
	ptr := StringPtr(input)

	if ptr == nil {
		t.Error("StringPtr returned nil")
	}
	if *ptr != input {
		t.Errorf("StringPtr() = %v, want %v", *ptr, input)
	}

	// Verify it's a different address
	if ptr == &input {
		t.Error("StringPtr should return a new pointer, not the same address")
	}
}

func TestInt64Ptr(t *testing.T) {
	input := int64(42)
	ptr := Int64Ptr(input)

	if ptr == nil {
		t.Error("Int64Ptr returned nil")
	}
	if *ptr != input {
		t.Errorf("Int64Ptr() = %v, want %v", *ptr, input)
	}
}

func TestIntPtr(t *testing.T) {
	input := 42
	ptr := IntPtr(input)

	if ptr == nil {
		t.Error("IntPtr returned nil")
	}
	if *ptr != input {
		t.Errorf("IntPtr() = %v, want %v", *ptr, input)
	}
}

func TestFloat64Ptr(t *testing.T) {
	input := 3.14
	ptr := Float64Ptr(input)

	if ptr == nil {
		t.Error("Float64Ptr returned nil")
	}
	if *ptr != input {
		t.Errorf("Float64Ptr() = %v, want %v", *ptr, input)
	}
}

func TestBoolPtr(t *testing.T) {
	tests := []bool{true, false}

	for _, input := range tests {
		ptr := BoolPtr(input)
		if ptr == nil {
			t.Errorf("BoolPtr(%v) returned nil", input)
		}
		if *ptr != input {
			t.Errorf("BoolPtr() = %v, want %v", *ptr, input)
		}
	}
}

// Test that pointers can be dereferenced and modified independently
func TestPointerIndependence(t *testing.T) {
	s1 := StringPtr("hello")
	s2 := StringPtr("hello")

	*s1 = "world"

	if *s2 != "hello" {
		t.Error("Modifying one pointer affected another")
	}
}
