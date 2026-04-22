// Antigravity Accounts MCP Server
package main

import (
	"fmt"

	"mcp-servers/pkg/mcp"
	"mcp-servers/pkg/sheets"
	"mcp-servers/pkg/totp"
)

const spreadsheetID = "1ZYonq5Lla0FriY-wfvqW5XEwVyfF6osB1yAtTVABFgA"

var (
	sheetMain       = "AG"
	sheetCodeBackup = "Code_AG"
	statusValues    = []string{"AVAILABLE", "CHECK", "YUMNA", "KANTOR", "NO QUOTA", "SUSPENDED"}
)

type account struct {
	RowIndex     int    `json:"rowIndex"`
	Email        string `json:"email"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	CodeRecovery string `json:"codeRecovery"`
	Status       string `json:"status"`
	UserYumna    string `json:"userYumna"`
}

type accountWithBackup struct {
	account
	OneBackupCode string `json:"oneBackupCode"`
}

var sheetsService *sheets.Service

func main() {
	var err error
	sheetsService, err = sheets.NewService(spreadsheetID)
	if err != nil {
		fmt.Printf("Failed to initialize sheets: %v\n", err)
		return
	}

	server := mcp.NewServer("mcp-antigravity-accounts", "1.0.0")
	server.RegisterTools(buildToolDefinitions())

	server.SetHandler(func(name string, args map[string]interface{}) (interface{}, error) {
		switch name {
		case "get_menu":
			return getMenu(), nil
		case "get_accounts_by_status":
			return getAccountsByStatus(args["status"].(string))
		case "delete_backup_code":
			return deleteBackupCode(args["code"].(string))
		case "add_account":
			return addAccount(args)
		case "delete_account":
			return deleteAccount(args["email"].(string))
		case "get_summary":
			return getSummary()
		case "get_all_accounts":
			return getAllAccounts()
		case "generate_2fa":
			return totp.GenerateWithInfo(args["secret"].(string))
		default:
			return nil, fmt.Errorf("unknown tool: %s", name)
		}
	})

	server.Run()
}

func getMenu() map[string]interface{} {
	return map[string]interface{}{
		"title":       "🚀 MCP Antigravity Accounts Menu",
		"description": "Kelola akun Antigravity via Google Sheets",
		"shortcuts":   []string{"MCP AG", "AG"},
		"triggers": []map[string]interface{}{
			{"trigger": "MCP AG / AG Menu", "label": "📋 Menu"},
			{"trigger": "AG status [STATUS]", "label": "🔍 Filter Status", "statusOptions": statusValues},
			{"trigger": "AG status all", "label": "📋 Semua Akun"},
			{"trigger": "AG [KODE]", "label": "🗑️ Hapus Kode Backup"},
			{"trigger": "AG add", "label": "➕ Tambah Akun"},
			{"trigger": "AG delete [EMAIL]", "label": "❌ Hapus Akun"},
			{"trigger": "AG summary", "label": "📊 Ringkasan"},
			{"trigger": "AG 2fa [SECRET]", "label": "🔐 Generate 2FA"},
		},
	}
}

func getAccounts() ([]account, error) {
	values, err := sheetsService.GetValues(sheetMain + "!A:G")
	if err != nil {
		return nil, err
	}

	var accounts []account
	for idx, row := range values {
		if idx < 2 {
			continue
		}
		acc := account{RowIndex: idx + 1}
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

func enrichWithBackupCode(accounts []account) ([]accountWithBackup, error) {
	result := make([]accountWithBackup, len(accounts))
	for i, acc := range accounts {
		result[i] = accountWithBackup{account: acc}
		codes, _ := getBackupCodes(acc.Email)
		if len(codes) > 0 {
			result[i].OneBackupCode = codes[0]
		}
	}
	return result, nil
}
