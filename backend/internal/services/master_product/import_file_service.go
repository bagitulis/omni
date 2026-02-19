package master_product

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

const (
	maxImportTemplateRows = 5000
)

// FileImportRow represents one file-import row.
type FileImportRow struct {
	RowNumber   int      `json:"row_number"`
	ItemName    string   `json:"item_name"`
	ItemSku     string   `json:"item_sku"`
	VariantName string   `json:"variant_name,omitempty"`
	Price       float64  `json:"price"`
	Stock       int      `json:"stock"`
	BatchKey    string   `json:"batch_key,omitempty"`
	Description string   `json:"description,omitempty"`
	ImageUrls   []string `json:"image_urls,omitempty"`
	Valid       bool     `json:"valid"`
	Errors      []string `json:"errors,omitempty"`
}

// FileImportPreviewResult represents preview response for upload import.
type FileImportPreviewResult struct {
	TotalRows   int             `json:"total_rows"`
	ValidRows   int             `json:"valid_rows"`
	InvalidRows int             `json:"invalid_rows"`
	Rows        []FileImportRow `json:"rows"`
}

// FileImportResult represents import execution result for uploaded rows.
type FileImportResult struct {
	Imported        int      `json:"imported"`
	ProductsCreated int      `json:"products_created"`
	SkusCreated     int      `json:"skus_created"`
	RowsSkipped     int      `json:"rows_skipped"`
	Errors          []string `json:"errors"`
}

// BuildImportTemplate returns an import template file.
func (s *ImportService) BuildImportTemplate(format string) ([]byte, string, string, error) {
	switch strings.ToLower(format) {
	case "", "xlsx":
		return buildXlsxTemplate()
	case "csv":
		return buildCSVTemplate()
	default:
		return nil, "", "", fmt.Errorf("unsupported template format: %s", format)
	}
}

// PreviewFromFile parses and validates an upload file.
func (s *ImportService) PreviewFromFile(_ context.Context, filename string, file io.Reader) (*FileImportPreviewResult, error) {
	rows, err := parseImportRows(filename, file)
	if err != nil {
		return nil, err
	}

	validated := validateFileImportRows(rows)

	result := &FileImportPreviewResult{Rows: validated}
	result.TotalRows = len(validated)
	for _, row := range validated {
		if row.Valid {
			result.ValidRows++
			continue
		}
		result.InvalidRows++
	}

	return result, nil
}

// ImportFromRows imports validated rows into master products/SKUs.
func (s *ImportService) ImportFromRows(ctx context.Context, tenantID string, rows []FileImportRow) (*FileImportResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}
	if len(rows) == 0 {
		return nil, errors.New("rows are required")
	}

	validated := validateFileImportRows(rows)
	result := &FileImportResult{Errors: make([]string, 0)}

	type importGroup struct {
		title       string
		description string
		imageURLs   []string
		rows        []FileImportRow
	}

	groups := make(map[string]*importGroup)
	for _, row := range validated {
		if !row.Valid {
			result.RowsSkipped++
			result.Errors = append(result.Errors, formatRowErrors(row.RowNumber, row.Errors)...)
			continue
		}

		if _, err := s.repo.FindBySku(ctx, tenantID, row.ItemSku); err == nil {
			result.RowsSkipped++
			result.Errors = append(result.Errors, fmt.Sprintf("row %d: seller_sku already exists (%s)", row.RowNumber, row.ItemSku))
			continue
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to check existing SKU %s: %w", row.ItemSku, err)
		}

		groupKey := normalizedBatchKey(row)
		group, exists := groups[groupKey]
		if !exists {
			groups[groupKey] = &importGroup{
				title:       row.ItemName,
				description: row.Description,
				imageURLs:   row.ImageUrls,
				rows:        []FileImportRow{row},
			}
			continue
		}

		if group.description == "" && row.Description != "" {
			group.description = row.Description
		}
		if len(group.imageURLs) == 0 && len(row.ImageUrls) > 0 {
			group.imageURLs = row.ImageUrls
		}
		group.rows = append(group.rows, row)
	}

	for _, group := range groups {
		if len(group.rows) == 0 {
			continue
		}

		masterProduct := &models.MasterProduct{
			TenantID:    tenantID,
			Title:       truncateString(group.title, models.MasterProductMaxTitleLength),
			Description: truncateString(group.description, models.MasterProductMaxDescriptionLength),
			Status:      models.MasterProductStatusActive,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		images := make(models.JSONArray, 0, len(group.imageURLs))
		for _, imageURL := range group.imageURLs {
			images = append(images, imageURL)
		}
		masterProduct.Images = images

		if err := s.repo.Create(ctx, masterProduct); err != nil {
			return nil, fmt.Errorf("failed to create master product for batch %s: %w", normalizedBatchKey(group.rows[0]), err)
		}

		skusCreatedInGroup := 0
		for _, row := range group.rows {
			sku := &models.MasterProductSku{
				TenantID:        tenantID,
				MasterProductID: masterProduct.ID,
				SellerSku:       row.ItemSku,
				VariantName:     truncateString(row.VariantName, 255),
				VariantData:     make(models.JSONMap),
				Price:           row.Price,
				Stock:           row.Stock,
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			}

			if err := s.repo.CreateSku(ctx, sku); err != nil {
				result.RowsSkipped++
				result.Errors = append(result.Errors, fmt.Sprintf("row %d: failed to create SKU (%s): %v", row.RowNumber, row.ItemSku, err))
				continue
			}

			skusCreatedInGroup++
			result.SkusCreated++
		}

		if skusCreatedInGroup == 0 {
			_ = s.repo.Delete(ctx, masterProduct.ID)
			result.Errors = append(result.Errors, fmt.Sprintf("batch %s: skipped because all SKUs failed", normalizedBatchKey(group.rows[0])))
			continue
		}

		result.ProductsCreated++
	}

	result.Imported = result.ProductsCreated
	return result, nil
}
