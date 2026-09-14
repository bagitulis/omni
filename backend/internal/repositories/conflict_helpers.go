package repositories

import "gorm.io/gorm/clause"

// clauseOnConflictDoNothing returns the ON CONFLICT DO NOTHING clause used for
// idempotent bulk inserts.
//
// Extracted so the "skip duplicates instead of failing" decision is declared in
// one place: resume after a paused scrape deliberately re-reads the boundary
// page, so overlapping rows are expected rather than an error.
func clauseOnConflictDoNothing() clause.Expression {
	return clause.OnConflict{DoNothing: true}
}
