package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type AddAccountRequest struct {
	Email        string   `json:"email"`
	Password     string   `json:"password"`
	CodeRecovery string   `json:"codeRecovery"`
	Codes        []string `json:"codes"`
}

func main() {
	accounts := []AddAccountRequest{
		{
			Email:        "yuoistarlk@gmail.com",
			Password:     "Qwerty123.",
			CodeRecovery: "lgqbxji3hzb5c7bfivkzsg4vea3k5cld",
			Codes: []string{
				"9473-7712", "7289-6841", "7914-9548", "2555-8935", "2438-2271",
				"0659-5199", "4450-0953", "5933-7321", "2837-6601", "6111-9444",
			},
		},
		{
			Email:        "ddanangttiossawardi@gmail.com",
			Password:     "mochicantik9",
			CodeRecovery: "vpkyti52aleb5ah6262mq2ieqemwfcco",
			Codes: []string{
				"7754-0065", "4006-7521", "2694-1892", "0445-0715", "6481-2169",
				"8595-9728", "0348-8157", "9009-1861", "4202-7296", "0423-5002",
			},
		},
	}

	fmt.Println("=== Adding Antigravity Accounts ===\n")

	for i, acc := range accounts {
		fmt.Printf("[%d/%d] Adding: %s\n", i+1, len(accounts), acc.Email)

		argsJSON, _ := json.Marshal(acc)

		// Build MCP JSON-RPC request
		request := map[string]interface{}{
			"jsonrpc": "2.0",
			"id":      i + 1,
			"method":  "tools/call",
			"params": map[string]interface{}{
				"name":      "add_account",
				"arguments": json.RawMessage(argsJSON),
			},
		}

		requestJSON, _ := json.MarshalIndent(request, "", "  ")
		fmt.Printf("Request: %s\n", requestJSON)

		// Call MCP server
		cmd := exec.Command("mcp-servers/bin/mcp-antigravity.exe")
		cmd.Stdin = strings.NewReader(string(requestJSON) + "\n")

		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Error: %v\nOutput: %s\n", err, output)
			continue
		}

		fmt.Printf("Response: %s\n", output)
		fmt.Println(strings.Repeat("-", 60))
	}

	fmt.Println("\n✓ Done!")
}
