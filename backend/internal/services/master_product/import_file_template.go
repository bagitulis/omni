package master_product

import (
	"bytes"
	"encoding/csv"
	"github.com/xuri/excelize/v2"
)

func buildXlsxTemplate() ([]byte, string, string, error) {
	f := excelize.NewFile()
	defer func() {
		_ = f.Close()
	}()

	templateSheet := "products_template"
	instructionSheet := "instructions"
	index, err := f.NewSheet(templateSheet)
	if err != nil {
		return nil, "", "", err
	}
	f.SetActiveSheet(index)

	headers := []string{"item_name", "item_sku", "variant_name", "price", "stock", "batch_key", "description", "image_urls"}
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		if setErr := f.SetCellValue(templateSheet, cell, header); setErr != nil {
			return nil, "", "", setErr
		}
	}

	sampleRows := [][]string{
		{"Running Shoes", "RUN-RED-42", "Red / 42", "299000", "25", "RUN-SHOES-RED", "Lightweight running shoes", "https://cdn.example.com/shoes/red-front.jpg|https://cdn.example.com/shoes/red-side.jpg"},
		{"Running Shoes", "RUN-RED-43", "Red / 43", "299000", "18", "RUN-SHOES-RED", "", ""},
	}
	for rowIndex, row := range sampleRows {
		for colIndex, value := range row {
			cell, _ := excelize.CoordinatesToCellName(colIndex+1, rowIndex+2)
			if setErr := f.SetCellValue(templateSheet, cell, value); setErr != nil {
				return nil, "", "", setErr
			}
		}
	}

	if _, err := f.NewSheet(instructionSheet); err != nil {
		return nil, "", "", err
	}

	instructions := []string{
		"Product Import Guide",
		"",
		"How to use this template:",
		"1. Fill in the 'products_template' sheet with your product data.",
		"2. One row = one SKU/variant. Keep each seller SKU unique.",
		"3. batch_key groups rows into one master product. Same batch_key = same product.",
		"4. Delete the example rows (rows 2-3) before importing your own data.",
		"",
		"Column descriptions:",
		"  item_name      — Product display name (required)",
		"  item_sku       — Unique seller SKU identifier, e.g. BAKIW5971 (required)",
		"  variant_name   — Variant label, e.g. Red / 42 (optional)",
		"  price          — Product price, numbers only, must be > 0 (required)",
		"  stock          — Available stock quantity, must be >= 0 (required)",
		"  batch_key      — Groups rows into one product; same key = same product (optional, defaults to item_name)",
		"  description    — Product description, shared at product level inside a batch (optional)",
		"  image_urls     — URL(s) to product images separated by | (optional, max 8, must be valid https:// URLs)",
		"",
		"Notes:",
		"  - SKU is case-insensitive (BAKIW5971 = bakiw5971).",
		"  - Do not modify the column headers in row 1.",
		"  - Preferred format is .xlsx. CSV is also supported.",
		"  - image_urls: only URLs are supported (no file uploads). Images will be downloaded and cached automatically.",
		"  - price must be > 0 and stock must be >= 0.",
		"  - description max length is 5000 characters.",
	}

	for rowIndex, value := range instructions {
		cell, _ := excelize.CoordinatesToCellName(1, rowIndex+1)
		if setErr := f.SetCellValue(instructionSheet, cell, value); setErr != nil {
			return nil, "", "", setErr
		}
	}

	_ = f.SetColWidth(templateSheet, "A", "H", 24)
	_ = f.SetColWidth(instructionSheet, "A", "A", 120)

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, "", "", err
	}

	return buf.Bytes(), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", "products_import_template.xlsx", nil
}

func buildCSVTemplate() ([]byte, string, string, error) {
	buf := &bytes.Buffer{}
	writer := csv.NewWriter(buf)
	headers := []string{"item_name", "item_sku", "variant_name", "price", "stock", "batch_key", "description", "image_urls"}
	if err := writer.Write(headers); err != nil {
		return nil, "", "", err
	}
	if err := writer.Write([]string{"Running Shoes", "RUN-RED-42", "Red / 42", "299000", "25", "RUN-SHOES-RED", "Lightweight running shoes", "https://cdn.example.com/shoes/red-front.jpg|https://cdn.example.com/shoes/red-side.jpg"}); err != nil {
		return nil, "", "", err
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		return nil, "", "", err
	}

	return buf.Bytes(), "text/csv", "products_import_template.csv", nil
}
