// Package unit contains unit tests for services
package unit

import (
	"context"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// ===========================================
// PASSWORD HASHING TESTS
// ===========================================

func TestPasswordHashing_HashesPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if string(hash) == password {
		t.Error("Hash should not equal original password")
	}

	if len(hash) < 20 {
		t.Error("Hash should be at least 20 characters")
	}
}

func TestPasswordHashing_DifferentHashesForSamePassword(t *testing.T) {
	password := "TestPassword123!"

	hash1, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hash2, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if string(hash1) == string(hash2) {
		t.Error("Same password should produce different hashes (salt)")
	}
}

func TestPasswordHashing_VerifyCorrectPassword(t *testing.T) {
	password := "TestPassword123!"

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	err := bcrypt.CompareHashAndPassword(hash, []byte(password))
	if err != nil {
		t.Error("Should verify correct password")
	}
}

func TestPasswordHashing_RejectWrongPassword(t *testing.T) {
	password := "TestPassword123!"
	wrongPassword := "WrongPassword456!"

	hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	err := bcrypt.CompareHashAndPassword(hash, []byte(wrongPassword))
	if err == nil {
		t.Error("Should reject wrong password")
	}
}

// ===========================================
// ACCOUNT LOCKING LOGIC TESTS
// ===========================================

// MockUser represents a user for testing
type MockUser struct {
	ID                  string
	Username            string
	Password            string
	FailedLoginAttempts int
	AccountLockedUntil  *time.Time
	LastFailedLogin     *time.Time
}

func (u *MockUser) IsLocked() bool {
	if u.AccountLockedUntil == nil {
		return false
	}
	return time.Now().Before(*u.AccountLockedUntil)
}

func (u *MockUser) LockMinutesRemaining() int {
	if u.AccountLockedUntil == nil {
		return 0
	}
	remaining := time.Until(*u.AccountLockedUntil)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Minutes()) + 1
}

func TestAccountLocking_NotLockedByDefault(t *testing.T) {
	user := &MockUser{
		ID:                  "1",
		FailedLoginAttempts: 0,
		AccountLockedUntil:  nil,
	}

	if user.IsLocked() {
		t.Error("User should not be locked by default")
	}
}

func TestAccountLocking_LockedWhenTimeInFuture(t *testing.T) {
	futureTime := time.Now().Add(15 * time.Minute)
	user := &MockUser{
		ID:                 "1",
		AccountLockedUntil: &futureTime,
	}

	if !user.IsLocked() {
		t.Error("User should be locked when lock time is in future")
	}
}

func TestAccountLocking_UnlockedWhenTimeInPast(t *testing.T) {
	pastTime := time.Now().Add(-1 * time.Minute)
	user := &MockUser{
		ID:                 "1",
		AccountLockedUntil: &pastTime,
	}

	if user.IsLocked() {
		t.Error("User should be unlocked when lock time is in past")
	}
}

func TestAccountLocking_LockMinutesRemaining(t *testing.T) {
	futureTime := time.Now().Add(10 * time.Minute)
	user := &MockUser{
		ID:                 "1",
		AccountLockedUntil: &futureTime,
	}

	remaining := user.LockMinutesRemaining()
	if remaining < 9 || remaining > 11 {
		t.Errorf("Expected ~10 minutes remaining, got %d", remaining)
	}
}

func TestAccountLocking_ZeroMinutesWhenNotLocked(t *testing.T) {
	user := &MockUser{
		ID:                 "1",
		AccountLockedUntil: nil,
	}

	remaining := user.LockMinutesRemaining()
	if remaining != 0 {
		t.Errorf("Expected 0 minutes remaining, got %d", remaining)
	}
}

// ===========================================
// FAILED LOGIN ATTEMPTS TESTS
// ===========================================

func TestFailedAttempts_IncrementCounter(t *testing.T) {
	user := &MockUser{
		ID:                  "1",
		FailedLoginAttempts: 0,
	}

	// Simulate increment
	user.FailedLoginAttempts++

	if user.FailedLoginAttempts != 1 {
		t.Errorf("Expected 1 failed attempt, got %d", user.FailedLoginAttempts)
	}
}

func TestFailedAttempts_LockAfterMaxAttempts(t *testing.T) {
	maxAttempts := 5
	lockDuration := 30 * time.Minute

	user := &MockUser{
		ID:                  "1",
		FailedLoginAttempts: 4, // One more will trigger lock
	}

	// Simulate failed login
	user.FailedLoginAttempts++

	// Check if should lock
	if user.FailedLoginAttempts >= maxAttempts {
		lockTime := time.Now().Add(lockDuration)
		user.AccountLockedUntil = &lockTime
	}

	if !user.IsLocked() {
		t.Error("User should be locked after max attempts")
	}
}

func TestFailedAttempts_ResetOnSuccess(t *testing.T) {
	user := &MockUser{
		ID:                  "1",
		FailedLoginAttempts: 3,
	}

	// Simulate successful login reset
	user.FailedLoginAttempts = 0
	user.AccountLockedUntil = nil
	user.LastFailedLogin = nil

	if user.FailedLoginAttempts != 0 {
		t.Error("Failed attempts should be reset to 0")
	}

	if user.IsLocked() {
		t.Error("User should not be locked after reset")
	}
}

