package sku_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/omni/backend/internal/services/sku"
)

// ---------------------------------------------------------------------------
// Mock PlatformSKUChecker
// ---------------------------------------------------------------------------

type mockChecker struct {
	result *sku.SKUStatus
	err    error
	calls  atomic.Int32
}

func (m *mockChecker) CheckSKU(_ context.Context, _ string) (*sku.SKUStatus, error) {
	m.calls.Add(1)
	return m.result, m.err
}

// ---------------------------------------------------------------------------
// SKUCache tests
// ---------------------------------------------------------------------------

func TestNewSKUCache_CreatesNonNilCache(t *testing.T) {
	c := sku.NewSKUCache()
	if c == nil {
		t.Fatal("NewSKUCache() returned nil")
	}
}

func TestSKUCache_Set_And_Get(t *testing.T) {
	c := sku.NewSKUCache()
	status := &sku.PlatformSKUStatus{
		SKU: "SKU-001",
		Shopee: &sku.SKUStatus{
			Exists: true,
			ItemID: 12345,
			Status: "active",
		},
	}

	c.Set("SKU-001", status)
	got := c.Get("SKU-001")

	if got == nil {
		t.Fatal("Get() returned nil after Set()")
	}
	if got.SKU != "SKU-001" {
		t.Errorf("SKU = %q, want SKU-001", got.SKU)
	}
	if got.Shopee == nil {
		t.Fatal("Shopee status should not be nil")
	}
	if !got.Shopee.Exists {
		t.Error("Shopee.Exists should be true")
	}
	if got.Shopee.ItemID != 12345 {
		t.Errorf("Shopee.ItemID = %d, want 12345", got.Shopee.ItemID)
	}
}

func TestSKUCache_Get_MissingKey_ReturnsNil(t *testing.T) {
	c := sku.NewSKUCache()
	got := c.Get("NONEXISTENT-SKU")
	if got != nil {
		t.Errorf("Get() = %v, want nil for missing key", got)
	}
}

func TestSKUCache_Delete(t *testing.T) {
	c := sku.NewSKUCache()
	status := &sku.PlatformSKUStatus{SKU: "SKU-DELETE"}
	c.Set("SKU-DELETE", status)

	// Verify it's set
	if c.Get("SKU-DELETE") == nil {
		t.Fatal("expected value before delete")
	}

	c.Delete("SKU-DELETE")

	if got := c.Get("SKU-DELETE"); got != nil {
		t.Errorf("Get() after Delete() = %v, want nil", got)
	}
}

func TestSKUCache_Delete_NonExistent_IsNoop(t *testing.T) {
	c := sku.NewSKUCache()
	// Should not panic
	c.Delete("NONEXISTENT")
}

func TestSKUCache_Clear(t *testing.T) {
	c := sku.NewSKUCache()

	// Add multiple entries
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("SKU-%03d", i)
		c.Set(key, &sku.PlatformSKUStatus{SKU: key})
	}

	// Verify at least one is set
	if c.Get("SKU-000") == nil {
		t.Fatal("expected value before clear")
	}

	c.Clear()

	// All should be gone
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("SKU-%03d", i)
		if got := c.Get(key); got != nil {
			t.Errorf("Get(%q) after Clear() = %v, want nil", key, got)
		}
	}
}

func TestSKUCache_SetTTL(t *testing.T) {
	c := sku.NewSKUCache()
	// Should not panic; functionality verified implicitly via expiry test below
	c.SetTTL(100 * time.Millisecond)
}

func TestSKUCache_Expiry(t *testing.T) {
	c := sku.NewSKUCache()
	c.SetTTL(50 * time.Millisecond) // very short TTL

	c.Set("EXPIRING-SKU", &sku.PlatformSKUStatus{SKU: "EXPIRING-SKU"})

	// Should be present immediately
	if c.Get("EXPIRING-SKU") == nil {
		t.Fatal("expected value immediately after Set()")
	}

	// Wait for expiry
	time.Sleep(100 * time.Millisecond)

	// Should now be nil
	if got := c.Get("EXPIRING-SKU"); got != nil {
		t.Errorf("Get() after TTL expiry = %v, want nil", got)
	}
}

