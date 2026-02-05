package handlers

import (
	"testing"
)

// TestOAuthDTOStructure verifies the OAuth DTO package is importable
func TestOAuthDTOStructure(t *testing.T) {
	// Verify LazadaTokenResponse struct exists
	lazadaResp := LazadaTokenResponse{
		AccessToken:      "access123",
		RefreshToken:     "refresh123",
		Country:          "SG",
		RefreshExpiresIn: 2592000,
		ExpiresIn:        3600,
		Account:          "account123",
	}

	if lazadaResp.Country != "SG" {
		t.Errorf("Expected country 'SG', got %s", lazadaResp.Country)
	}

	// Verify TiktokTokenResponse struct exists
	tiktokResp := TiktokTokenResponse{
		Code:    0,
		Message: "success",
		Data: TiktokTokenData{
			AccessToken:  "access456",
			RefreshToken: "refresh456",
			OpenID:       "openid123",
			SellerName:   "TestSeller",
		},
	}

	if tiktokResp.Data.SellerName != "TestSeller" {
		t.Errorf("Expected seller_name 'TestSeller', got %s", tiktokResp.Data.SellerName)
	}

	// Verify ShopeeTokenResponse struct exists
	shopeeResp := ShopeeTokenResponse{
		AccessToken:  "access789",
		RefreshToken: "refresh789",
		ExpireIn:     3600,
		PartnerID:    123456,
		ShopID:       789012,
	}

	if shopeeResp.ShopID != 789012 {
		t.Errorf("Expected shop_id 789012, got %d", shopeeResp.ShopID)
	}
}
