package main

import (
	"fmt"
	"strings"

	"mcp-servers/pkg/sheets"
)

const spreadsheetID = "1ZYonq5Lla0FriY-wfvqW5XEwVyfF6osB1yAtTVABFgA"

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
	accounts := []account{
		{
			Email:        "felixstarlk576@gmail.com",
			Password:     "PASSword12",
			CodeRecovery: "dbxz2cqmwqv3ntvnhrheq2sbvi6f4yf4",
			Codes: []string{
				"0805-0848", "0722-7662", "8513-6883", "5612-3979", "1796-7937",
				"8629-7886", "0357-0474", "5903-8418", "5913-2061", "9220-3418",
			},
		},
	}

	sheetsService, err := sheets.NewService(spreadsheetID)
	if err != nil {
		fmt.Printf("Failed to initialize sheets: %v\n", err)
		return
	}

	fmt.Println("=== Adding Antigravity Accounts ===\n")

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
