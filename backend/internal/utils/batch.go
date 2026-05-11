package utils

import "fmt"

const (
	DefaultMaxBatchSize = 100
)

// ValidateBatchSize checks that the number of items does not exceed maxSize.
// Returns an error if items > maxSize.
func ValidateBatchSize(items int, maxSize int) error {
	if maxSize <= 0 {
		maxSize = DefaultMaxBatchSize
	}
	if items > maxSize {
		return fmt.Errorf("batch size %d exceeds maximum allowed %d", items, maxSize)
	}
	return nil
}
