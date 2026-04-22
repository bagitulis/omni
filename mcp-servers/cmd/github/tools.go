package main

import "mcp-servers/pkg/mcp"

// buildToolDefinitions returns all MCP tool definitions for the GitHub accounts server
func buildToolDefinitions() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "get_menu",
			Description: "TRIGGER: 'MCP GH' or 'GH Menu' - Shows available GitHub account management tools",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_accounts_by_status",
			Description: "TRIGGER: 'MCP GH status [STATUS]' or 'GH status [STATUS]' - Get accounts filtered by status (no passwords/2FA secrets returned)",
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
			Description: "TRIGGER: 'MCP GH add' or 'GH add' - Add new account. NOTE: Avoid pasting passwords/2FA/backup codes into chat. Prefer 'add_account_from_file' to keep secrets local. RECOGNITION RULES: 1) Code Recovery/2FA Secret = 16 characters, Base32, NO dash (e.g. BUCD6EOGDBBIOAKK). 2) Backup Codes = 11 characters with dash format xxxxx-xxxxx (e.g. 7c197-eb3b9). Always check character count and dash presence to distinguish them.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"email":        {Type: "string", Description: "Email address"},
					"password":     {Type: "string", Description: "Password (optional; prefer file-based flow)"},
					"username":     {Type: "string", Description: "Username (optional)"},
					"codeRecovery": {Type: "string", Description: "2FA/TOTP secret key - 16 chars, Base32, NO dash (optional; prefer file-based flow)"},
					"codes":        {Type: "array", Description: "Backup codes - format xxxxx-xxxxx with dash (optional; prefer file-based flow)", Items: &mcp.Items{Type: "string"}},
				},
				Required: []string{"email"},
			},
		},
		{
			Name:        "add_account_from_file",
			Description: "SAFE: Add new account from a local JSON file path, so secrets don't get pasted into chat. File format: {email, username?, password?, codeRecovery?, codes?: [..]}. You can place the file OUTSIDE the repo and optionally delete it after import.",
			InputSchema: mcp.InputSchema{
				Type: "object",
				Properties: map[string]mcp.Property{
					"path":        {Type: "string", Description: "Path to JSON file containing account payload"},
					"deleteAfter": {Type: "boolean", Description: "If true, delete the JSON file after successful import"},
				},
				Required: []string{"path"},
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
			Description: "TRIGGER: 'MCP GH status all' - Get all accounts (no passwords/2FA secrets/backup codes returned)",
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
	}
}
