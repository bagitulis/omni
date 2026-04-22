// GitHub Accounts MCP Server
package main

import (
	"fmt"
	"strings"

	"mcp-servers/pkg/mcp"
	"mcp-servers/pkg/sheets"
	"mcp-servers/pkg/totp"
)

const spreadsheetID = "1ZYonq5Lla0FriY-wfvqW5XEwVyfF6osB1yAtTVABFgA"

var (
	sheetMain       = "GH"
	sheetCodeBackup = "Code_GH"
	statusValues    = []string{"AVAILABLE", "CHECK", "YUMNA", "KANTOR", "NO QUOTA", "SUSPENDED", "BANNED"}

	// statusAliases maps typos/variations to canonical status values
	statusAliases = map[string]string{
		// NO QUOTA variations
		"NOQUOTA": "NO QUOTA", "NO_QUOTA": "NO QUOTA", "NO-QUOTA": "NO QUOTA",
		"NOQUATA": "NO QUOTA", "NO QUATA": "NO QUOTA", "NOQOUTA": "NO QUOTA",
		"NO QOUTA": "NO QUOTA", "HABIS": "NO QUOTA", "LIMIT": "NO QUOTA",
		"LIMITED": "NO QUOTA", "NOQUOTE": "NO QUOTA", "NO QUOTE": "NO QUOTA",
		// AVAILABLE variations
		"AVAIL": "AVAILABLE", "TERSEDIA": "AVAILABLE", "READY": "AVAILABLE",
		"OK": "AVAILABLE", "AKTIF": "AVAILABLE", "ACTIVE": "AVAILABLE",
		"AVL": "AVAILABLE", "AVAIALBLE": "AVAILABLE", "AVIALABLE": "AVAILABLE",
		"AVAILABEL": "AVAILABLE",
		// CHECK variations
		"CEK": "CHECK", "CHECKING": "CHECK", "VERIFY": "CHECK",
		"CHEK": "CHECK", "CECK": "CHECK",
		// SUSPENDED variations
		"SUSPEND": "SUSPENDED", "SUSPENED": "SUSPENDED", "SUSPENDE": "SUSPENDED",
		"BLOKIR": "SUSPENDED", "BLOCKED": "SUSPENDED",
		// BANNED variations
		"BAN": "BANNED", "BANED": "BANNED", "BAND": "BANNED", "BANN": "BANNED",
		// KANTOR variations
		"OFFICE": "KANTOR", "KNTOR": "KANTOR", "KANTRO": "KANTOR",
		// YUMNA variations
		"YUMNA'S": "YUMNA", "YUMNAA": "YUMNA",
	}
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

type accountView struct {
	RowIndex         int    `json:"rowIndex"`
	Email            string `json:"email"`
	Username         string `json:"username"`
	Status           string `json:"status"`
	UserYumna        string `json:"userYumna"`
	PasswordSet      bool   `json:"passwordSet"`
	Has2FASecret     bool   `json:"has2faSecret"`
	BackupCodesCount int    `json:"backupCodesCount"`
}

type addAccountPayload struct {
	Email        string   `json:"email"`
	Username     string   `json:"username"`
	Password     string   `json:"password"`
	CodeRecovery string   `json:"codeRecovery"`
	Codes        []string `json:"codes"`
}

var sheetsService *sheets.Service

func main() {
	var err error
	sheetsService, err = sheets.NewService(spreadsheetID)
	if err != nil {
		fmt.Printf("Failed to initialize sheets: %v\n", err)
		return
	}

	server := mcp.NewServer("mcp-github-accounts", "1.0.0")
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
		case "add_account_from_file":
			return addAccountFromFile(args)
		case "delete_account":
			return deleteAccount(args["email"].(string))
		case "get_summary":
			return getSummary()
		case "get_all_accounts":
			return getAllAccounts()
		case "generate_2fa":
			return totp.GenerateWithInfo(args["secret"].(string))
		case "update_status":
			return updateStatus(args["email"].(string), args["status"].(string))
		default:
			return nil, fmt.Errorf("unknown tool: %s", name)
		}
	})

	server.Run()
}

func getMenu() map[string]interface{} {
	return map[string]interface{}{
		"title":       "🐙 MCP GitHub Accounts Menu",
		"description": "Kelola akun GitHub via Google Sheets",
		"shortcuts":   []string{"MCP GH", "GH"},
		"triggers": []map[string]interface{}{
			{"trigger": "MCP GH / GH Menu", "label": "📋 Menu"},
			{"trigger": "GH status [STATUS]", "label": "🔍 Filter Status", "statusOptions": statusValues},
			{"trigger": "GH status all", "label": "📋 Semua Akun"},
			{"trigger": "GH [KODE]", "label": "🗑️ Hapus Kode Backup"},
			{"trigger": "GH add", "label": "➕ Tambah Akun"},
			{"trigger": "GH delete [EMAIL]", "label": "❌ Hapus Akun"},
			{"trigger": "GH update [EMAIL] [STATUS]", "label": "🔄 Update Status"},
			{"trigger": "GH summary", "label": "📊 Ringkasan"},
			{"trigger": "MCP 2fa [SECRET]", "label": "🔐 Generate 2FA"},
		},
	}
}

// getAccounts fetches all accounts from the main sheet
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

// toAccountViews converts accounts to safe views (no passwords/secrets)
func toAccountViews(accounts []account) ([]accountView, error) {
	result := make([]accountView, 0, len(accounts))
	for _, acc := range accounts {
		codes, _ := getBackupCodes(acc.Email)
		result = append(result, accountView{
			RowIndex:         acc.RowIndex,
			Email:            acc.Email,
			Username:         acc.Username,
			Status:           acc.Status,
			UserYumna:        acc.UserYumna,
			PasswordSet:      strings.TrimSpace(acc.Password) != "",
			Has2FASecret:     strings.TrimSpace(acc.CodeRecovery) != "",
			BackupCodesCount: len(codes),
		})
	}
	return result, nil
}

// normalizeStatus validates and normalizes status input
func normalizeStatus(input string) (string, bool) {
	normalized := strings.ToUpper(strings.TrimSpace(input))

	for _, valid := range statusValues {
		if normalized == valid {
			return valid, true
		}
	}

	if corrected, ok := statusAliases[normalized]; ok {
		return corrected, true
	}

	return "", false
}
