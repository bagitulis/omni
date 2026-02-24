package price_test

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/services/price"
)

// --- Constructor tests ---

func TestNewPriceService(t *testing.T) {
	svc := price.NewPriceService(nil, "tenant1")
	if svc == nil {
		t.Fatal("NewPriceService returned nil")
	}
}

func TestNewPriceServiceWithCreds(t *testing.T) {
	svc := price.NewPriceServiceWithCreds(nil, "tenant1", "/tmp/test.db")
	if svc == nil {
		t.Fatal("NewPriceServiceWithCreds returned nil")
	}
}

func TestNewPriceService_EmptyTenant(t *testing.T) {
	svc := price.NewPriceService(nil, "")
	if svc == nil {
		t.Fatal("NewPriceService returned nil for empty tenant")
	}
}

// --- PriceDTO struct and JSON tags ---

func TestPriceDTO_Fields(t *testing.T) {
	dto := price.PriceDTO{
		SKU:          "SKU-001",
		ProductName:  "Widget",
		CurrentPrice: 19.99,
		Cost:         10.00,
		Margin:       99.9,
		PlatformPrices: []price.PlatformPrice{
			{
				Platform:       "shopee",
				PlatformItemID: "item-1",
				Price:          19.99,
				OriginalPrice:  24.99,
				Currency:       "IDR",
				LastSyncedAt:   "2026-01-01T00:00:00Z",
			},
		},
	}

	if dto.SKU != "SKU-001" {
		t.Errorf("SKU: got %q, want %q", dto.SKU, "SKU-001")
	}
	if dto.ProductName != "Widget" {
		t.Errorf("ProductName: got %q, want %q", dto.ProductName, "Widget")
	}
	if dto.CurrentPrice != 19.99 {
		t.Errorf("CurrentPrice: got %v, want %v", dto.CurrentPrice, 19.99)
	}
	if dto.Cost != 10.00 {
		t.Errorf("Cost: got %v, want %v", dto.Cost, 10.00)
	}
	if dto.Margin != 99.9 {
		t.Errorf("Margin: got %v, want %v", dto.Margin, 99.9)
	}
	if len(dto.PlatformPrices) != 1 {
		t.Fatalf("PlatformPrices len: got %d, want 1", len(dto.PlatformPrices))
	}
}

