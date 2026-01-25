package inventory

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
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

// findColumnIndex finds the index of a column in headers
func findColumnIndex(headers []string, column string) int {
	for i, h := range headers {
		if h == column {
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
