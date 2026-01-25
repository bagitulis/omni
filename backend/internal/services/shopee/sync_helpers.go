package shopee

import (
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// buildVariantName creates variant name from tier index and tier variations
func buildVariantName(tierIndex []int, tierVariations []shopeePkg.TierVariation) string {
	if len(tierIndex) == 0 || len(tierVariations) == 0 {
		return ""
	}

	var names []string
	for i, idx := range tierIndex {
		if i < len(tierVariations) && idx < len(tierVariations[i].OptionList) {
			names = append(names, tierVariations[i].OptionList[idx].Option)
		}
	}
	return joinStrings(names, ", ")
}

// joinStrings joins strings with separator
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}
