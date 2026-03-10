package main

import (
	"fmt"
	"os"
	"strings"

	"mcp-servers/pkg/sheets"
)

var spreadsheetID = os.Getenv("AG_SPREADSHEET_ID")

var (
	sheetMain       = "AG"
	sheetCodeBackup = "Code_AG"
)

type account struct {
	Email        string
	Password     string
	CodeRecovery string
	Codes        []string
}

func main() {
	if spreadsheetID == "" {
		fmt.Println("Error: AG_SPREADSHEET_ID environment variable is required")
		return
	}

	// Accounts must be provided via environment/config — not hardcoded in source
	email := os.Getenv("AG_ACCOUNT_EMAIL")
	password := os.Getenv("AG_ACCOUNT_PASSWORD")
	codeRecovery := os.Getenv("AG_ACCOUNT_RECOVERY")
	codesStr := os.Getenv("AG_ACCOUNT_CODES") // comma-separated

	if email == "" || password == "" {
		fmt.Println("Error: AG_ACCOUNT_EMAIL and AG_ACCOUNT_PASSWORD are required")
		return
	}

	var codes []string
	if codesStr != "" {
		for _, c := range strings.Split(codesStr, ",") {
			c = strings.TrimSpace(c)
			if c != "" {
				codes = append(codes, c)
			}
		}
	}

	accounts := []account{
		{
			Email:        email,
			Password:     password,
			CodeRecovery: codeRecovery,
			Codes:        codes,
		},
	}

	sheetsService, err := sheets.NewService(spreadsheetID)
	if err != nil {
		fmt.Printf("Failed to initialize sheets: %v\n", err)
		return
	}

	fmt.Println("=== Adding Antigravity Accounts ===")

	for i, acc := range accounts {
		fmt.Printf("[%d/%d] Adding: %s\n", i+1, len(accounts), acc.Email)

		// Check if account exists
		existing, _ := getAccounts(sheetsService)
		exists := false
		for _, e := range existing {
			if strings.EqualFold(e.Email, acc.Email) {
				exists = true
				break
			}
		}

		if exists {
			fmt.Printf("   ⚠️  Account already exists, skipping\n\n")
			continue
		}

		// Add account row
		row := []interface{}{acc.Email, "", acc.Password, acc.CodeRecovery, "AVAILABLE", ""}
		if err := sheetsService.AppendRow(sheetMain, row); err != nil {
			fmt.Printf("   ❌ Error adding account: %v\n\n", err)
			continue
		}

		// Add backup codes
		if len(acc.Codes) > 0 {
			if err := addBackupCodes(sheetsService, acc.Email, acc.Codes); err != nil {
				fmt.Printf("   ⚠️  Account added but error adding codes: %v\n\n", err)
			} else {
				fmt.Printf("   ✓ Account added with %d backup codes\n\n", len(acc.Codes))
			}
		} else {
			fmt.Printf("   ✓ Account added (no backup codes)\n\n")
		}
	}

	fmt.Println("=== Done ===")
}

type accountRow struct {
	Email        string
	Username     string
	Password     string
	CodeRecovery string
	Status       string
	UserYumna    string
}

func getAccounts(sheetsService *sheets.Service) ([]accountRow, error) {
	values, err := sheetsService.GetValues(sheetMain + "!A:G")
	if err != nil {
		return nil, err
	}

	var accounts []accountRow
	for idx, row := range values {
		if idx < 2 {
			continue
		}
		acc := accountRow{}
		if len(row) > 0 {
			acc.Email = row[0]
		}
		if len(row) > 1 {
			acc.Username = row[1]
		}
		if len(row) > 2 {
			acc.Password = row[2]
		}
		if len(row) > 3 {
			acc.CodeRecovery = row[3]
		}
		if len(row) > 4 {
			acc.Status = row[4]
		}
		if len(row) > 5 {
			acc.UserYumna = row[5]
		}
		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func addBackupCodes(sheetsService *sheets.Service, email string, codes []string) error {
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