func TestPriceDTO_JSONTags(t *testing.T) {
	dto := price.PriceDTO{
		SKU:          "SKU-JSON",
		ProductName:  "JSON Product",
		CurrentPrice: 9.99,
		Cost:         5.0,
		Margin:       99.8,
	}

	data, err := json.Marshal(dto)
	if err != nil {
		t.Fatalf("marshal PriceDTO: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"sku", "product_name", "current_price", "cost", "margin", "platform_prices"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}

	if m["sku"] != "SKU-JSON" {
		t.Errorf("sku: got %v, want %q", m["sku"], "SKU-JSON")
	}
}

func TestPriceDTO_JSONRoundTrip(t *testing.T) {
	original := price.PriceDTO{
		SKU:          "RT-001",
		ProductName:  "Round Trip",
		CurrentPrice: 55.5,
		Cost:         22.2,
		Margin:       150.0,
		PlatformPrices: []price.PlatformPrice{
			{Platform: "lazada", Price: 55.5, Currency: "MYR"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded price.PriceDTO
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.SKU != original.SKU {
		t.Errorf("SKU: got %q, want %q", decoded.SKU, original.SKU)
	}
	if decoded.CurrentPrice != original.CurrentPrice {
		t.Errorf("CurrentPrice: got %v, want %v", decoded.CurrentPrice, original.CurrentPrice)
	}
	if len(decoded.PlatformPrices) != 1 {
		t.Fatalf("PlatformPrices len: got %d, want 1", len(decoded.PlatformPrices))
	}
	if decoded.PlatformPrices[0].Platform != "lazada" {
		t.Errorf("platform: got %q, want %q", decoded.PlatformPrices[0].Platform, "lazada")
	}
}

// --- PlatformPrice struct and JSON tags ---

func TestPlatformPrice_Fields(t *testing.T) {
	pp := price.PlatformPrice{
		Platform:       "tiktok",
		PlatformItemID: "tt-123",
		Price:          29.99,
		OriginalPrice:  39.99,
		Currency:       "THB",
		LastSyncedAt:   "2026-02-01T10:00:00Z",
	}

	if pp.Platform != "tiktok" {
		t.Errorf("Platform: got %q, want %q", pp.Platform, "tiktok")
	}
	if pp.PlatformItemID != "tt-123" {
		t.Errorf("PlatformItemID: got %q, want %q", pp.PlatformItemID, "tt-123")
	}
	if pp.Price != 29.99 {
		t.Errorf("Price: got %v, want %v", pp.Price, 29.99)
	}
	if pp.OriginalPrice != 39.99 {
		t.Errorf("OriginalPrice: got %v, want %v", pp.OriginalPrice, 39.99)
	}
	if pp.Currency != "THB" {
		t.Errorf("Currency: got %q, want %q", pp.Currency, "THB")
	}
}

func TestPlatformPrice_JSONTags(t *testing.T) {
	pp := price.PlatformPrice{
		Platform:       "shopee",
		PlatformItemID: "sp-456",
		Price:          15.0,
		Currency:       "SGD",
	}

	data, err := json.Marshal(pp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"platform", "platform_item_id", "price", "currency"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestPlatformPrice_OmitEmptyFields(t *testing.T) {
	// OriginalPrice=0 and LastSyncedAt="" should be omitted
	pp := price.PlatformPrice{
		Platform: "shopee",
		Price:    10.0,
		Currency: "IDR",
	}

	data, err := json.Marshal(pp)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := m["original_price"]; ok {
		t.Error("original_price should be omitted when zero")
	}
	if _, ok := m["last_synced_at"]; ok {
		t.Error("last_synced_at should be omitted when empty")
	}
}

// --- PriceUpdateRequest struct and JSON tags ---

func TestPriceUpdateRequest_Fields(t *testing.T) {
	req := price.PriceUpdateRequest{
		SKU:       "SKU-UPD",
		Price:     99.0,
		Platforms: []string{"shopee", "lazada"},
	}

	if req.SKU != "SKU-UPD" {
		t.Errorf("SKU: got %q, want %q", req.SKU, "SKU-UPD")
	}
	if req.Price != 99.0 {
		t.Errorf("Price: got %v, want %v", req.Price, 99.0)
	}
	if len(req.Platforms) != 2 {
		t.Errorf("Platforms len: got %d, want 2", len(req.Platforms))
	}
}

func TestPriceUpdateRequest_JSONTags(t *testing.T) {
	req := price.PriceUpdateRequest{
		SKU:       "SKU-J",
		Price:     50.0,
		Platforms: []string{"tiktok"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"sku", "price", "platforms"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}

	if m["sku"] != "SKU-J" {
		t.Errorf("sku: got %v, want %q", m["sku"], "SKU-J")
	}
}

func TestPriceUpdateRequest_EmptyPlatforms(t *testing.T) {
	req := price.PriceUpdateRequest{
		SKU:   "SKU-ALL",
		Price: 25.0,
		// Platforms nil → all platforms
	}
	if req.Platforms != nil {
		t.Error("Platforms should be nil when not set")
	}
}

// --- BulkPriceUpdateRequest struct and JSON tags ---

func TestBulkPriceUpdateRequest_Fields(t *testing.T) {
	bulk := price.BulkPriceUpdateRequest{
		Updates: []price.PriceUpdateRequest{
			{SKU: "A", Price: 10.0},
			{SKU: "B", Price: 20.0},
		},
	}

	if len(bulk.Updates) != 2 {
		t.Fatalf("Updates len: got %d, want 2", len(bulk.Updates))
	}
	if bulk.Updates[0].SKU != "A" {
		t.Errorf("Updates[0].SKU: got %q, want %q", bulk.Updates[0].SKU, "A")
	}
}

func TestBulkPriceUpdateRequest_JSONTags(t *testing.T) {
	bulk := price.BulkPriceUpdateRequest{
		Updates: []price.PriceUpdateRequest{
			{SKU: "C", Price: 5.0},
		},
	}

	data, err := json.Marshal(bulk)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := m["updates"]; !ok {
		t.Error("missing JSON key 'updates'")
	}
}

// --- PriceUpdateResult struct and JSON tags ---

func TestPriceUpdateResult_Fields(t *testing.T) {
	result := price.PriceUpdateResult{
		SKU:     "SKU-RES",
		Success: true,
		Message: "updated",
		PlatformResults: []price.PlatformPriceResult{
			{
				Platform: "shopee",
				Success:  true,
				OldPrice: 10.0,
				NewPrice: 12.0,
			},
		},
	}

	if result.SKU != "SKU-RES" {
		t.Errorf("SKU: got %q, want %q", result.SKU, "SKU-RES")
	}
	if !result.Success {
		t.Error("Success should be true")
	}
	if len(result.PlatformResults) != 1 {
		t.Fatalf("PlatformResults len: got %d, want 1", len(result.PlatformResults))
	}
}

func TestPriceUpdateResult_JSONTags(t *testing.T) {
	result := price.PriceUpdateResult{
		SKU:     "SKU-JR",
		Success: false,
		Message: "failed",
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"sku", "success"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestPriceUpdateResult_OmitEmpty(t *testing.T) {
	// Message="" and PlatformResults=nil should be omitted
	result := price.PriceUpdateResult{
		SKU:     "SKU-OE",
		Success: true,
	}

	data, err := json.Marshal(result)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := m["message"]; ok {
		t.Error("message should be omitted when empty")
	}
	if _, ok := m["platform_results"]; ok {
		t.Error("platform_results should be omitted when nil")
	}
}

// --- PlatformPriceResult struct and JSON tags ---

func TestPlatformPriceResult_Fields(t *testing.T) {
	res := price.PlatformPriceResult{
		Platform: "lazada",
		Success:  true,
		Message:  "ok",
		OldPrice: 8.0,
		NewPrice: 9.0,
	}

	if res.Platform != "lazada" {
		t.Errorf("Platform: got %q, want %q", res.Platform, "lazada")
	}
	if !res.Success {
		t.Error("Success should be true")
	}
	if res.OldPrice != 8.0 {
		t.Errorf("OldPrice: got %v, want %v", res.OldPrice, 8.0)
	}
	if res.NewPrice != 9.0 {
		t.Errorf("NewPrice: got %v, want %v", res.NewPrice, 9.0)
	}
}

func TestPlatformPriceResult_JSONTags(t *testing.T) {
	res := price.PlatformPriceResult{
		Platform: "tiktok",
		Success:  false,
		OldPrice: 5.0,
		NewPrice: 6.0,
	}

	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"platform", "success", "old_price", "new_price"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestPlatformPriceResult_OmitEmpty(t *testing.T) {
	// Message="", OldPrice=0, NewPrice=0 should be omitted
	res := price.PlatformPriceResult{
		Platform: "shopee",
		Success:  true,
	}

	data, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if _, ok := m["message"]; ok {
		t.Error("message should be omitted when empty")
	}
	if _, ok := m["old_price"]; ok {
		t.Error("old_price should be omitted when zero")
	}
	if _, ok := m["new_price"]; ok {
		t.Error("new_price should be omitted when zero")
	}
}

// --- BulkPriceUpdateResult struct and JSON tags ---

func TestBulkPriceUpdateResult_Fields(t *testing.T) {
	bulk := price.BulkPriceUpdateResult{
		TotalRequested: 3,
		TotalSuccess:   2,
		TotalFailed:    1,
		Results: []price.PriceUpdateResult{
			{SKU: "A", Success: true},
			{SKU: "B", Success: true},
			{SKU: "C", Success: false, Message: "not found"},
		},
	}

	if bulk.TotalRequested != 3 {
		t.Errorf("TotalRequested: got %d, want 3", bulk.TotalRequested)
	}
	if bulk.TotalSuccess != 2 {
		t.Errorf("TotalSuccess: got %d, want 2", bulk.TotalSuccess)
	}
	if bulk.TotalFailed != 1 {
		t.Errorf("TotalFailed: got %d, want 1", bulk.TotalFailed)
	}
	if len(bulk.Results) != 3 {
		t.Fatalf("Results len: got %d, want 3", len(bulk.Results))
	}
}

func TestBulkPriceUpdateResult_JSONTags(t *testing.T) {
	bulk := price.BulkPriceUpdateResult{
		TotalRequested: 1,
		TotalSuccess:   1,
		TotalFailed:    0,
		Results: []price.PriceUpdateResult{
			{SKU: "X", Success: true},
		},
	}

	data, err := json.Marshal(bulk)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	for _, key := range []string{"total_requested", "total_success", "total_failed", "results"} {
		if _, ok := m[key]; !ok {
			t.Errorf("missing JSON key %q", key)
		}
	}
}

func TestBulkPriceUpdateResult_JSONRoundTrip(t *testing.T) {
	original := price.BulkPriceUpdateResult{
		TotalRequested: 2,
		TotalSuccess:   1,
		TotalFailed:    1,
		Results: []price.PriceUpdateResult{
			{SKU: "P1", Success: true},
			{SKU: "P2", Success: false, Message: "error"},
		},
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded price.BulkPriceUpdateResult
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.TotalRequested != original.TotalRequested {
		t.Errorf("TotalRequested: got %d, want %d", decoded.TotalRequested, original.TotalRequested)
	}
	if decoded.TotalFailed != original.TotalFailed {
		t.Errorf("TotalFailed: got %d, want %d", decoded.TotalFailed, original.TotalFailed)
	}
	if len(decoded.Results) != 2 {
		t.Fatalf("Results len: got %d, want 2", len(decoded.Results))
	}
	if decoded.Results[1].Message != "error" {
		t.Errorf("Results[1].Message: got %q, want %q", decoded.Results[1].Message, "error")
	}
}
