package wholesale_test

import (
	"context"
	"strings"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/wholesale"
)

// ---------------------------------------------------------------------------
// WholesaleService constructor
// ---------------------------------------------------------------------------

func TestNewWholesaleService_NotNil(t *testing.T) {
	svc := wholesale.NewWholesaleService(nil, "tenant-1")
	if svc == nil {
		t.Fatal("NewWholesaleService returned nil")
	}
}

func TestNewWholesaleService_DifferentTenants(t *testing.T) {
	svc1 := wholesale.NewWholesaleService(nil, "tenant-a")
	svc2 := wholesale.NewWholesaleService(nil, "tenant-b")
	if svc1 == nil || svc2 == nil {
		t.Fatal("NewWholesaleService returned nil for one of the tenants")
	}
	if svc1 == svc2 {
		t.Error("expected separate service instances for different tenants")
	}
}

// ---------------------------------------------------------------------------
// ShopeeWholesaleService constructor
// ---------------------------------------------------------------------------

func TestNewShopeeWholesaleService_NotNil(t *testing.T) {
	svc := wholesale.NewShopeeWholesaleService(nil, "tenant-1", nil)
	if svc == nil {
		t.Fatal("NewShopeeWholesaleService returned nil")
	}
}

// ---------------------------------------------------------------------------
// ShopeeMpqService constructor
// ---------------------------------------------------------------------------

func TestNewShopeeMpqService_NotNil(t *testing.T) {
	svc := wholesale.NewShopeeMpqService(nil, "tenant-1", nil)
	if svc == nil {
		t.Fatal("NewShopeeMpqService returned nil")
	}
}

// ---------------------------------------------------------------------------
// MpqResult struct
// ---------------------------------------------------------------------------

func TestMpqResult_SuccessFields(t *testing.T) {
	r := wholesale.MpqResult{
		ItemID:  12345,
		Success: true,
		Message: "MPQ mode set: MPQ=5, Price=100.00",
	}
	if r.ItemID != 12345 {
		t.Errorf("expected ItemID=12345, got %d", r.ItemID)
	}
	if !r.Success {
		t.Error("expected Success=true")
	}
	if r.Message == "" {
		t.Error("expected non-empty Message")
	}
	if r.Error != "" {
		t.Errorf("expected empty Error, got %q", r.Error)
	}
}

func TestMpqResult_ErrorFields(t *testing.T) {
	r := wholesale.MpqResult{
		ItemID:  99,
		Success: false,
		Error:   "Shopee API client not configured",
	}
	if r.ItemID != 99 {
		t.Errorf("expected ItemID=99, got %d", r.ItemID)
	}
	if r.Success {
		t.Error("expected Success=false")
	}
	if r.Error == "" {
		t.Error("expected non-empty Error")
	}
}

// ---------------------------------------------------------------------------
// WholesaleTier struct
// ---------------------------------------------------------------------------

func TestWholesaleTier_Fields(t *testing.T) {
	tier := wholesale.WholesaleTier{
		MinCount:  5,
		MaxCount:  9,
		UnitPrice: 90.0,
	}
	if tier.MinCount != 5 {
		t.Errorf("expected MinCount=5, got %d", tier.MinCount)
	}
	if tier.MaxCount != 9 {
		t.Errorf("expected MaxCount=9, got %d", tier.MaxCount)
	}
	if tier.UnitPrice != 90.0 {
		t.Errorf("expected UnitPrice=90.0, got %f", tier.UnitPrice)
	}
}

// ---------------------------------------------------------------------------
// ShopeeWholesaleService: nil API returns error
// ---------------------------------------------------------------------------

func TestShopeeWholesaleService_DeleteWholesaleTiers_NilAPI(t *testing.T) {
	svc := wholesale.NewShopeeWholesaleService(nil, "tenant-1", nil)
	err := svc.DeleteWholesaleTiers(context.Background(), 12345)
	if err == nil {
		t.Fatal("expected error when shopeeAPI is nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' error, got: %v", err)
	}
}

func TestShopeeWholesaleService_UpdateWholesaleTiers_NilAPI(t *testing.T) {
	svc := wholesale.NewShopeeWholesaleService(nil, "tenant-1", nil)
	tiers := []wholesale.WholesaleTier{
		{MinCount: 5, MaxCount: 9, UnitPrice: 90.0},
	}
	err := svc.UpdateWholesaleTiers(context.Background(), 12345, tiers)
	if err == nil {
		t.Fatal("expected error when shopeeAPI is nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' error, got: %v", err)
	}
}

