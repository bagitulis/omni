package analytics

import (
	"fmt"


// validateMonthYear validates month (1-12) and year (2000-2099).
func validateMonthYear(month, year int) error {
	if month < 1 || month > 12 {
		return fmt.Errorf("invalid month: %d (must be 1-12)", month)
	}
	if year < 2000 || year > 2099 {
		return fmt.Errorf("invalid year: %d (must be 2000-2099)", year)
	}
	return nil
}

// stringPtr returns a pointer to the given string value.
func stringPtr(s string) *string {
	return &s
}
