package analytics_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// test helpers
// ---------------------------------------------------------------------------

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite: %v", err)
	}

	// Auto-migrate the tables we need
	if err := db.AutoMigrate(
		&models.AnalyticsSettings{},
		&models.ShopeeEscrowSync{},
		&models.ShopeeEscrowOrder{},
		&models.ShopeeEscrowItem{},
	); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}

	return db
}

func setupService(t *testing.T, db *gorm.DB) *analytics.ShopeeAnalyticsService {
	t.Helper()
	return analytics.NewShopeeAnalyticsService(db, db, "test-tenant")
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func TestGetSettings_WithData_ReturnsDTO(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Insert test settings
	settings := models.AnalyticsSettings{
		ID:                uuid.New().String(),
		TenantID:          tenantID,
		Platform:          "shopee",
		PriceColumn:       "DISCOUNTED_PRICE",
		FormulaDeduction:  2000,
		FormulaMultiplier: 0.90,
	}
	if err := db.Create(&settings).Error; err != nil {
		t.Fatalf("failed to insert test settings: %v", err)
	}

	result, err := svc.GetSettings(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PriceColumn != "DISCOUNTED_PRICE" {
		t.Errorf("expected PriceColumn=DISCOUNTED_PRICE, got %s", result.PriceColumn)
	}
	if result.FormulaDeduction != 2000 {
		t.Errorf("expected FormulaDeduction=2000, got %f", result.FormulaDeduction)
	}
	if result.FormulaMultiplier != 0.90 {
		t.Errorf("expected FormulaMultiplier=0.90, got %f", result.FormulaMultiplier)
	}
}

func TestGetSettings_NoData_ReturnsDefaults(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()

	result, err := svc.GetSettings(ctx, "nonexistent-tenant")
	if err != nil {
		t.Fatalf("GetSettings returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PriceColumn != "HARGA" {
		t.Errorf("expected default PriceColumn=HARGA, got %s", result.PriceColumn)
	}
	if result.FormulaDeduction != 1500 {
		t.Errorf("expected default FormulaDeduction=1500, got %f", result.FormulaDeduction)
	}
	if result.FormulaMultiplier != 0.84 {
		t.Errorf("expected default FormulaMultiplier=0.84, got %f", result.FormulaMultiplier)
	}
}

func TestSaveSettings_ValidInput_Saves(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	input := &dto.AnalyticsSettingsDTO{
		PriceColumn:       "HARGA",
		FormulaDeduction:  3000,
		FormulaMultiplier: 0.75,
	}

	result, err := svc.SaveSettings(ctx, tenantID, input)
	if err != nil {
		t.Fatalf("SaveSettings returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.FormulaDeduction != 3000 {
		t.Errorf("expected FormulaDeduction=3000, got %f", result.FormulaDeduction)
	}
	if result.FormulaMultiplier != 0.75 {
		t.Errorf("expected FormulaMultiplier=0.75, got %f", result.FormulaMultiplier)
	}

	// Verify persistence
	readBack, err := svc.GetSettings(ctx, tenantID)
	if err != nil {
		t.Fatalf("GetSettings after save returned error: %v", err)
	}
	if readBack.FormulaDeduction != 3000 {
		t.Errorf("persisted FormulaDeduction=3000, got %f", readBack.FormulaDeduction)
	}
}

// ---------------------------------------------------------------------------
// Sync Status
// ---------------------------------------------------------------------------

func TestGetSyncStatus_ExistingSync_ReturnsStatus(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	now := time.Now().UTC()
	sync := models.ShopeeEscrowSync{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		Month:        3,
		Year:         2026,
		TotalOrders:  50,
		FailedOrders: 2,
		SyncedAt:     now,
	}
	if err := db.Create(&sync).Error; err != nil {
		t.Fatalf("failed to insert test sync: %v", err)
	}

	result, err := svc.GetSyncStatus(ctx, tenantID, 3, 2026)
	if err != nil {
		t.Fatalf("GetSyncStatus returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if !result.Synced {
		t.Error("expected Synced=true")
	}
	if result.TotalOrders != 50 {
		t.Errorf("expected TotalOrders=50, got %d", result.TotalOrders)
	}
	if result.FailedOrders != 2 {
		t.Errorf("expected FailedOrders=2, got %d", result.FailedOrders)
	}
	if result.SyncedAt == nil {
		t.Fatal("expected non-nil SyncedAt")
	}
}

func TestGetSyncStatus_NoSync_ReturnsUnsynced(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()

	result, err := svc.GetSyncStatus(ctx, "test-tenant", 4, 2026)
	if err != nil {
		t.Fatalf("GetSyncStatus returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Synced {
		t.Error("expected Synced=false for missing sync record")
	}
}

// ---------------------------------------------------------------------------
// Delete Sync Data
// ---------------------------------------------------------------------------

func TestDeleteSyncData_ClearsOrdersAndItems(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Insert sync record
	sync := models.ShopeeEscrowSync{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		Month:    3,
		Year:     2026,
		SyncedAt: time.Now(),
	}
	if err := db.Create(&sync).Error; err != nil {
		t.Fatalf("failed to insert sync: %v", err)
	}

	// Insert order
	order := models.ShopeeEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		OrderSN:  "ORD-001",
		Month:    3,
		Year:     2026,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to insert order: %v", err)
	}

	// Insert item
	item := models.ShopeeEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		EscrowOrderID: order.ID,
		Month:         intPtr(3),
		Year:          intPtr(2026),
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to insert item: %v", err)
	}

	// Execute delete
	if err := svc.DeleteSyncData(ctx, tenantID, 3, 2026); err != nil {
		t.Fatalf("DeleteSyncData returned error: %v", err)
	}

	// Verify all records deleted
	var syncCount int64
	db.Model(&models.ShopeeEscrowSync{}).Where("month = 3 AND year = 2026").Count(&syncCount)
	if syncCount != 0 {
		t.Errorf("expected 0 sync records, got %d", syncCount)
	}

	var orderCount int64
	db.Model(&models.ShopeeEscrowOrder{}).Where("month = 3 AND year = 2026").Count(&orderCount)
	if orderCount != 0 {
		t.Errorf("expected 0 orders, got %d", orderCount)
	}

	var itemCount int64
	db.Model(&models.ShopeeEscrowItem{}).Where("tenant_id = ?", tenantID).Count(&itemCount)
	if itemCount != 0 {
		t.Errorf("expected 0 items, got %d", itemCount)
	}
}

// ---------------------------------------------------------------------------
// Reconciliation
// ---------------------------------------------------------------------------

func TestGetReconciliation_WithOrders_ReturnsSummaryAndDetails(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Insert order
	order := models.ShopeeEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		OrderSN:  "ORD-RECON-001",
		Month:    3,
		Year:     2026,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to insert order: %v", err)
	}

	// Insert items with known prices
	sku := "SKU-TEST-001"
	modelSku := "MODEL-SKU-001"
	items := []models.ShopeeEscrowItem{
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order.ID,
			Sku:           &sku,
			ModelSku:      &modelSku,
			ItemName:      stringPtr("Test Item"),
			Quantity:      2,
			SellingPrice:  10000, // total = 20000
			OriginalPrice: 9500,  // total = 19000
			Month:         intPtr(3),
			Year:          intPtr(2026),
		},
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order.ID,
			Sku:           &sku,
			ModelSku:      &modelSku,
			ItemName:      stringPtr("Test Item"),
			Quantity:      1,
			SellingPrice:  10000, // total = 10000
			OriginalPrice: 10000, // total = 10000
			Month:         intPtr(3),
			Year:          intPtr(2026),
		},
	}
	for _, item := range items {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to insert item: %v", err)
		}
	}

	result, err := svc.GetReconciliation(ctx, tenantID, 3, 2026)
	if err != nil {
		t.Fatalf("GetReconciliation returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Summary checks
	if result.Summary.TotalSKU != 1 {
		t.Errorf("expected TotalSKU=1, got %d", result.Summary.TotalSKU)
	}
	if result.Summary.TotalTransactions != 3 {
		t.Errorf("expected TotalTransactions=3, got %d", result.Summary.TotalTransactions)
	}

	// Detail checks
	if len(result.Details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(result.Details))
	}
	detail := result.Details[0]
	if detail.SKU != modelSku {
		t.Errorf("expected SKU=%s, got %s", modelSku, detail.SKU)
	}
	if detail.TotalQuantity != 3 {
		t.Errorf("expected TotalQuantity=3, got %d", detail.TotalQuantity)
	}
	if detail.TotalAmount != 30000 {
		t.Errorf("expected TotalAmount=30000, got %f", detail.TotalAmount)
	}
	if detail.SystemAmount != 29000 {
		t.Errorf("expected SystemAmount=29000, got %f", detail.SystemAmount)
	}
	if detail.PriceDiff != -1000 {
		t.Errorf("expected PriceDiff=-1000, got %f", detail.PriceDiff)
	}
}

func TestGetReconciliation_NoData_ReturnsEmpty(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()

	result, err := svc.GetReconciliation(ctx, "test-tenant", 6, 2026)
	if err != nil {
		t.Fatalf("GetReconciliation returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Summary.TotalSKU != 0 {
		t.Errorf("expected TotalSKU=0, got %d", result.Summary.TotalSKU)
	}
	if result.Summary.TotalTransactions != 0 {
		t.Errorf("expected TotalTransactions=0, got %d", result.Summary.TotalTransactions)
	}
	if len(result.Details) != 0 {
		t.Errorf("expected empty details, got %d items", len(result.Details))
	}
}

// ---------------------------------------------------------------------------
// Shipping Fee Analysis
// ---------------------------------------------------------------------------

func TestGetShippingFeeAnalysis_WithOrders_ReturnsFeeSummary(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	orderDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	orders := []models.ShopeeEscrowOrder{
		{
			ID:                   uuid.New().String(),
			TenantID:             tenantID,
			OrderSN:              "ORD-SHIP-001",
			Month:                3,
			Year:                 2026,
			BuyerPaidShippingFee: 10000, // platform fee
			ActualShippingFee:    8000,  // actual fee -> profit 2000
			OrderDate:            &orderDate,
		},
		{
			ID:                   uuid.New().String(),
			TenantID:             tenantID,
			OrderSN:              "ORD-SHIP-002",
			Month:                3,
			Year:                 2026,
			BuyerPaidShippingFee: 5000,
			ActualShippingFee:    7000, // actual fee > platform -> loss -2000
			OrderDate:            &orderDate,
		},
		{
			ID:                   uuid.New().String(),
			TenantID:             tenantID,
			OrderSN:              "ORD-SHIP-003",
			Month:                3,
			Year:                 2026,
			BuyerPaidShippingFee: 10000,
			ActualShippingFee:    10000, // no diff
			OrderDate:            &orderDate,
		},
	}
	for _, o := range orders {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("failed to insert order: %v", err)
		}
	}

	result, err := svc.GetShippingFeeAnalysis(ctx, tenantID, 3, 2026)
	if err != nil {
		t.Fatalf("GetShippingFeeAnalysis returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	// Summary checks
	if result.Summary.TotalOrders != 3 {
		t.Errorf("expected TotalOrders=3, got %d", result.Summary.TotalOrders)
	}
	if result.Summary.OrdersWithDifference != 2 {
		t.Errorf("expected OrdersWithDifference=2, got %d", result.Summary.OrdersWithDifference)
	}
	if result.Summary.TotalProfit != 2000 {
		t.Errorf("expected TotalProfit=2000, got %f", result.Summary.TotalProfit)
	}
	if result.Summary.TotalLoss != -2000 {
		t.Errorf("expected TotalLoss=-2000, got %f", result.Summary.TotalLoss)
	}
	if result.Summary.NetImpact != 0 {
		t.Errorf("expected NetImpact=0, got %f", result.Summary.NetImpact)
	}

	// Detail checks
	if len(result.Details) != 3 {
		t.Fatalf("expected 3 details, got %d", len(result.Details))
	}

	// First order: profit
	if result.Details[0].OrderSN != "ORD-SHIP-001" {
		t.Errorf("expected OrderSN=ORD-SHIP-001, got %s", result.Details[0].OrderSN)
	}
	if result.Details[0].Difference != 2000 {
		t.Errorf("expected Difference=2000, got %f", result.Details[0].Difference)
	}
	if result.Details[0].Status != "profit" {
		t.Errorf("expected Status=profit, got %s", result.Details[0].Status)
	}

	// Second order: loss
	if result.Details[1].Difference != -2000 {
		t.Errorf("expected Difference=-2000, got %f", result.Details[1].Difference)
	}
	if result.Details[1].Status != "loss" {
		t.Errorf("expected Status=loss, got %s", result.Details[1].Status)
	}

	// Third order: ok
	if result.Details[2].Difference != 0 {
		t.Errorf("expected Difference=0, got %f", result.Details[2].Difference)
	}
	if result.Details[2].Status != "ok" {
		t.Errorf("expected Status=ok, got %s", result.Details[2].Status)
	}
}

func TestGetShippingFeeAnalysis_NoOrders_ReturnsEmpty(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()

	result, err := svc.GetShippingFeeAnalysis(ctx, "test-tenant", 7, 2026)
	if err != nil {
		t.Fatalf("GetShippingFeeAnalysis returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Summary.TotalOrders != 0 {
		t.Errorf("expected TotalOrders=0, got %d", result.Summary.TotalOrders)
	}
	if len(result.Details) != 0 {
		t.Errorf("expected empty details, got %d items", len(result.Details))
	}
}


// ---------------------------------------------------------------------------
// Formula Deterministic Tests
// ---------------------------------------------------------------------------

func TestComputeShopeeShippingDiff_WithFixture_ReturnsCorrectDifference(t *testing.T) {
	// Fixture: buyer_paid=10000, actual=7000, rebate=1500
	// Expected: 10000 - 7000 + 1500 = 4500
	diff := analytics.ComputeShopeeShippingDiff(10000, 7000, 1500)
	if diff != 4500 {
		t.Errorf("expected 4500, got %f", diff)
	}
}

func TestComputeShopeeShippingDiff_NoRebate_Works(t *testing.T) {
	// When rebate is 0, formula reduces to buyerPaid - actual
	diff := analytics.ComputeShopeeShippingDiff(10000, 8000, 0)
	if diff != 2000 {
		t.Errorf("expected 2000, got %f", diff)
	}
}

func TestComputeShopeeShippingDiff_NegativeDifference_Works(t *testing.T) {
	// When actual > buyerPaid, result is negative
	diff := analytics.ComputeShopeeShippingDiff(5000, 7000, 500)
	// 5000 - 7000 + 500 = -1500
	if diff != -1500 {
		t.Errorf("expected -1500, got %f", diff)
	}
}
// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