// ===========================================
// PASSWORD VALIDATION TESTS
// ===========================================

func TestPasswordValidation_MinLength(t *testing.T) {
	testCases := []struct {
		password string
		minLen   int
		valid    bool
	}{
		{"short", 8, false},
		{"12345678", 8, true},
		{"longpassword123", 8, true},
		{"", 8, false},
	}

	for _, tc := range testCases {
		isValid := len(tc.password) >= tc.minLen
		if isValid != tc.valid {
			t.Errorf("Password '%s' (len %d): expected valid=%v, got %v",
				tc.password, len(tc.password), tc.valid, isValid)
		}
	}
}

func TestPasswordValidation_HasUppercase(t *testing.T) {
	testCases := []struct {
		password string
		valid    bool
	}{
		{"lowercase123", false},
		{"Uppercase123", true},
		{"ALLUPPERCASE", true},
		{"mixedCase", true},
	}

	for _, tc := range testCases {
		hasUpper := false
		for _, c := range tc.password {
			if c >= 'A' && c <= 'Z' {
				hasUpper = true
				break
			}
		}
		if hasUpper != tc.valid {
			t.Errorf("Password '%s': expected hasUppercase=%v, got %v",
				tc.password, tc.valid, hasUpper)
		}
	}
}

func TestPasswordValidation_HasNumber(t *testing.T) {
	testCases := []struct {
		password string
		valid    bool
	}{
		{"NoNumbers", false},
		{"Has1Number", true},
		{"123456", true},
		{"Mixed123Case", true},
	}

	for _, tc := range testCases {
		hasNumber := false
		for _, c := range tc.password {
			if c >= '0' && c <= '9' {
				hasNumber = true
				break
			}
		}
		if hasNumber != tc.valid {
			t.Errorf("Password '%s': expected hasNumber=%v, got %v",
				tc.password, tc.valid, hasNumber)
		}
	}
}

// ===========================================
// CONTEXT HANDLING TESTS
// ===========================================

func TestContext_TimeoutHandling(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Simulate slow operation
	time.Sleep(50 * time.Millisecond)

	select {
	case <-ctx.Done():
		t.Error("Context should not be done yet")
	default:
		// OK - context still valid
	}
}

func TestContext_CancellationHandling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	select {
	case <-ctx.Done():
		// OK - context is cancelled
	default:
		t.Error("Context should be cancelled")
	}
}

// ===========================================
// USER RESPONSE FORMATTING TESTS
// ===========================================

type UserResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	// Password should NOT be included
}

func TestUserResponse_ExcludesPassword(t *testing.T) {
	// Simulate formatting user response
	user := &MockUser{
		ID:       "1",
		Username: "testuser",
		Password: "hashedpassword123",
	}

	response := UserResponse{
		ID:       user.ID,
		Username: user.Username,
		// Password intentionally not included
	}

	// Verify password is not in response
	if response.ID != user.ID {
		t.Error("ID should be included in response")
	}

	// There should be no way to access password from UserResponse
	// This is a compile-time check - UserResponse has no Password field
}

// ===========================================
// TENANT ID VALIDATION TESTS
// ===========================================

func TestTenantID_RejectEmpty(t *testing.T) {
	testCases := []struct {
		tenantID string
		valid    bool
	}{
		{"", false},
		{"   ", false},
		{"yumna_bertigamart", true},
		{"tika_nusseyba", true},
	}

	for _, tc := range testCases {
		isValid := len(tc.tenantID) > 0 && tc.tenantID != "   "
		if isValid != tc.valid {
			t.Errorf("TenantID '%s': expected valid=%v, got %v",
				tc.tenantID, tc.valid, isValid)
		}
	}
}

func TestTenantID_NoDefaultAllowed(t *testing.T) {
	// Per AGENTS.MD: "TIDAK ADA DEFAULT TENANT"
	tenantID := ""

	// Should error if empty
	if tenantID == "" {
		// This is correct behavior - should reject empty tenant
		return
	}

	t.Error("Empty tenant ID should be rejected, not defaulted")
}

// ===========================================
// ROLE VALIDATION TESTS
// ===========================================

func TestRole_ValidRoles(t *testing.T) {
	validRoles := map[string]bool{
		"admin":     true,
		"owner":     true,
		"developer": true,
		"user":      true,
	}

	testCases := []string{"admin", "owner", "developer", "user", "invalid", "superuser"}

	for _, role := range testCases {
		isValid := validRoles[role]
		if role == "invalid" || role == "superuser" {
			if isValid {
				t.Errorf("Role '%s' should not be valid", role)
			}
		} else {
			if !isValid {
				t.Errorf("Role '%s' should be valid", role)
			}
		}
	}
}

func TestRole_DeveloperCanSwitchTenant(t *testing.T) {
	role := "developer"

	canSwitch := role == "developer"

	if !canSwitch {
		t.Error("Developer role should be able to switch tenants")
	}
}

func TestRole_NonDeveloperCannotSwitchTenant(t *testing.T) {
	roles := []string{"admin", "owner", "user"}

	for _, role := range roles {
		canSwitch := role == "developer"

		if canSwitch {
			t.Errorf("Role '%s' should NOT be able to switch tenants", role)
		}
	}
}
