package main

import (
	"fmt"
	"strings"
)

func getAllAccounts() ([]accountWithBackup, error) {
	accounts, err := getAccounts()
	if err != nil {
		return nil, err
	}
	return enrichWithBackupCode(accounts)
}

func getAccountsByStatus(status string) ([]accountWithBackup, error) {
	status = strings.ToUpper(status)
	accounts, err := getAccounts()
	if err != nil {
		return nil, err
	}

	var filtered []account
	for _, acc := range accounts {
		if strings.ToUpper(acc.Status) == status {
			filtered = append(filtered, acc)
		}
	}

	return enrichWithBackupCode(filtered)
}

func addAccount(args map[string]interface{}) (map[string]interface{}, error) {
	email := args["email"].(string)
	password := args["password"].(string)
	username := ""
	codeRecovery := ""
	var codes []string

	if u, ok := args["username"].(string); ok {
		username = u
	}
	if cr, ok := args["codeRecovery"].(string); ok {
		codeRecovery = cr
	}
	if c, ok := args["codes"].([]interface{}); ok {
		for _, code := range c {
			if s, ok := code.(string); ok {
				codes = append(codes, s)
			}
		}
	}

	accounts, _ := getAccounts()
	for _, acc := range accounts {
		if strings.EqualFold(acc.Email, email) {
			return map[string]interface{}{"success": false, "message": fmt.Sprintf("Email %s sudah ada", email)}, nil
		}
	}

	row := []interface{}{email, username, password, codeRecovery, "AVAILABLE", ""}
	if err := sheetsService.AppendRow(sheetMain, row); err != nil {
		return nil, err
	}

	if len(codes) > 0 {
		addBackupCodes(email, codes)
	}

	return map[string]interface{}{
		"success":    true,
		"message":    fmt.Sprintf("Akun %s berhasil ditambahkan", email),
		"email":      email,
		"codesAdded": len(codes),
	}, nil
}

func deleteAccount(email string) (map[string]interface{}, error) {
	accounts, _ := getAccounts()
	var found *account
	for _, acc := range accounts {
		if strings.EqualFold(acc.Email, email) {
			found = &acc
			break
		}
	}
	if found == nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("Akun %s tidak ditemukan", email)}, nil
	}

	deleteBackupColumn(email)
	sheetsService.DeleteRow(sheetMain, int64(found.RowIndex-1))

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Akun %s beserta semua code backup berhasil dihapus", email),
		"email":   email,
	}, nil
}

func getSummary() (map[string]interface{}, error) {
	accounts, err := getAccounts()
	if err != nil {
		return nil, err
	}

	statusCount := make(map[string]int)
	for _, acc := range accounts {
		status := acc.Status
		if status == "" {
			status = "UNKNOWN"
		}
		statusCount[status]++
	}

	return map[string]interface{}{
		"totalAccounts": len(accounts),
		"byStatus":      statusCount,
	}, nil
}