func TestSKUCache_Stats_ReturnsCounts(t *testing.T) {
	c := sku.NewSKUCache()
	c.Set("A", &sku.PlatformSKUStatus{SKU: "A"})
	c.Set("B", &sku.PlatformSKUStatus{SKU: "B"})

	stats := c.Stats()
	if stats == nil {
		t.Fatal("Stats() returned nil")
	}
	total, ok := stats["total"]
	if !ok {
		t.Fatal("Stats() missing 'total' key")
	}
	if total.(int) < 2 {
		t.Errorf("total = %v, want >= 2", total)
	}
	if _, ok := stats["valid"]; !ok {
		t.Error("Stats() missing 'valid' key")
	}
	if _, ok := stats["expired"]; !ok {
		t.Error("Stats() missing 'expired' key")
	}
	if _, ok := stats["ttl"]; !ok {
		t.Error("Stats() missing 'ttl' key")
	}
}

func TestSKUCache_Overwrite(t *testing.T) {
	c := sku.NewSKUCache()
	first := &sku.PlatformSKUStatus{SKU: "SKU-OW", Shopee: &sku.SKUStatus{Exists: true}}
	second := &sku.PlatformSKUStatus{SKU: "SKU-OW", Shopee: &sku.SKUStatus{Exists: false}}

	c.Set("SKU-OW", first)
	c.Set("SKU-OW", second)

	got := c.Get("SKU-OW")
	if got == nil {
		t.Fatal("Get() returned nil")
	}
	if got.Shopee.Exists {
		t.Error("expected second write to overwrite first (Exists should be false)")
	}
}

// ---------------------------------------------------------------------------
// NewCheckService constructor tests
// ---------------------------------------------------------------------------

func TestNewCheckService_ReturnsNonNil(t *testing.T) {
	svc := sku.NewCheckService(nil, "tenant-123")
	if svc == nil {
		t.Fatal("NewCheckService() returned nil")
	}
}

// ---------------------------------------------------------------------------
// CheckService no-op methods
// ---------------------------------------------------------------------------

func TestCheckService_GetCachedStatus_AlwaysReturnsNil(t *testing.T) {
	svc := sku.NewCheckService(nil, "tenant-abc")
	got := svc.GetCachedStatus("any-sku")
	if got != nil {
		t.Errorf("GetCachedStatus() = %v, want nil (no-op)", got)
	}
}

func TestCheckService_InvalidateCache_IsNoop(t *testing.T) {
	svc := sku.NewCheckService(nil, "tenant-abc")
	// Must not panic
	svc.InvalidateCache("some-sku")
}

func TestCheckService_ClearCache_IsNoop(t *testing.T) {
	svc := sku.NewCheckService(nil, "tenant-abc")
	// Must not panic
	svc.ClearCache()
}

// ---------------------------------------------------------------------------
// CheckSingle tests
// ---------------------------------------------------------------------------

func TestCheckSingle_NoAPIs_ReturnsEmptyResult(t *testing.T) {
	svc := sku.NewCheckService(nil, "t1")
	result := svc.CheckSingle(context.Background(), "SKU-X", nil)

	if result == nil {
		t.Fatal("CheckSingle() returned nil")
	}
	if result.SKU != "SKU-X" {
		t.Errorf("SKU = %q, want SKU-X", result.SKU)
	}
	if result.Shopee != nil {
		t.Error("Shopee should be nil when not in apis map")
	}
	if result.Lazada != nil {
		t.Error("Lazada should be nil when not in apis map")
	}
	if result.Tiktok != nil {
		t.Error("Tiktok should be nil when not in apis map")
	}
}

