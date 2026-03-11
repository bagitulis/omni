package google

import (
	"fmt"
	"regexp"
	"sort"
)

// extractSpreadsheetID extracts spreadsheet ID from URL or returns as-is
func extractSpreadsheetID(input string) string {
	if !regexp.MustCompile(`/`).MatchString(input) {
		return input
	}

	re := regexp.MustCompile(`/spreadsheets/d/([a-zA-Z0-9-_]+)`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 1 {
		return matches[1]
	}

	return input
}

// reconstructURL converts spreadsheet ID to full Google Sheets URL
func reconstructURL(spreadsheetID string) string {
	if spreadsheetID == "" {
		return ""
	}
	if regexp.MustCompile(`^https?://`).MatchString(spreadsheetID) {
		return spreadsheetID
	}
	return fmt.Sprintf("https://docs.google.com/spreadsheets/d/%s/edit", spreadsheetID)
}

// orderedMap returns a map with keys sorted alphabetically (for deterministic DB updates)
func orderedMap(input map[string]interface{}) map[string]interface{} {
	keys := make([]string, 0, len(input))
	for key := range input {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ordered := make(map[string]interface{}, len(input))
	for _, key := range keys {
		ordered[key] = input[key]
	}
	return ordered
}
