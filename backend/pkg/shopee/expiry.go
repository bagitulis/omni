package shopee

import "time"

// IsPartnerKeyExpiring reports whether a Shopee partner key needs attention.
// Returns true when:
//   - `expiresAt` is in the past (already expired — hardest case), OR
//   - `expiresAt` is within `warningDays` days of the current time.
//
// A zero-value `expiresAt` is treated as "not expiring" so legacy rows
// migrated before the new column existed do not spawn phantom banners.
//
// `warningDays <= 0` is treated as "0" — only past dates are flagged.
// This is defensive: a caller with a garbage negative window still gets a
// sensible result rather than a silently-inverted check.
func IsPartnerKeyExpiring(expiresAt time.Time, warningDays int) bool {
	if expiresAt.IsZero() {
		return false
	}
	if warningDays < 0 {
		warningDays = 0
	}
	deadline := time.Now().Add(time.Duration(warningDays) * 24 * time.Hour)
	// Not-After to include the exact-boundary case (expires_at == now + 7d
	// should already be flagged so the seller has today, not zero days, to
	// rotate).
	return !expiresAt.After(deadline)
}
