package extensions

import (
	"strconv"
	"strings"
)

// parseUserID converts a JWT subject-style user identifier into the numeric form
// stored on extension rows.
//
// The JWT carries user_id as a string while the column is numeric, so the
// conversion is explicit. A non-numeric identifier returns an error and the
// caller stores NULL: recording a wrong number would silently attribute the
// pairing to another user, whereas an absent owner is merely less informative.
func parseUserID(userID string) (int64, error) {
	trimmed := strings.TrimSpace(userID)
	if trimmed == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(trimmed, 10, 64)
}
