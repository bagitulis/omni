package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// getAllAccounts returns all accounts as safe views
func getAllAccounts() ([]accountView, error) {
	accounts, err := getAccounts()
	if err != nil {
		return nil, err
	}
	return toAccountViews(accounts)
}

// getAccountsByStatus returns accounts filtered by status
func getAccountsByStatus(status string) ([]accountView, error) {
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

	return toAccountViews(filtered)
}

// addAccount adds a new account from tool arguments
func addAccount(args map[string]interface{}) (map[string]interface{}, error) {
	email := strings.TrimSpace(args["email"].(string))
	username := ""
	password := ""
	codeRecovery := ""
	var codes []string

	if u, ok := args["username"].(string); ok {
		username = strings.TrimSpace(u)
	}
	if p, ok := args["password"].(string); ok {
		password = p
	}
	if cr, ok := args["codeRecovery"].(string); ok {
		codeRecovery = cr
	}
	if c, ok := args["codes"].([]interface{}); ok {
		for _, code := range c {
			if s, ok := code.(string); ok {
				trimmed := strings.TrimSpace(s)
				if trimmed != "" {
					codes = append(codes, trimmed)
				}
			}
		}
	}

	if email == "" {
		return map[string]interface{}{"success": false, "message": "Email wajib diisi"}, nil
	}

	return addAccountData(addAccountPayload{
		Email:        email,
		Username:     username,
		Password:     password,
		CodeRecovery: codeRecovery,
		Codes:        codes,
	})
}

// addAccountFromFile adds a new account from a local JSON file
func addAccountFromFile(args map[string]interface{}) (map[string]interface{}, error) {
	pathRaw := strings.TrimSpace(args["path"].(string))
	deleteAfter := false
	if v, ok := args["deleteAfter"].(bool); ok {
		deleteAfter = v
	}

	if pathRaw == "" {
		return map[string]interface{}{"success": false, "message": "Path wajib diisi"}, nil
	}

	path := filepath.Clean(pathRaw)
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("Gagal membaca file: %v", err)}, nil
	}

	var payload addAccountPayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return map[string]interface{}{"success": false, "message": fmt.Sprintf("Format JSON tidak valid: %v", err)}, nil
	}
	payload.Email = strings.TrimSpace(payload.Email)
	payload.Username = strings.TrimSpace(payload.Username)

	if payload.Email == "" {
		return map[string]interface{}{"success": false, "message": "Field 'email' wajib diisi"}, nil
	}

	result, err := addAccountData(payload)
	if err != nil {
		return nil, err
	}

	if deleteAfter {
		_ = os.Remove(path)
		result["fileDeleted"] = true
	}
	result["source"] = "file"
	result["path"] = path
	return result, nil
}

// addAccountData persists a new account to the sheet
func addAccountData(payload addAccountPayload) (map[string]interface{}, error) {
	accounts, _ := getAccounts()
	for _, acc := range accounts {
		if strings.EqualFold(acc.Email, payload.Email) {
			return map[string]interface{}{"success": false, "message": fmt.Sprintf("Email %s sudah ada", payload.Email)}, nil
		}
	}

	row := []interface{}{payload.Email, payload.Username, payload.Password, payload.CodeRecovery, "AVAILABLE", ""}
	if err := sheetsService.AppendRow(sheetMain, row); err != nil {
		return nil, err
	}

	if len(payload.Codes) > 0 {
		_ = addBackupCodes(payload.Email, payload.Codes)
	}

	return map[string]interface{}{
		"success":        true,
		"message":        fmt.Sprintf("Akun %s berhasil ditambahkan", payload.Email),
		"email":          payload.Email,
		"codesAdded":     len(payload.Codes),
		"passwordStored": strings.TrimSpace(payload.Password) != "",
		"has2faSecret":   strings.TrimSpace(payload.CodeRecovery) != "",
	}, nil
}

// deleteAccount deletes an account and its backup codes
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

// getSummary returns account count summary by status
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

// updateStatus updates the status of an account by email
func updateStatus(email, status string) (map[string]interface{}, error) {
	normalizedStatus, valid := normalizeStatus(status)
	if !valid {
		return map[string]interface{}{
			"success":       false,
			"message":       fmt.Sprintf("Status '%s' tidak dikenali", status),
			"validStatuses": statusValues,
			"hint":          "Gunakan salah satu status yang valid di atas",
		}, nil
	}

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
	if strings.ToUpper(oldStatus) == normalizedStatus {
		return map[string]interface{}{
			"success": true,
			"message": fmt.Sprintf("Status sudah '%s', tidak perlu diubah", normalizedStatus),
			"email":   email,
			"status":  normalizedStatus,
		}, nil
	}

	rangeStr := fmt.Sprintf("%s!E%d", sheetMain, found.RowIndex)
	err = sheetsService.UpdateValues(rangeStr, [][]interface{}{{normalizedStatus}})
	if err != nil {
		return nil, fmt.Errorf("gagal update status: %w", err)
	}

	response := map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("Status akun %s berhasil diubah dari '%s' ke '%s'", email, oldStatus, normalizedStatus),
		"email":     email,
		"oldStatus": oldStatus,
		"newStatus": normalizedStatus,
	}

	if strings.ToUpper(strings.TrimSpace(status)) != normalizedStatus {
		response["correctedFrom"] = status
		response["note"] = fmt.Sprintf("Input '%s' telah dinormalisasi menjadi '%s'", status, normalizedStatus)
	}

	return response, nil
}
