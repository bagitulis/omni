package utils

import (
	"fmt"
	"strconv"
)

const (
	DefaultPage     = 1
	DefaultPageSize = 20
	MaxPageSize     = 100
	MaxPage         = 10000
)

// ValidatePagination parses and validates pagination parameters.
// Returns sanitized page/pageSize with bounds enforcement.
func ValidatePagination(pageStr, pageSizeStr string) (page, pageSize int, err error) {
	if pageStr == "" {
		page = DefaultPage
	} else {
		page, err = strconv.Atoi(pageStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid page parameter: %s", pageStr)
		}
	}

	if pageSizeStr == "" {
		pageSize = DefaultPageSize
	} else {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			return 0, 0, fmt.Errorf("invalid pageSize parameter: %s", pageSizeStr)
		}
	}

	if page < 1 {
		return 0, 0, fmt.Errorf("page must be >= 1, got %d", page)
	}
	if page > MaxPage {
		return 0, 0, fmt.Errorf("page must be <= %d, got %d", MaxPage, page)
	}
	if pageSize < 1 {
		return 0, 0, fmt.Errorf("pageSize must be >= 1, got %d", pageSize)
	}
	if pageSize > MaxPageSize {
		return 0, 0, fmt.Errorf("pageSize must be <= %d, got %d", MaxPageSize, pageSize)
	}

	return page, pageSize, nil
}
