//go:build tools
// +build tools

package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Generate hash for password123 (used by both yumna and tika)
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	fmt.Println("-- Hash for password123 --")
	fmt.Printf("Hash: %s\n\n", string(hash))

	// SQL to update PostgreSQL
	fmt.Println("-- SQL to update PostgreSQL --")
	fmt.Printf("UPDATE tenant_yumna_bertigamart.users SET password = '%s', failed_login_attempts = 0 WHERE username = 'yumna';\n", string(hash))

	hash2, _ := bcrypt.GenerateFromPassword([]byte("password123"), 10)
	fmt.Printf("UPDATE tenant_tika_nusseyba.users SET password = '%s', failed_login_attempts = 0 WHERE username = 'tika';\n", string(hash2))

	// Verify
	fmt.Println("\n-- Verification --")
	err := bcrypt.CompareHashAndPassword(hash, []byte("password123"))
	fmt.Printf("Verify password123: %v\n", err == nil)
}