func TestCheckSingle_ShopeeAPI(t *testing.T) {
	shopeeStatus := &sku.SKUStatus{Exists: true, ItemID: 99, Status: "active"}
	mock := &mockChecker{result: shopeeStatus}

	svc := sku.NewCheckService(nil, "t1")
	apis := map[string]sku.PlatformSKUChecker{
		"shopee": mock,
	}

	result := svc.CheckSingle(context.Background(), "PROD-001", apis)

	if result == nil {
		t.Fatal("CheckSingle() returned nil")
	}
	if result.SKU != "PROD-001" {
		t.Errorf("SKU = %q, want PROD-001", result.SKU)
	}
	if result.Shopee == nil {
		t.Fatal("Shopee should not be nil")
	}
	if !result.Shopee.Exists {
		t.Error("Shopee.Exists should be true")
	}
	if result.Shopee.ItemID != 99 {
		t.Errorf("Shopee.ItemID = %d, want 99", result.Shopee.ItemID)
	}
	if mock.calls.Load() != 1 {
		t.Errorf("mock called %d times, want 1", mock.calls.Load())
	}
}

func TestCheckSingle_LazadaAPI(t *testing.T) {
	lazadaStatus := &sku.SKUStatus{Exists: true, ProductID: "LAZ-123", Status: "active"}
	mock := &mockChecker{result: lazadaStatus}

	svc := sku.NewCheckService(nil, "t1")
	result := svc.CheckSingle(context.Background(), "PROD-002", map[string]sku.PlatformSKUChecker{
		"lazada": mock,
	})

	if result.Lazada == nil {
		t.Fatal("Lazada should not be nil")
	}
	if result.Lazada.ProductID != "LAZ-123" {
		t.Errorf("Lazada.ProductID = %q, want LAZ-123", result.Lazada.ProductID)
	}
}

func TestCheckSingle_TiktokAPI(t *testing.T) {
	tiktokStatus := &sku.SKUStatus{Exists: false, Error: "not found"}
	mock := &mockChecker{result: tiktokStatus}

	svc := sku.NewCheckService(nil, "t1")
	result := svc.CheckSingle(context.Background(), "PROD-003", map[string]sku.PlatformSKUChecker{
		"tiktok": mock,
	})

	if result.Tiktok == nil {
		t.Fatal("Tiktok should not be nil")
	}
	if result.Tiktok.Exists {
		t.Error("Tiktok.Exists should be false")
	}
	if result.Tiktok.Error != "not found" {
		t.Errorf("Tiktok.Error = %q, want not found", result.Tiktok.Error)
	}
}

func TestCheckSingle_MultiPlatform(t *testing.T) {
	shopeeStatus := &sku.SKUStatus{Exists: true, ItemID: 1}
	lazadaStatus := &sku.SKUStatus{Exists: true, ProductID: "LP-1"}
	tiktokStatus := &sku.SKUStatus{Exists: false}

	svc := sku.NewCheckService(nil, "t1")
	apis := map[string]sku.PlatformSKUChecker{
		"shopee": &mockChecker{result: shopeeStatus},
		"lazada": &mockChecker{result: lazadaStatus},
		"tiktok": &mockChecker{result: tiktokStatus},
	}

	result := svc.CheckSingle(context.Background(), "MULTI-SKU", apis)

	if result.Shopee == nil || !result.Shopee.Exists {
		t.Error("Shopee should exist")
	}
	if result.Lazada == nil || !result.Lazada.Exists {
		t.Error("Lazada should exist")
	}
	if result.Tiktok == nil || result.Tiktok.Exists {
		t.Error("Tiktok should not exist")
	}
}

func TestCheckSingle_APIError_DoesNotPanic(t *testing.T) {
	// CheckSingle ignores errors (uses _ = status, _ = err style)
	mock := &mockChecker{result: nil, err: fmt.Errorf("network timeout")}

	svc := sku.NewCheckService(nil, "t1")
	result := svc.CheckSingle(context.Background(), "ERR-SKU", map[string]sku.PlatformSKUChecker{
		"shopee": mock,
	})

	if result == nil {
		t.Fatal("CheckSingle() should not return nil even on API error")
	}
	// Shopee may be nil because the API returned nil status
	// The important thing is no panic
}

