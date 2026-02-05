package handlers

import (
	"testing"
)

// TestAuthDTOStructure verifies the auth DTO package is importable
func TestAuthDTOStructure(t *testing.T) {
	// Verify LoginRequest struct exists
	loginReq := LoginRequest{
		Username: "testuser",
		Password: "password123",
	}

	if loginReq.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got %s", loginReq.Username)
	}

	// Verify RefreshTokenRequest struct exists
	refreshReq := RefreshTokenRequest{
		RefreshToken: "token123",
	}

	if refreshReq.RefreshToken != "token123" {
		t.Errorf("Expected refresh_token 'token123', got %s", refreshReq.RefreshToken)
	}

	// Verify ChangePasswordRequest struct exists
	changeReq := ChangePasswordRequest{
		OldPassword: "oldpass",
		NewPassword: "newpass123",
	}

	if changeReq.OldPassword != "oldpass" {
		t.Errorf("Expected old_password 'oldpass', got %s", changeReq.OldPassword)
	}

	// Verify RegisterRequest struct exists
	regReq := RegisterRequest{
		Username: "newuser",
		Email:    "test@example.com",
		Password: "password123",
		ShopName: "My Shop",
	}

	if regReq.Email != "test@example.com" {
		t.Errorf("Expected email 'test@example.com', got %s", regReq.Email)
	}

	// Verify SwitchTenantRequest struct exists
	switchReq := SwitchTenantRequest{
		TenantID: "tenant123",
	}

	if switchReq.TenantID != "tenant123" {
		t.Errorf("Expected tenant_id 'tenant123', got %s", switchReq.TenantID)
	}
}
