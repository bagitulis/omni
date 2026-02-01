package utils

import (
	"errors"
	"strings"
	"unicode"
)

// PasswordPolicy configuration
type PasswordPolicy struct {
	MinLength        int
	MaxLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
}

// DefaultPasswordPolicy returns recommended policy
func DefaultPasswordPolicy() *PasswordPolicy {
	return &PasswordPolicy{
		MinLength:        8,
		MaxLength:        128,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireNumber:    true,
		RequireSpecial:   false, // Optional for usability
	}
}

// ValidatePassword validates password against policy
func (p *PasswordPolicy) ValidatePassword(password string) error {
	if len(password) < p.MinLength {
		return errors.New("password must be at least 8 characters")
	}

	if len(password) > p.MaxLength {
		return errors.New("password must be less than 128 characters")
	}

	var hasUpper, hasLower, hasNumber, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if p.RequireUppercase && !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if p.RequireLowercase && !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if p.RequireNumber && !hasNumber {
		return errors.New("password must contain at least one number")
	}
	if p.RequireSpecial && !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	// Check for common weak passwords
	if isCommonPassword(strings.ToLower(password)) {
		return errors.New("password is too common, please choose a stronger password")
	}

	return nil
}

// Common passwords to reject
var commonPasswords = map[string]bool{
	"password": true, "123456": true, "12345678": true, "123456789": true,
	"qwerty": true, "abc123": true, "password123": true, "password1": true,
	"admin": true, "letmein": true, "welcome": true, "monkey": true,
	"dragon": true, "master": true, "1234567890": true, "login": true,
	"iloveyou": true, "sunshine": true, "princess": true, "football": true,
	"baseball": true, "trustno1": true, "shadow": true, "superman": true,
}

func isCommonPassword(password string) bool {
	return commonPasswords[password]
}

// ValidatePasswordStrength is the main validation function
func ValidatePasswordStrength(password string) error {
	return DefaultPasswordPolicy().ValidatePassword(password)
}