func TestShopeeWholesaleService_GetWholesaleTiers_NilAPI(t *testing.T) {
	svc := wholesale.NewShopeeWholesaleService(nil, "tenant-1", nil)
	result, err := svc.GetWholesaleTiers(context.Background(), 12345)
	if err == nil {
		t.Fatal("expected error when shopeeAPI is nil")
	}
	if result != nil {
		t.Error("expected nil result when shopeeAPI is nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// ShopeeMpqService: nil API returns error
// ---------------------------------------------------------------------------

func TestShopeeMpqService_SetMpq_NilAPI(t *testing.T) {
	svc := wholesale.NewShopeeMpqService(nil, "tenant-1", nil)
	err := svc.SetMpq(context.Background(), 12345, 5)
	if err == nil {
		t.Fatal("expected error when shopeeAPI is nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' error, got: %v", err)
	}
}

func TestShopeeMpqService_UpdatePrice_NilAPI(t *testing.T) {
	svc := wholesale.NewShopeeMpqService(nil, "tenant-1", nil)
	err := svc.UpdatePrice(context.Background(), 12345, nil, 99.99)
	if err == nil {
		t.Fatal("expected error when shopeeAPI is nil")
	}
	if !strings.Contains(err.Error(), "not configured") {
		t.Errorf("expected 'not configured' error, got: %v", err)
	}
}

func TestShopeeMpqService_SetMpqMode_NilAPI_WithPrice(t *testing.T) {
	svc := wholesale.NewShopeeMpqService(nil, "tenant-1", nil)
	result := svc.SetMpqMode(context.Background(), 12345, 5, 100.0, nil)
	if result == nil {
		t.Fatal("expected non-nil MpqResult")
	}
	if result.Success {
		t.Error("expected Success=false when shopeeAPI is nil and price > 0")
	}
	if result.Error == "" {
		t.Error("expected non-empty Error in MpqResult")
	}
	if result.ItemID != 12345 {
		t.Errorf("expected ItemID=12345, got %d", result.ItemID)
	}
}

func TestShopeeMpqService_SetMpqMode_NilAPI_ZeroPrice(t *testing.T) {
	svc := wholesale.NewShopeeMpqService(nil, "tenant-1", nil)
	result := svc.SetMpqMode(context.Background(), 12345, 5, 0, nil)
	if result == nil {
		t.Fatal("expected non-nil MpqResult")
	}
	if result.Success {
		t.Error("expected Success=false when shopeeAPI is nil")
	}
	if result.ItemID != 12345 {
		t.Errorf("expected ItemID=12345, got %d", result.ItemID)
	}
}

// ---------------------------------------------------------------------------
// WholesaleService.CalculateTiersFromSettings (admin fee formula)
// ---------------------------------------------------------------------------

func TestCalculateTiersFromSettings_AdminFee(t *testing.T) {
	svc := wholesale.NewWholesaleService(nil, "tenant-1")
	settings := &models.WholesaleSettings{
		AdminFee:      1500,
		MinOrder1:     2,
		MaxOrder1:     3,
		MaxOrderTier3: 1000,
	}
	tiers := svc.CalculateTiersFromSettings(100000, settings)

	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}

	// Tier 1: min=2, max=3, price = 100000 - 1500 + 1500/2 = 99250
	if tiers[0].MinCount != 2 {
		t.Errorf("tier[0] MinCount: expected 2, got %d", tiers[0].MinCount)
	}
	if tiers[0].MaxCount != 3 {
		t.Errorf("tier[0] MaxCount: expected 3, got %d", tiers[0].MaxCount)
	}
	if tiers[0].UnitPrice != 99250 {
		t.Errorf("tier[0] UnitPrice: expected 99250, got %f", tiers[0].UnitPrice)
	}

	// Tier 2: min=4, max=5, price = 100000 - 1500 + 1500/4 = 98875
	if tiers[1].MinCount != 4 {
		t.Errorf("tier[1] MinCount: expected 4, got %d", tiers[1].MinCount)
	}
	if tiers[1].MaxCount != 5 {
		t.Errorf("tier[1] MaxCount: expected 5, got %d", tiers[1].MaxCount)
	}
	if tiers[1].UnitPrice != 98875 {
		t.Errorf("tier[1] UnitPrice: expected 98875, got %f", tiers[1].UnitPrice)
	}

	// Tier 3: min=6, max=1000, price = 100000 - 1500 + 1500/6 = 98750
	if tiers[2].MinCount != 6 {
		t.Errorf("tier[2] MinCount: expected 6, got %d", tiers[2].MinCount)
	}
	if tiers[2].MaxCount != 1000 {
		t.Errorf("tier[2] MaxCount: expected 1000, got %d", tiers[2].MaxCount)
	}
	if tiers[2].UnitPrice != 98750 {
		t.Errorf("tier[2] UnitPrice: expected 98750, got %f", tiers[2].UnitPrice)
	}
}

func TestCalculateTiersFromSettings_ZeroAdminFee(t *testing.T) {
	svc := wholesale.NewWholesaleService(nil, "tenant-1")
	settings := &models.WholesaleSettings{
		AdminFee:      0,
		MinOrder1:     2,
		MaxOrder1:     3,
		MaxOrderTier3: 1000,
	}
	tiers := svc.CalculateTiersFromSettings(50000, settings)

	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}
	// 0 admin fee = price stays the same
	for i, tier := range tiers {
		if tier.UnitPrice != 50000 {
			t.Errorf("tier[%d] UnitPrice: expected 50000 for 0 admin fee, got %f", i, tier.UnitPrice)
		}
	}
}

func TestDefaultSettings(t *testing.T) {
	s := wholesale.DefaultSettings("test-tenant")
	if s.TenantID != "test-tenant" {
		t.Errorf("expected TenantID=test-tenant, got %s", s.TenantID)
	}
	if s.AdminFee != 1500 {
		t.Errorf("expected AdminFee=1500, got %d", s.AdminFee)
	}
	if s.MinOrder1 != 2 {
		t.Errorf("expected MinOrder1=2, got %d", s.MinOrder1)
	}
	if s.MaxOrder1 != 3 {
		t.Errorf("expected MaxOrder1=3, got %d", s.MaxOrder1)
	}
	if s.MaxOrderTier3 != 1000 {
		t.Errorf("expected MaxOrderTier3=1000, got %d", s.MaxOrderTier3)
	}
}
