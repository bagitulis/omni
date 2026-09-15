package shopee

// PartnerPair is a Shopee partner_id + partner_key tuple. The zero-value is
// treated as "not configured" — see SelectActivePair for the fallback rules.
type PartnerPair struct {
	ID  int64
	Key string
}

// IsValid reports whether both fields are populated. A pair with only one of
// {ID, Key} set cannot sign an API request and is treated as missing.
func (p PartnerPair) IsValid() bool {
	return p.ID > 0 && p.Key != ""
}

// SelectActivePair picks the pair that should sign the next Shopee API call
// given the tenant's `active_partner_env` toggle. Rules:
//   - env == "test" AND `test` pair is valid → use `test`.
//   - Anything else (env == "live", empty, unknown, or `test` requested but
//     missing) → use `live`.
//
// The safety default is `live` so a mis-configured tenant hits production
// with valid credentials rather than silently sending a bad request. When
// the caller actually wants test but the pair is missing, they'll notice on
// the outbound API 401 or via the UI's "Test Configured: false" badge.
func SelectActivePair(live, test PartnerPair, env string) PartnerPair {
	if env == "test" && test.IsValid() {
		return test
	}
	return live
}
