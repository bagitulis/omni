package main

import (
	"fmt"
	"mcp-servers/pkg/sheets"
	"strings"
)

func main() {
	sheetsService := sheets.NewService()
	spreadsheetID := "1ZYonq5Lla0FriY-wfvqW5XEwVyfF6osB1yAtTVABFgA"
	
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Sheet: GH (Main Accounts)")
	fmt.Println(strings.Repeat("=", 60))
	
	values, err := sheetsService.GetValues(spreadsheetID, "GH!A:G")
	if err != nil {
		fmt.Printf("Error reading GH sheet: %v\n", err)
		return
	}
	
	for i, row := range values {
		if i < 10 {
			fmt.Printf("Row %d: %v\n", i+1, row)
		}
	}
	
	fmt.Printf("\nTotal rows in GH: %d\n\n", len(values))
	
	// Check backup codes
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("Sheet: Code_GH (Backup Codes)")
	fmt.Println(strings.Repeat("=", 60))
	
	codes, err := sheetsService.GetValues(spreadsheetID, "Code_GH!A:B")
	if err != nil {
		fmt.Printf("Error reading Code_GH sheet: %v\n", err)
		return
	}
	
	for i, row := range codes {
		if i < 20 {
			fmt.Printf("Row %d: %v\n", i+1, row)
		}
	}
	
	fmt.Printf("\nTotal rows in Code_GH: %d\n", len(codes))
}
