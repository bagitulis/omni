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

type accountRow struct {
	RowIndex     int
	Email        string
	Username     string
	Password     string
	CodeRecovery string
	Status       string
	UserYumna    string
	BackupCodes  []string
}

func main() {
	if spreadsheetID == "" {
		fmt.Println("Error: AG_SPREADSHEET_ID environment variable is required")
		return
	}

	sheetsService, err := sheets.NewService(spreadsheetID)
	if err != nil {
		fmt.Printf("Failed to initialize sheets: %v\n", err)
		return
	}

	accounts, err := getAccounts(sheetsService)
	if err != nil {
		fmt.Printf("Error getting accounts: %v\n", err)
		return
	}

	// Target emails from env (comma-separated) instead of hardcoded
	targetEmailsStr := os.Getenv("AG_TARGET_EMAILS")
	if targetEmailsStr == "" {
		fmt.Println("Error: AG_TARGET_EMAILS environment variable is required (comma-separated)")
		return
	}
	var targetEmails []string
	for _, e := range strings.Split(targetEmailsStr, ",") {
		e = strings.TrimSpace(e)
		if e != "" {
			targetEmails = append(targetEmails, e)
		}
	}

	fmt.Println("=== Antigravity Account Status ===")

	for _, email := range targetEmails {
		var found *accountRow
		for _, acc := range accounts {
			if strings.EqualFold(acc.Email, email) {
				found = &acc
				break
			}
		}

		if found == nil {
			fmt.Printf("❌ %s - NOT FOUND\n\n", email)
			continue
		}

		fmt.Printf("✓ %s\n", found.Email)
		fmt.Printf("  Username: %s\n", found.Username)
		fmt.Printf("  Password: %s\n", found.Password)
		fmt.Printf("  2FA Secret: %s\n", found.CodeRecovery)
		fmt.Printf("  Status: %s\n", found.Status)
		fmt.Printf("  User: %s\n", found.UserYumna)
		fmt.Printf("  Backup Codes: %d codes\n", len(found.BackupCodes))
		if len(found.BackupCodes) > 0 {
			fmt.Printf("    - First code: %s\n", found.BackupCodes[0])
			if len(found.BackupCodes) > 1 {
				fmt.Printf("    - Last code: %s\n", found.BackupCodes[len(found.BackupCodes)-1])
			}
		}
		fmt.Println()
	}
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
		acc := accountRow{RowIndex: idx + 1}
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

		// Get backup codes
		codes, _ := getBackupCodes(sheetsService, acc.Email)
		acc.BackupCodes = codes

		accounts = append(accounts, acc)
	}
	return accounts, nil
}

func getBackupCodes(sheetsService *sheets.Service, email string) ([]string, error) {
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
