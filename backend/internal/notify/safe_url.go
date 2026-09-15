package notify

import (
	"errors"
	"regexp"
	"strings"
)

// SPA-relative path allowlist. Anything that does not match is rejected.
// Design decision (see docs/superpowers/specs/2026-09-16-notifications-v2-design.md):
// only paths relative to the SPA root are allowed; absolute URLs, protocol-
// relative URLs, and non-http schemes are blocked at write time to eliminate
// open-redirect and injection surface.
var safePathRe = regexp.MustCompile(`^/[A-Za-z0-9\-_./?=&%#]*$`)

// MaxActionURLLength is the ceiling for action_url (matches the DB column
// size). Producers that need something longer must persist a resource ID and
// resolve the URL client-side.
const MaxActionURLLength = 500

// ErrInvalidActionURL is returned by SafeActionURL when the input is not a
// SPA-relative path.
var ErrInvalidActionURL = errors.New("invalid action_url")

// SafeActionURL validates an action URL. Empty input is allowed (treated as
// "no action"). Any non-empty value MUST be a path starting with "/" and
// composed of allowlisted characters, under MaxActionURLLength.
func SafeActionURL(raw string) (string, error) {
	if raw == "" {
		return "", nil
	}
	if strings.TrimSpace(raw) == "" {
		return "", ErrInvalidActionURL
	}
	if len(raw) > MaxActionURLLength {
		return "", ErrInvalidActionURL
	}
	if !safePathRe.MatchString(raw) {
		return "", ErrInvalidActionURL
	}
	// Reject protocol-relative "//..." which technically starts with "/" but
	// browsers resolve to a different origin.
	if strings.HasPrefix(raw, "//") {
		return "", ErrInvalidActionURL
	}
	return raw, nil
}
