package inventory

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"
)

// parseHeaders converts interface row to string headers
func parseHeaders(row []interface{}) []string {
	headers := make([]string, len(row))
	for i, v := range row {
		if v != nil {
			headers[i] = fmt.Sprintf("%v", v)
		}
	}
	return headers
}

// findColumnIndex finds the index of a column in headers (case-insensitive)
func findColumnIndex(headers []string, column string) int {
	target := strings.ToLower(strings.TrimSpace(column))
	for i, h := range headers {
		if strings.ToLower(strings.TrimSpace(h)) == target {
			return i
		}
	}
	return -1
}

// hashStrings creates MD5 hash from string slice
func hashStrings(strs []string) string {
	h := md5.New()
	for _, s := range strs {
		h.Write([]byte(s))
	}
	return hex.EncodeToString(h.Sum(nil))
}
