package utils

import (
	"testing"
)

func TestDefaultPasswordPolicy(t *testing.T) {
	policy := DefaultPasswordPolicy()

	if policy.MinLength != 8 {
		t.Errorf("MinLength = %d, expected 8", policy.MinLength)
	}
	if policy.MaxLength != 128 {
		t.Errorf("MaxLength = %d, expected 128", policy.MaxLength)
	}
	if !policy.RequireUppercase {
		t.Error("RequireUppercase should be true")
	}
	if !policy.RequireLowercase {
		t.Error("RequireLowercase should be true")
	}
	if !policy.RequireNumber {
		t.Error("RequireNumber should be true")
	}
	if policy.RequireSpecial {
		t.Error("RequireSpecial should be false (optional)")
	}
}

func TestValidatePasswordStrength_ValidPasswords(t *testing.T) {
	validPasswords := []string{
		"MySecure123",
		"Str0ngP@ss",
		"ABCabc123",
		"TestPass99",
		"Unique2024Pass",
		"SecureKey789",
	}

	for _, pwd := range validPasswords {
		t.Run(pwd, func(t *testing.T) {
			err := ValidatePasswordStrength(pwd)
			if err != nil {
				t.Errorf("ValidatePasswordStrength(%q) = %v, expected nil", pwd, err)
			}
		})
	}
}

func TestValidatePasswordStrength_TooShort(t *testing.T) {
	shortPasswords := []string{
		"Pass1",
		"Ab1",
		"1234567",
		"",
	}

	for _, pwd := range shortPasswords {
		t.Run(pwd, func(t *testing.T) {
			err := ValidatePasswordStrength(pwd)
			if err == nil {
				t.Errorf("ValidatePasswordStrength(%q) should fail for short password", pwd)
			}
		})
	}
}

func TestValidatePasswordStrength_NoUppercase(t *testing.T) {
	passwords := []string{
		"password123",
		"alllowercase1",
		"nouppercasehere9",
	}

	for _, pwd := range passwords {
		t.Run(pwd, func(t *testing.T) {
			err := ValidatePasswordStrength(pwd)
			if err == nil {
				t.Errorf("ValidatePasswordStrength(%q) should fail for missing uppercase", pwd)
			}
		})
	}
}

func TestValidatePasswordStrength_NoLowercase(t *testing.T) {
	passwords := []string{
		"PASSWORD123",
		"ALLUPPERCASE1",
		"NOLOWERCASEHERE9",
	}

	for _, pwd := range passwords {
		t.Run(pwd, func(t *testing.T) {
			err := ValidatePasswordStrength(pwd)
			if err == nil {
				t.Errorf("ValidatePasswordStrength(%q) should fail for missing lowercase", pwd)
			}
		})
	}
}

func TestValidatePasswordStrength_NoNumber(t *testing.T) {
	passwords := []string{
		"PasswordOnly",
		"NoNumbersHere",
		"JustLettersABC",
	}

	for _, pwd := range passwords {
		t.Run(pwd, func(t *testing.T) {
			err := ValidatePasswordStrength(pwd)
			if err == nil {
				t.Errorf("ValidatePasswordStrength(%q) should fail for missing number", pwd)
			}
		})
	}
}

func TestValidatePasswordStrength_CommonPasswords(t *testing.T) {
	// These should be rejected even if they meet character requirements
	commonPasswords := []string{
		"Password1", // "password" is common - but has uppercase so will pass char check
		"password",  // Will fail lowercase AND common check
		"123456",    // Common and fails other checks
		"qwerty",    // Common and fails uppercase
	}

	// Only lowercase common passwords should be caught by the common password check
	// Others fail other requirements first
	err := ValidatePasswordStrength("password")
	if err == nil {
		t.Error("'password' should be rejected")
	}

	// "Password1" actually passes because "password" check is case-insensitive
	// but "Password1" lowercase is "password1" which is different from "password"
	// Let me check if Password1 passes
	for _, pwd := range commonPasswords {
		err := ValidatePasswordStrength(pwd)
		if err == nil {
			t.Errorf("ValidatePasswordStrength(%q) should fail for common password", pwd)
		}
	}
}

func TestValidatePasswordStrength_TooLong(t *testing.T) {
	// Password over 128 characters
	longPassword := "Aa1" + string(make([]byte, 130)) // Will be > 128 chars
	err := ValidatePasswordStrength(longPassword)
	if err == nil {
		t.Error("Password over 128 characters should fail")
	}
}

func TestPasswordPolicy_CustomPolicy(t *testing.T) {
	// Test with custom policy requiring special characters
	policy := &PasswordPolicy{
		MinLength:        12,
		MaxLength:        64,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSpecial:   true,
	}

	// Should fail without special character
	err := policy.ValidatePassword("Password123")
	if err == nil {
		t.Error("Should require special character")
	}

	// Should pass with special character
	err = policy.ValidatePassword("Password123!")
	if err != nil {
		t.Errorf("Password123! should pass: %v", err)
	}
}

func TestIsCommonPassword(t *testing.T) {
	common := []string{
		"password",
		"123456",
		"12345678",
		"qwerty",
		"abc123",
		"password123",
		"admin",
		"letmein",
	}

	for _, pwd := range common {
		if !isCommonPassword(pwd) {
			t.Errorf("isCommonPassword(%q) = false, expected true", pwd)
		}
	}

	notCommon := []string{
		"myUniquePwd",
		"randomString",
		"xyzabc999",
	}

	for _, pwd := range notCommon {
		if isCommonPassword(pwd) {
			t.Errorf("isCommonPassword(%q) = true, expected false", pwd)
		}
	}
}
