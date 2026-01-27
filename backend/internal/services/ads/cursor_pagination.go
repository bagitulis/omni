package ads

import (
	"encoding/base64"
	"strconv"
)

// CursorPaginationResult holds cursor pagination response
type CursorPaginationResult struct {
	Data       interface{} `json:"data"`
	NextCursor *string     `json:"next_cursor,omitempty"`
	HasMore    bool        `json:"has_more"`
	Total      int64       `json:"total"`
}

// EncodeIntCursor encodes an integer ID to base64 string
func encodeIntCursor(id int) string {
	return base64.StdEncoding.EncodeToString([]byte(strconv.Itoa(id)))
}

// DecodeIntCursor decodes a base64 cursor string to integer ID (exported for handler use)
func DecodeIntCursor(cursor string) (*int, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return nil, err
	}
	id, err := strconv.Atoi(string(decoded))
	if err != nil {
		return nil, err
	}
	return &id, nil
}
