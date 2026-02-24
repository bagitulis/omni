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
	// They are separate instances (pointer inequality)
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
	// price > 0 triggers UpdatePrice which returns error when API is nil
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
	// price == 0 skips UpdatePrice, but SetMpq still fails when API is nil
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
// WholesaleService.CalculateTiersFromSettings (pure function, no DB/API)
// ---------------------------------------------------------------------------

func TestCalculateTiersFromSettings_ThreeTiers(t *testing.T) {
	svc := wholesale.NewWholesaleService(nil, "tenant-1")
	settings := &models.WholesaleSettings{
		MinQty1:   5,
		Discount1: 10.0,
		MinQty2:   10,
		Discount2: 20.0,
		MinQty3:   20,
		Discount3: 30.0,
	}
	tiers := svc.CalculateTiersFromSettings(100.0, settings)

	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}

	// Tier 1: 100 * (1 - 10/100) = 90.0
	if tiers[0].MinCount != 5 {
		t.Errorf("tier[0] MinCount: expected 5, got %d", tiers[0].MinCount)
	}
	if tiers[0].MaxCount != 9 { // MinQty2 - 1
		t.Errorf("tier[0] MaxCount: expected 9, got %d", tiers[0].MaxCount)
	}
	if tiers[0].UnitPrice != 90.0 {
		t.Errorf("tier[0] UnitPrice: expected 90.0, got %f", tiers[0].UnitPrice)
	}

	// Tier 2: 100 * (1 - 20/100) = 80.0
	if tiers[1].MinCount != 10 {
		t.Errorf("tier[1] MinCount: expected 10, got %d", tiers[1].MinCount)
	}
	if tiers[1].MaxCount != 19 { // MinQty3 - 1
		t.Errorf("tier[1] MaxCount: expected 19, got %d", tiers[1].MaxCount)
	}
	if tiers[1].UnitPrice != 80.0 {
		t.Errorf("tier[1] UnitPrice: expected 80.0, got %f", tiers[1].UnitPrice)
	}

	// Tier 3: 100 * (1 - 30/100) = 70.0, MaxCount=0 (unlimited)
	if tiers[2].MinCount != 20 {
		t.Errorf("tier[2] MinCount: expected 20, got %d", tiers[2].MinCount)
	}
	if tiers[2].MaxCount != 0 {
		t.Errorf("tier[2] MaxCount: expected 0 (unlimited), got %d", tiers[2].MaxCount)
	}
	if tiers[2].UnitPrice != 70.0 {
		t.Errorf("tier[2] UnitPrice: expected 70.0, got %f", tiers[2].UnitPrice)
	}
}

func TestCalculateTiersFromSettings_ZeroDiscount(t *testing.T) {
	svc := wholesale.NewWholesaleService(nil, "tenant-1")
	settings := &models.WholesaleSettings{
		MinQty1:   1,
		Discount1: 0.0,
		MinQty2:   5,
		Discount2: 0.0,
		MinQty3:   10,
		Discount3: 0.0,
	}
	tiers := svc.CalculateTiersFromSettings(50.0, settings)

	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}
	// 0% discount means full price retained
	for i, tier := range tiers {
		if tier.UnitPrice != 50.0 {
			t.Errorf("tier[%d] UnitPrice: expected 50.0 for 0%% discount, got %f", i, tier.UnitPrice)
		}
	}
}

func TestCalculateTiersFromSettings_FullDiscount(t *testing.T) {
	svc := wholesale.NewWholesaleService(nil, "tenant-1")
	settings := &models.WholesaleSettings{
		MinQty1:   1,
		Discount1: 100.0,
		MinQty2:   5,
		Discount2: 100.0,
		MinQty3:   10,
		Discount3: 100.0,
	}
	tiers := svc.CalculateTiersFromSettings(200.0, settings)

	if len(tiers) != 3 {
		t.Fatalf("expected 3 tiers, got %d", len(tiers))
	}
	// 100% discount → price = 0
	for i, tier := range tiers {
		if tier.UnitPrice != 0.0 {
			t.Errorf("tier[%d] UnitPrice: expected 0.0 for 100%% discount, got %f", i, tier.UnitPrice)
		}
	}
}
