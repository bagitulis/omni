package master_product

import (
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
)

func validateFileImportRows(rows []FileImportRow) []FileImportRow {
	validated := make([]FileImportRow, 0, len(rows))
	seenSku := make(map[string]int)

	for _, row := range rows {
		current := row
		current.ItemName = strings.TrimSpace(current.ItemName)
		current.ItemSku = strings.TrimSpace(current.ItemSku)
		current.VariantName = strings.TrimSpace(current.VariantName)
		current.BatchKey = strings.TrimSpace(current.BatchKey)
		current.Description = strings.TrimSpace(current.Description)

		errorsForRow := make([]string, 0)
		if current.ItemName == "" {
			errorsForRow = append(errorsForRow, "item_name is required")
		}
		if current.ItemSku == "" {
			errorsForRow = append(errorsForRow, "item_sku is required")
		}
		if current.Price <= 0 {
			errorsForRow = append(errorsForRow, "price must be greater than 0")
		}
		if current.Stock < 0 {
			errorsForRow = append(errorsForRow, "stock must be greater than or equal to 0")
		}
		if len(current.Description) > models.MasterProductMaxDescriptionLength {
			errorsForRow = append(errorsForRow, fmt.Sprintf("description exceeds %d characters", models.MasterProductMaxDescriptionLength))
		}

		if len(current.ImageUrls) > models.MasterProductMaxImages {
			errorsForRow = append(errorsForRow, fmt.Sprintf("image_urls exceeds max %d URLs", models.MasterProductMaxImages))
		}

		for _, imageURL := range current.ImageUrls {
			if !isHTTPURL(imageURL) {
				errorsForRow = append(errorsForRow, fmt.Sprintf("invalid image url: %s", imageURL))
			}
		}

		if current.ItemSku != "" {
			lookupKey := strings.ToLower(current.ItemSku)
			if firstRow, exists := seenSku[lookupKey]; exists {
				errorsForRow = append(errorsForRow, fmt.Sprintf("duplicate item_sku in file (first seen at row %d)", firstRow))
			} else {
				seenSku[lookupKey] = current.RowNumber
			}
		}

		current.Errors = errorsForRow
		current.Valid = len(errorsForRow) == 0
		validated = append(validated, current)
	}

	return validated
}

func normalizedBatchKey(row FileImportRow) string {
	if strings.TrimSpace(row.BatchKey) != "" {
		return strings.TrimSpace(row.BatchKey)
	}
	if strings.TrimSpace(row.ItemName) != "" {
		return strings.TrimSpace(row.ItemName)
	}
	return fmt.Sprintf("row-%d", row.RowNumber)
}

func formatRowErrors(rowNumber int, errorsForRow []string) []string {
	result := make([]string, 0, len(errorsForRow))
	for _, errMsg := range errorsForRow {
		result = append(result, fmt.Sprintf("row %d: %s", rowNumber, errMsg))
	}
	return result
}
