package services

import (
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// configEntry tracks a single config_key row with its timestamp for duplicate resolution.
type configEntry struct {
	value     string
	updatedAt time.Time
}

// buildConfigMap creates a key-value map from platform config rows, resolving
// duplicate keys using a precedence policy:
//   - Duplicate keys with one empty and one non-empty → use non-empty
//   - Duplicate keys with conflicting non-empty values → log warning, use latest updated_at
func buildConfigMap(rows []inventoryRowData) map[string]string {
	keyEntries := make(map[string][]configEntry)

	for _, row := range rows {
		key := strings.TrimSpace(asString(row["config_key"]))
		value := asString(row["config_value"])
		if key == "" {
			continue
		}
		var ts time.Time
		if v := row["updated_at"]; v != nil {
			switch t := v.(type) {
			case time.Time:
				ts = t
			case string:
				ts, _ = time.Parse(time.RFC3339, t)
			}
		}
		keyEntries[key] = append(keyEntries[key], configEntry{value: value, updatedAt: ts})
	}

	m := make(map[string]string, len(keyEntries))
	for key, entries := range keyEntries {
		if len(entries) == 1 {
			m[key] = entries[0].value
			continue
		}
		m[key] = resolveDuplicateKey(key, entries)
	}
	return m
}

// resolveDuplicateKey applies the duplicate key precedence policy.
// Returns the winning value for a config_key with multiple rows.
func resolveDuplicateKey(key string, entries []configEntry) string {
	// Find indices of all non-empty values
	nonEmpty := make([]int, 0, len(entries))
	for i, e := range entries {
		if e.value != "" {
			nonEmpty = append(nonEmpty, i)
		}
	}

	// All empty → return empty
	if len(nonEmpty) == 0 {
		return ""
	}

	// Exactly one non-empty → use it (regardless of timestamp)
	if len(nonEmpty) == 1 {
		return entries[nonEmpty[0]].value
	}

	// Multiple non-empty values → find latest updated_at among non-empty
	bestIdx := nonEmpty[0]
	for _, idx := range nonEmpty[1:] {
		if entries[idx].updatedAt.After(entries[bestIdx].updatedAt) {
			bestIdx = idx
		}
	}

	// Log warning about conflicting values
	log.Warn().
		Str("config_key", key).
		Int("duplicate_count", len(entries)).
		Time("selected_updated_at", entries[bestIdx].updatedAt).
		Msg("Duplicate config_key resolved — latest updated_at wins")

	return entries[bestIdx].value
}
