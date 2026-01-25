package handlers

// extractHeaders extracts keys from a map as headers
func extractHeaders(data map[string]interface{}) []string {
	headers := make([]string, 0, len(data))
	for k := range data {
		headers = append(headers, k)
	}
	return headers
}

// buildSheetValues builds a 2D array for Google Sheets from data
func buildSheetValues(data []map[string]interface{}, headers []string) [][]interface{} {
	values := make([][]interface{}, 0, len(data)+1)

	// Build header row
	headerRow := make([]interface{}, len(headers))
	for i, h := range headers {
		headerRow[i] = h
	}
	values = append(values, headerRow)

	// Build data rows
	for _, item := range data {
		row := make([]interface{}, len(headers))
		for i, h := range headers {
			if v, ok := item[h]; ok {
				row[i] = v
			} else {
				row[i] = ""
			}
		}
		values = append(values, row)
	}

	return values
}

// parseSheetHeaders parses headers from a sheet data row
func parseSheetHeaders(row []interface{}) []string {
	headers := make([]string, len(row))
	for i, h := range row {
		if s, ok := h.(string); ok {
			headers[i] = s
		} else {
			headers[i] = ""
		}
	}
	return headers
}

// convertRowsToMaps converts sheet rows to maps using headers
func convertRowsToMaps(rows [][]interface{}, headers []string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		item := make(map[string]interface{})
		for i, h := range headers {
			if h == "" {
				continue
			}
			if i < len(row) {
				item[h] = row[i]
			} else {
				item[h] = ""
			}
		}
		result = append(result, item)
	}
	return result
}
