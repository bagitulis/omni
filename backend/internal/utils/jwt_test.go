package utils

import (
	"strings"
	"testing"
	"time"
)

func TestNewJWTService(t *testing.T) {
	secret := "test-secret-key"
	svc := NewJWTService(secret)

	if svc == nil {
		t.Error("NewJWTService returned nil")
	}
	if string(svc.secretKey) != secret {
		t.Error("JWTService secretKey not set correctly")
	}
}

func TestJWTService_GenerateToken(t *testing.T) {
	svc := NewJWTService("this-is-a-very-long-secret-key-for-testing")

	token, err := svc.GenerateToken("user123", "tenant_test", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("GenerateToken returned empty token")
	}

	// JWT tokens have 3 parts separated by dots
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Errorf("JWT token should have 3 parts, got %d", len(parts))
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	svc := NewJWTService("this-is-a-very-long-secret-key-for-testing")

	// Generate a token
	token, err := svc.GenerateToken("user123", "tenant_test", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Validate it
	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("claims.UserID = %v, want 'user123'", claims.UserID)
	}
	if claims.TenantID != "tenant_test" {
		t.Errorf("claims.TenantID = %v, want 'tenant_test'", claims.TenantID)
	}
	if claims.Role != "admin" {
		t.Errorf("claims.Role = %v, want 'admin'", claims.Role)
	}
}

func TestJWTService_ValidateToken_WithBearerPrefix(t *testing.T) {
	svc := NewJWTService("this-is-a-very-long-secret-key-for-testing")

	token, err := svc.GenerateToken("user123", "tenant_test", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Validate with Bearer prefix
	claims, err := svc.ValidateToken("Bearer " + token)
	if err != nil {
		t.Fatalf("ValidateToken with Bearer prefix failed: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("claims.UserID = %v, want 'user123'", claims.UserID)
	}
}

func TestJWTService_ValidateToken_InvalidToken(t *testing.T) {
	svc := NewJWTService("this-is-a-very-long-secret-key-for-testing")

	tests := []struct {
		name  string
		token string
	}{
		{"empty token", ""},
		{"invalid format", "not-a-jwt"},
		{"invalid signature", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VySWQiOiIxMjMifQ.invalidsignature"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ValidateToken(tt.token)
			if err == nil {
				t.Error("expected error for invalid token, got nil")
			}
		})
	}
}

func TestJWTService_ValidateToken_WrongSecret(t *testing.T) {
	svc1 := NewJWTService("secret-key-one-for-testing-purposes")
	svc2 := NewJWTService("secret-key-two-for-testing-purposes")

	// Generate with one service
	token, err := svc1.GenerateToken("user123", "tenant_test", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Try to validate with different secret
	_, err = svc2.ValidateToken(token)
	if err == nil {
		t.Error("expected error when validating with wrong secret")
	}
}

func TestJWTService_TokenExpiration(t *testing.T) {
	svc := NewJWTService("this-is-a-very-long-secret-key-for-testing")

	token, err := svc.GenerateToken("user123", "tenant_test", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Token should expire in ~24 hours
	if claims.ExpiresAt == nil {
		t.Error("ExpiresAt should not be nil")
		return
	}

	expiry := claims.ExpiresAt.Time
	now := time.Now()
	diff := expiry.Sub(now)

	// Should be between 23 and 25 hours from now
	if diff < 23*time.Hour || diff > 25*time.Hour {
		t.Errorf("Token expiry diff = %v, expected ~24 hours", diff)
	}
}

func TestJWTClaims(t *testing.T) {
	claims := JWTClaims{
		UserID:   "user123",
		TenantID: "tenant_test",
		Role:     "admin",
	}

	if claims.UserID != "user123" {
		t.Errorf("claims.UserID = %v, want 'user123'", claims.UserID)
	}
	if claims.TenantID != "tenant_test" {
		t.Errorf("claims.TenantID = %v, want 'tenant_test'", claims.TenantID)
	}
	if claims.Role != "admin" {
		t.Errorf("claims.Role = %v, want 'admin'", claims.Role)
	}
}
