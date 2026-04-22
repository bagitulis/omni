package main

import "mcp-servers/pkg/mcp"

func buildToolDefinitions() []mcp.Tool {
	return []mcp.Tool{
		{
			Name:        "get_menu",
			Description: "TRIGGER: 'MCP AG' or 'AG Menu' - Shows available Antigravity account management tools",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_accounts_by_status",
			Description: "TRIGGER: 'MCP AG status [STATUS]' or 'AG status [STATUS]' - Get accounts filtered by status",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"status": {Type: "string", Description: "Status filter"}},
				Required:   []string{"status"},
			},
		},
		{
			Name:        "delete_backup_code",
			Description: "TRIGGER: 'MCP AG [CODE]' or 'AG [CODE]' - Delete a specific backup code",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"code": {Type: "string", Description: "Backup code to delete"}},
				Required:   []string{"code"},
			},
		},
		{
			Name:        "add_account",
			Description: "TRIGGER: 'MCP AG add' or 'AG add' - Add new account. RECOGNITION RULES: 1) Code Recovery/2FA Secret = 16 characters, Base32, NO dash (e.g. BUCD6EOGDBBIOAKK). 2) Backup Codes = 11 characters with dash format xxxxx-xxxxx (e.g. 7c197-eb3b9). Always check character count and dash presence to distinguish them.",
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
			Description: "TRIGGER: 'MCP AG delete [EMAIL]' - Delete account and all backup codes",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"email": {Type: "string", Description: "Email to delete"}},
				Required:   []string{"email"},
			},
		},
		{
			Name:        "get_summary",
			Description: "TRIGGER: 'MCP AG summary' - Get summary of all accounts",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "get_all_accounts",
			Description: "TRIGGER: 'MCP AG status all' - Get all accounts with status and backup codes",
			InputSchema: mcp.InputSchema{Type: "object", Properties: map[string]mcp.Property{}, Required: []string{}},
		},
		{
			Name:        "generate_2fa",
			Description: "Generate 2FA/TOTP 6-digit code from secret key. Use when user types 'AG 2fa [SECRET]'",
			InputSchema: mcp.InputSchema{
				Type:       "object",
				Properties: map[string]mcp.Property{"secret": {Type: "string", Description: "2FA secret key (base32)"}},
				Required:   []string{"secret"},
			},
		},
	}
}
