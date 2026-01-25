package unit_test

import (
	"testing"

	"github.com/omni/backend/internal/utils"
)

func TestJWTGenerateAndValidate(t *testing.T) {
	jwtService := utils.NewJWTService("test-secret-key")

	// Generate token
	token, err := jwtService.GenerateToken("user123", "tenant_test", "admin")
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	if token == "" {
		t.Error("Expected non-empty token")
	}

	// Validate token
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("Failed to validate token: %v", err)
	}

	if claims.UserID != "user123" {
		t.Errorf("Expected userID 'user123', got '%s'", claims.UserID)
	}
	if claims.TenantID != "tenant_test" {
		t.Errorf("Expected tenantID 'tenant_test', got '%s'", claims.TenantID)
	}
	if claims.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", claims.Role)
	}
}

func TestJWTInvalidToken(t *testing.T) {
	jwtService := utils.NewJWTService("test-secret-key")

	// Test invalid token
	_, err := jwtService.ValidateToken("invalid.token.here")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestJWTWrongSecret(t *testing.T) {
	jwtService1 := utils.NewJWTService("secret1")
	jwtService2 := utils.NewJWTService("secret2")

	// Generate with secret1
	token, _ := jwtService1.GenerateToken("user", "tenant", "role")

	// Validate with secret2
	_, err := jwtService2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating with wrong secret")
	}
}
