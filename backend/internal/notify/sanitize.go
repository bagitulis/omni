package notify

import "regexp"

// Port of frontend/src/lib/notificationSecurity.ts to Go — defense in depth
// at write time so leaks cannot reach any client (WS, HTTP list, mobile).

var blockedPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)\bsql\b|query failed|database|relation|column "\w+"`),
	regexp.MustCompile(`(?i)^SELECT\b|\bFROM\b|\bWHERE\b`),
	regexp.MustCompile(`(?i)stack|traceback|at\s+\w+\.\w+\s*\(`),
	regexp.MustCompile(`(?i)/internal/|/pkg/|/src/`),
	regexp.MustCompile(`(?i)token[=:]\s*\S|jwt[=:]\s*\S|bearer\s+\S|apikey|secret[=:]`),
	regexp.MustCompile(`(?i)(?:password|passwd|pwd)[=:]?\s*\S`),
	regexp.MustCompile(`(?i)127\.0\.0\.1:\d{4,5}|localhost:\d{4,5}`),
	regexp.MustCompile(`(?i)panic:|goroutine\s+\d`),
	regexp.MustCompile(`(?i)ECONNREFUSED|ECONNRESET|EPIPE`),
}

// SafeFallback is what SanitizeForUser returns when the input is blocked
// or empty.
const SafeFallback = "An error occurred. Please try again or contact support."

// SanitizeForUser returns raw if none of the blocked patterns match, otherwise
// SafeFallback. Strings longer than 500 chars are truncated.
func SanitizeForUser(raw string) string {
	if raw == "" {
		return SafeFallback
	}
	for _, re := range blockedPatterns {
		if re.MatchString(raw) {
			return SafeFallback
		}
	}
	if len(raw) > 500 {
		return raw[:497] + "..."
	}
	return raw
}
