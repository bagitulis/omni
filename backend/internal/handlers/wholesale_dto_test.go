package handlers

import (
	"testing"
)

// TestWholesaleDTOStructures tests all wholesale DTO structures
func TestWholesaleDTOPackage(t *testing.T) {
	// Test WholesaleTier
	tier := WholesaleTier{
		MinQty: 5,
		MaxQty: 10,
		Price:  90000,
	}
	if tier.MinQty != 5 {
		t.Errorf("Expected min_qty 5, got %d", tier.MinQty)
	}
	if tier.Price != 90000 {
		t.Errorf("Expected price 90000, got %f", tier.Price)
	}

	// Test WholesaleInfo
	info := WholesaleInfo{
		ItemID:  1001,
		HasTier: true,
		MPQ:     5,
	}
	if info.ItemID != 1001 {
		t.Errorf("Expected item_id 1001, got %d", info.ItemID)
	}

	// Test UpdateWholesaleRequest
	updateReq := UpdateWholesaleRequest{
		Tiers: []WholesaleTier{tier},
	}
	if len(updateReq.Tiers) != 1 {
		t.Errorf("Expected 1 tier, got %d", len(updateReq.Tiers))
	}

	// Test BatchDeleteRequest
	delReq := BatchDeleteRequest{
		ItemIDs: []int64{1001, 1002},
	}
	if len(delReq.ItemIDs) != 2 {
		t.Errorf("Expected 2 item IDs, got %d", len(delReq.ItemIDs))
	}

	// Test GenerateTiers function
	tiers := GenerateTiers(100000, []int{5, 10, 15})
	if len(tiers) != 3 {
		t.Errorf("Expected 3 tiers, got %d", len(tiers))
	}
	if tiers[0].Price != 95000 {
		t.Errorf("Expected first tier price 95000, got %f", tiers[0].Price)
	}
}