func TestCheckSingle_UnknownPlatform_IsIgnored(t *testing.T) {
	mock := &mockChecker{result: &sku.SKUStatus{Exists: true}}

	svc := sku.NewCheckService(nil, "t1")
	result := svc.CheckSingle(context.Background(), "P-SKU", map[string]sku.PlatformSKUChecker{
		"unknown_platform": mock,
	})

	// Unknown platform key: result fields remain nil
	if result == nil {
		t.Fatal("CheckSingle() should not return nil")
	}
	if result.Shopee != nil || result.Lazada != nil || result.Tiktok != nil {
		t.Error("Unknown platform should not populate Shopee/Lazada/Tiktok")
	}
}

// ---------------------------------------------------------------------------
// CheckBatch tests
// ---------------------------------------------------------------------------

func TestCheckBatch_EmptySlice_ReturnsEmptySlice(t *testing.T) {
	svc := sku.NewCheckService(nil, "t1")
	results := svc.CheckBatch(context.Background(), []string{}, nil)

	if results == nil {
		t.Fatal("CheckBatch() returned nil")
	}
	if len(results) != 0 {
		t.Errorf("len(results) = %d, want 0", len(results))
	}
}

func TestCheckBatch_MultipleSKUs_ReturnsCorrectCount(t *testing.T) {
	shopeeStatus := &sku.SKUStatus{Exists: true, ItemID: 1}
	mock := &mockChecker{result: shopeeStatus}

	svc := sku.NewCheckService(nil, "t1")
	skus := []string{"SKU-A", "SKU-B", "SKU-C"}
	results := svc.CheckBatch(context.Background(), skus, map[string]sku.PlatformSKUChecker{
		"shopee": mock,
	})

	if len(results) != 3 {
		t.Errorf("len(results) = %d, want 3", len(results))
	}
}

func TestCheckBatch_PreservesOrder(t *testing.T) {
	mockA := &mockChecker{result: &sku.SKUStatus{Exists: true, ItemID: 1}}

	svc := sku.NewCheckService(nil, "t1")
	skus := []string{"SKU-FIRST", "SKU-SECOND", "SKU-THIRD"}
	results := svc.CheckBatch(context.Background(), skus, map[string]sku.PlatformSKUChecker{
		"shopee": mockA,
	})

	for i, expected := range skus {
		if results[i].SKU != expected {
			t.Errorf("results[%d].SKU = %q, want %q", i, results[i].SKU, expected)
		}
	}
}

// ---------------------------------------------------------------------------
// PlatformSKUStatus struct tests
// ---------------------------------------------------------------------------

func TestPlatformSKUStatus_Fields(t *testing.T) {
	s := sku.PlatformSKUStatus{
		SKU: "MY-SKU",
		Shopee: &sku.SKUStatus{
			Exists:    true,
			ItemID:    999,
			ProductID: "P-ID",
			Status:    "active",
			Error:     "",
		},
		Lazada: &sku.SKUStatus{Exists: false},
	}
	if s.SKU != "MY-SKU" {
		t.Errorf("SKU = %q, want MY-SKU", s.SKU)
	}
	if s.Shopee.ItemID != 999 {
		t.Errorf("Shopee.ItemID = %d, want 999", s.Shopee.ItemID)
	}
	if s.Lazada.Exists {
		t.Error("Lazada.Exists should be false")
	}
	if s.Tiktok != nil {
		t.Error("Tiktok should be nil when not set")
	}
}

func TestSKUStatus_Fields(t *testing.T) {
	s := sku.SKUStatus{
		Exists:    true,
		ItemID:    111,
		ProductID: "prod-001",
		Status:    "deleted",
		Error:     "some error",
	}
	if !s.Exists {
		t.Error("Exists should be true")
	}
	if s.ItemID != 111 {
		t.Errorf("ItemID = %d, want 111", s.ItemID)
	}
	if s.Status != "deleted" {
		t.Errorf("Status = %q, want deleted", s.Status)
	}
}
