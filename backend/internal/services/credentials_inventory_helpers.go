package services

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

func aggregateInventoryRows(rows []InventoryRow) []InventoryRow {
	byKey := make(map[string]*InventoryRow)
	for _, row := range rows {
		key := strings.Join([]string{row.TenantID, row.Platform, row.RowShape}, "|")
		existing := byKey[key]
		if existing == nil {
			copyRow := row
			copyRow.MissingRequiredFields = append([]string{}, row.MissingRequiredFields...)
			byKey[key] = &copyRow
			continue
		}
		existing.Count += row.Count
		if row.DuplicateStoreCount > existing.DuplicateStoreCount {
			existing.DuplicateStoreCount = row.DuplicateStoreCount
		}
		existing.TargetBackfillCount += row.TargetBackfillCount
		existing.NeedsBackfill = existing.NeedsBackfill || row.NeedsBackfill
		existing.HasSensitiveValuesSeen = existing.HasSensitiveValuesSeen || row.HasSensitiveValuesSeen
		existing.MissingRequiredFields = mergeStrings(existing.MissingRequiredFields, row.MissingRequiredFields)
		if existing.BackfillTarget == "" {
			existing.BackfillTarget = row.BackfillTarget
		}
	}

	result := make([]InventoryRow, 0, len(byKey))
	for _, row := range byKey {
		result = append(result, *row)
	}
	sort.Slice(result, func(i, j int) bool {
		left := strings.Join([]string{result[i].TenantID, result[i].Platform, result[i].RowShape}, "|")
		right := strings.Join([]string{result[j].TenantID, result[j].Platform, result[j].RowShape}, "|")
		return left < right
	})
	return result
}

func mergeStrings(left, right []string) []string {
	seen := make(map[string]bool, len(left)+len(right))
	merged := make([]string, 0, len(left)+len(right))
	for _, item := range append(left, right...) {
		if !seen[item] {
			seen[item] = true
			merged = append(merged, item)
		}
	}
	sort.Strings(merged)
	return merged
}

func asString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case sql.NullString:
		if v.Valid {
			return v.String
		}
		return ""
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprintf("%v", v)
	}
}

func presentAndNonEmpty(value any) bool {
	if value == nil {
		return false
	}
	return strings.TrimSpace(asString(value)) != ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
