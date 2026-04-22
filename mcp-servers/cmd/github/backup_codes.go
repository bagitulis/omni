package main

import (
	"fmt"
	"strings"

	"mcp-servers/pkg/sheets"
)

// getBackupCodes retrieves backup codes for an email from the backup sheet
func getBackupCodes(email string) ([]string, error) {
	values, err := sheetsService.GetValues(sheetCodeBackup + "!A:Z")
	if err != nil || len(values) == 0 {
		return nil, err
	}

	colIndex := -1
	for i, e := range values[0] {
		if strings.EqualFold(e, email) {
			colIndex = i
			break
		}
	}
	if colIndex == -1 {
		return nil, nil
	}

	var codes []string
	for i := 1; i < len(values); i++ {
		if colIndex < len(values[i]) && strings.TrimSpace(values[i][colIndex]) != "" {
			codes = append(codes, values[i][colIndex])
		}
	}
	return codes, nil
}

// deleteBackupCode deletes a specific backup code from the sheet
func deleteBackupCode(code string) (map[string]interface{}, error) {
	values, err := sheetsService.GetValues(sheetCodeBackup + "!A:Z")
	if err != nil || len(values) == 0 {
		return map[string]interface{}{"success": false, "message": "Code not found"}, nil
	}

	for row := 1; row < len(values); row++ {
		for col := 0; col < len(values[row]); col++ {
			if values[row][col] == code {
				email := ""
				if col < len(values[0]) {
					email = values[0][col]
				}
				rangeStr := fmt.Sprintf("%s!%s%d", sheetCodeBackup, sheets.ColToLetter(col), row+1)
				sheetsService.ClearCell(rangeStr)
				return map[string]interface{}{
					"success":     true,
					"message":     fmt.Sprintf("Kode '%s' berhasil dihapus dari %s", code, email),
					"email":       email,
					"deletedCode": code,
				}, nil
			}
		}
	}
	return map[string]interface{}{"success": false, "message": fmt.Sprintf("Kode '%s' tidak ditemukan", code)}, nil
}

// addBackupCodes adds backup codes for an email to the backup sheet
func addBackupCodes(email string, codes []string) error {
	values, _ := sheetsService.GetValues(sheetCodeBackup + "!A:Z")
	var headerRow []string
	if len(values) > 0 {
		headerRow = values[0]
	}

	colIndex := -1
	for i, e := range headerRow {
		if strings.EqualFold(e, email) {
			colIndex = i
			break
		}
	}

	colLetter := ""
	if colIndex == -1 {
		colIndex = len(headerRow)
		colLetter = sheets.ColToLetter(colIndex)
		sheetsService.UpdateValues(fmt.Sprintf("%s!%s1", sheetCodeBackup, colLetter), [][]interface{}{{email}})
	} else {
		colLetter = sheets.ColToLetter(colIndex)
	}

	codeValues := make([][]interface{}, len(codes))
	for i, c := range codes {
		codeValues[i] = []interface{}{c}
	}
	rangeStr := fmt.Sprintf("%s!%s2:%s%d", sheetCodeBackup, colLetter, colLetter, 1+len(codes))
	return sheetsService.UpdateValues(rangeStr, codeValues)
}

// deleteBackupColumn deletes the entire backup code column for an email
func deleteBackupColumn(email string) {
	values, _ := sheetsService.GetValues(sheetCodeBackup + "!A:Z")
	if len(values) == 0 {
		return
	}
	for i, e := range values[0] {
		if strings.EqualFold(e, email) {
			sheetsService.DeleteColumn(sheetCodeBackup, int64(i))
			return
		}
	}
}
