package handlers

import (
	"testing"
)

// TestInventorySheetDTOStructure verifies the inventory sheet DTO package is importable
func TestInventorySheetDTOStructure(t *testing.T) {
	// Verify SyncStatusRequest struct exists
	syncReq := SyncStatusRequest{
		SheetID: "sheet123",
	}

	if syncReq.SheetID != "sheet123" {
		t.Errorf("Expected sheet_id 'sheet123', got %s", syncReq.SheetID)
	}

	// Verify ExportToSheetRequest struct exists
	exportReq := ExportToSheetRequest{
		SheetID: "sheet456",
		Data: []map[string]interface{}{
			{"name": "Product 1", "qty": 10},
		},
	}

	if exportReq.SheetID != "sheet456" {
		t.Errorf("Expected sheet_id 'sheet456', got %s", exportReq.SheetID)
	}

	// Verify ImportFromSheetRequest struct exists
	importReq := ImportFromSheetRequest{
		SheetID:   "sheet789",
		SheetName: "Inventory",
	}

	if importReq.SheetName != "Inventory" {
		t.Errorf("Expected sheet_name 'Inventory', got %s", importReq.SheetName)
	}
}
