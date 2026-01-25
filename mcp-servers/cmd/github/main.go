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
		"NOQUOTA":  "NO QUOTA",
		"NO_QUOTA": "NO QUOTA",
		"NO-QUOTA": "NO QUOTA",
		"NOQUATA":  "NO QUOTA",
		"NO QUATA": "NO QUOTA",
		"NOQOUTA":  "NO QUOTA",
		"NO QOUTA": "NO QUOTA",
		"HABIS":    "NO QUOTA",
		"LIMIT":    "NO QUOTA",
		"LIMITED":  "NO QUOTA",
		"NOQUOTE":  "NO QUOTA",
		"NO QUOTE": "NO QUOTA",
		// AVAILABLE variations
		"AVAIL":     "AVAILABLE",
		"TERSEDIA":  "AVAILABLE",
		"READY":     "AVAILABLE",
		"OK":        "AVAILABLE",
		"AKTIF":     "AVAILABLE",
		"ACTIVE":    "AVAILABLE",
		"AVL":       "AVAILABLE",
		"AVAIALBLE": "AVAILABLE",
		"AVIALABLE": "AVAILABLE",
		"AVAILABEL": "AVAILABLE",
		// CHECK variations
		"CEK":      "CHECK",
		"CHECKING": "CHECK",
		"VERIFY":   "CHECK",
		"CHEK":     "CHECK",
		"CECK":     "CHECK",
		// SUSPENDED variations
		"SUSPEND":  "SUSPENDED",
		"SUSPENED": "SUSPENDED",
		"SUSPENDE": "SUSPENDED",
		"BLOKIR":   "SUSPENDED",
		"BLOCKED":  "SUSPENDED",
		// BANNED variations
		"BAN":   "BANNED",
		"BANED": "BANNED",
		"BAND":  "BANNED",
		"BANN":  "BANNED",
		// KANTOR variations
		"OFFICE": "KANTOR",
		"KNTOR":  "KANTOR",
		"KANTRO": "KANTOR",
		// YUMNA variations
		"YUMNA'S": "YUMNA",
		"YUMNAA":  "YUMNA",
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

	server := mcp.NewServer("mcp-github-accounts", "1.0.0")

	server.RegisterTools([]mcp.Tool{
		{
			Name:        "get_menu",
			Description: "TRIGGER: 'MCP GH' or 'GH Menu' - Shows available GitHub account management tools",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_accounts_by_status",
			Description: "TRIGGER: 'MCP GH status [STATUS]' or 'GH status [STATUS]' - Get accounts filtered by status",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"status": {Type: "string", Description: "Status filter"}},
				Required:   []string{"status"},
			},
		},
		{
			Name:        "delete_backup_code",
			Description: "TRIGGER: 'MCP GH [CODE]' or 'GH [CODE]' - Delete a specific backup code",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"code": {Type: "string", Description: "Backup code to delete"}},
				Required:   []string{"code"},
			},
		},
		{
			Name:        "add_account",
			Description: "TRIGGER: 'MCP GH add' or 'GH add' - Add new account. RECOGNITION RULES: 1) Code Recovery/2FA Secret = 16 characters, Base32, NO dash (e.g. BUCD6EOGDBBIOAKK). 2) Backup Codes = 11 characters with dash format xxxxx-xxxxx (e.g. 7c197-eb3b9). Always check character count and dash presence to distinguish them.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"email":        {Type: "string", Description: "Email address"},
					"password":     {Type: "string", Description: "Password"},
					"username":     {Type: "string", Description: "Username (optional)"},
					"codeRecovery": {Type: "string", Description: "2FA/TOTP secret key - 16 chars, Base32, NO dash (e.g. BUCD6EOGDBBIOAKK)"},
					"codes":        {Type: "array", Description: "Backup codes - format xxxxx-xxxxx with dash (e.g. 7c197-eb3b9)", Items: &mcp.Items{Type: "string"}},
				},
				Required: []string{"email", "password", "codes"},
			},
		},
		{
			Name:        "delete_account",
			Description: "TRIGGER: 'MCP GH delete [EMAIL]' - Delete account and all backup codes",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"email": {Type: "string", Description: "Email to delete"}},
				Required:   []string{"email"},
			},
		},
		{
			Name:        "get_summary",
			Description: "TRIGGER: 'MCP GH summary' - Get summary of all accounts",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_all_accounts",
			Description: "TRIGGER: 'MCP GH status all' - Get all accounts with status and backup codes",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "generate_2fa",
			Description: "Generate 2FA/TOTP 6-digit code from secret key. ALWAYS USE THIS TOOL when user types 'MCP 2fa', '2fa', 'totp', 'otp' followed by a secret code. Example: 'MCP 2fa WKX2OSMHXAWA2VRR' should call this with secret='WKX2OSMHXAWA2VRR'",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"secret": {Type: "string", Description: "2FA secret key (base32 encoded, e.g. WKX2OSMHXAWA2VRR)"}},
				Required:   []string{"secret"},
			},
		},
		{
			Name:        "update_status",
			Description: "TRIGGER: 'MCP GH update [EMAIL] [STATUS]' or 'GH update [EMAIL] [STATUS]' - Update account status. Valid statuses: AVAILABLE, CHECK, YUMNA, KANTOR, NO QUOTA, SUSPENDED, BANNED. Typos will be auto-corrected.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"email":  {Type: "string", Description: "Email of the account to update"},
					"status": {Type: "string", Description: "New status (AVAILABLE, CHECK, YUMNA, KANTOR, NO QUOTA, SUSPENDED, BANNED)"},
				},
				Required: []string{"email", "status"},
			},
		},
	})

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

	// Check if exists
	accounts, _ := getAccounts()
	for _, acc := range accounts {
		if strings.EqualFold(acc.Email, email) {
			return map[string]interface{}{"success": false, "message": fmt.Sprintf("Email %s sudah ada", email)}, nil
		}
	}

	// Add account row
	row := []interface{}{email, username, password, codeRecovery, "AVAILABLE", ""}
	if err := sheetsService.AppendRow(sheetMain, row); err != nil {
		return nil, err
	}

	// Add backup codes
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

	// Delete backup codes column
	deleteBackupColumn(email)

	// Delete account row
	sheetsService.DeleteRow(sheetMain, int64(found.RowIndex-1))

	return map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Akun %s beserta semua code backup berhasil dihapus", email),
		"email":   email,
	}, nil
}

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

