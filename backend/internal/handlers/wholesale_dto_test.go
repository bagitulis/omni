package handlers

import (
	"testing"
)

// TestWholesaleDTOPackage tests all wholesale DTO structures with new field names
func TestWholesaleDTOPackage(t *testing.T) {
	// Test WholesaleTier with Shopee API field names
	tier := WholesaleTier{
		MinCount:  5,
		MaxCount:  10,
		UnitPrice: 90000,
	}
	if tier.MinCount != 5 {
		t.Errorf("Expected min_count 5, got %d", tier.MinCount)
	}
	if tier.UnitPrice != 90000 {
		t.Errorf("Expected unit_price 90000, got %f", tier.UnitPrice)
	}

	// Test WholesaleInfo
	info := WholesaleInfo{
		ItemID:       1001,
		HasWholesale: true,
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
}