// normalizeStatus validates and normalizes status input
// Returns normalized status and whether it's valid
func normalizeStatus(input string) (string, bool) {
	// Uppercase and trim
	normalized := strings.ToUpper(strings.TrimSpace(input))

	// Check if it's already a valid status
	for _, valid := range statusValues {
		if normalized == valid {
			return valid, true
		}
	}

	// Check aliases for typo correction
	if corrected, ok := statusAliases[normalized]; ok {
		return corrected, true
	}

	// Not recognized
	return "", false
}

// updateStatus updates the status of an account by email
func updateStatus(email, status string) (map[string]interface{}, error) {
	// Normalize and validate status
	normalizedStatus, valid := normalizeStatus(status)
	if !valid {
		return map[string]interface{}{
			"success":       false,
			"message":       fmt.Sprintf("Status '%s' tidak dikenali", status),
			"validStatuses": statusValues,
			"hint":          "Gunakan salah satu status yang valid di atas",
		}, nil
	}

	// Find account
	accounts, err := getAccounts()
	if err != nil {
		return nil, err
	}

	var found *account
	for _, acc := range accounts {
		if strings.EqualFold(acc.Email, email) {
			found = &acc
			break
		}
	}

	if found == nil {
		return map[string]interface{}{
			"success": false,
			"message": fmt.Sprintf("Akun dengan email '%s' tidak ditemukan", email),
		}, nil
	}

	oldStatus := found.Status

	// Check if status is the same
	if strings.ToUpper(oldStatus) == normalizedStatus {
		return map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Status sudah '%s', tidak perlu diubah", normalizedStatus),
			"email":   email,
			"status":  normalizedStatus,
		}, nil
	}

	// Update status in sheet (column E = index 4, so column letter is E)
	rangeStr := fmt.Sprintf("%s!E%d", sheetMain, found.RowIndex)
	err = sheetsService.UpdateValues(rangeStr, [][]interface{}{{normalizedStatus}})
	if err != nil {
		return nil, fmt.Errorf("gagal update status: %w", err)
	}

	// Build response
	response := map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("Status akun %s berhasil diubah dari '%s' ke '%s'", email, oldStatus, normalizedStatus),
		"email":     email,
		"oldStatus": oldStatus,
		"newStatus": normalizedStatus,
	}

	// Add note if typo was corrected
	if strings.ToUpper(strings.TrimSpace(status)) != normalizedStatus {
		response["correctedFrom"] = status
		response["note"] = fmt.Sprintf("Input '%s' telah dinormalisasi menjadi '%s'", status, normalizedStatus)
	}

	return response, nil
}
