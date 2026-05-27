package analytics_test

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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
		&models.InventoryRecord{},
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
			ModelName:     stringPtr("Test Item"),
			Quantity:      2,
			SellingPrice:  10000,
			OriginalPrice: 9500,
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
			ModelName:     stringPtr("Test Item"),
			Quantity:      1,
			SellingPrice:  10000,
			OriginalPrice: 10000,
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
	if result.Summary.TotalSku != 1 {
		t.Errorf("expected TotalSku=1, got %d", result.Summary.TotalSku)
	}
	if result.Summary.TotalTransactions != 2 {
		t.Errorf("expected TotalTransactions=2, got %d", result.Summary.TotalTransactions)
	}

	// Detail checks (new rich DTO)
	if len(result.SkuGroups) != 1 {
		t.Fatalf("expected 1 sku group, got %d", len(result.SkuGroups))
	}
	detail := result.SkuGroups[0]
	if detail.Sku != modelSku {
		t.Errorf("expected Sku=%s, got %s", modelSku, detail.Sku)
	}
	if detail.TotalTransactions != 2 {
		t.Errorf("expected TotalTransactions=2, got %d", detail.TotalTransactions)
	}
	// Without inventory, status should be NO_INVENTORY
	if detail.Status != "NO_INVENTORY" {
		t.Errorf("expected Status=NO_INVENTORY (no inventory record), got %s", detail.Status)
	}
	// Two items with unit prices 9500/2=4750 and 10000/1=10000
	if len(detail.UniqueUnitPrices) != 2 {
		t.Errorf("expected 2 unique unit prices, got %d: %v", len(detail.UniqueUnitPrices), detail.UniqueUnitPrices)
	}
	if !detail.HasMultiplePrices {
		t.Error("expected HasMultiplePrices=true")
	}
	if detail.InventoryPrice != nil {
		if detail.ModelSku != modelSku {
			t.Errorf("expected ModelSku=%s, got %s", modelSku, detail.ModelSku)
		}
		if detail.ItemName != "Test Item" {
			t.Errorf("expected ItemName=Test Item, got %s", detail.ItemName)
		}
		if detail.VariantName != "Test Item" {
			t.Errorf("expected VariantName=Test Item, got %s", detail.VariantName)
		}
		// UniqueActualIncomes should be empty (multi-item orders have no per-order escrow mapping)
		if len(detail.UniqueActualIncomes) != 0 {
			t.Errorf("expected 0 unique actual incomes for multi-item orders, got %d: %v", len(detail.UniqueActualIncomes), detail.UniqueActualIncomes)
		}
		if detail.HasPriceDifference {
			t.Error("expected HasPriceDifference=false (no inventory to compare)")
		}
		t.Errorf("expected InventoryPrice=nil (no inventory), got %v", *detail.InventoryPrice)
	}
}

func TestGetReconciliation_NonPositiveQuantity_DefaultsToOne(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	order := models.ShopeeEscrowOrder{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		OrderSN:      "ORD-NEG-QTY-001",
		Month:        4,
		Year:         2026,
		EscrowAmount: 12000,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to insert order: %v", err)
	}

	modelSku := "MODEL-NEG-QTY"
	item := models.ShopeeEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		EscrowOrderID: order.ID,
		ModelSku:      &modelSku,
		Quantity:      -2,
		OriginalPrice: 12000,
		Month:         intPtr(4),
		Year:          intPtr(2026),
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to insert item: %v", err)
	}

	result, err := svc.GetReconciliation(ctx, tenantID, 4, 2026)
	if err != nil {
		t.Fatalf("GetReconciliation returned error: %v", err)
	}
	if len(result.SkuGroups) != 1 {
		t.Fatalf("expected 1 sku group, got %d", len(result.SkuGroups))
	}
	if got := result.SkuGroups[0].UniqueUnitPrices[0]; got != 12000 {
		t.Fatalf("expected non-positive quantity to default unit price to 12000, got %f", got)
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
	if result.Summary.TotalSku != 0 {
		t.Errorf("expected TotalSku=0, got %d", result.Summary.TotalSku)
	}
	if result.Summary.TotalTransactions != 0 {
		t.Errorf("expected TotalTransactions=0, got %d", result.Summary.TotalTransactions)
	}
	if len(result.SkuGroups) != 0 {
		t.Errorf("expected empty sku groups, got %d items", len(result.SkuGroups))
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
// Rich Reconciliation Tests
// ---------------------------------------------------------------------------

func TestGetReconciliation_OKStatus_WithInventory_ReturnsOK(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Insert order
	order := models.ShopeeEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		OrderSN:  "ORD-OK-001",
		Month:    3,
		Year:     2026,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to insert order: %v", err)
	}

	// Insert a single item with unit price matching inventory
	inventoryPrice := 10000.0
	sku := "SKU-OK-TEST"
	items := []models.ShopeeEscrowItem{
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order.ID,
			Sku:           &sku,
			ModelSku:      &sku,
			ItemName:      stringPtr("OK Item"),
			Quantity:      1,
			OriginalPrice: 10000,
			Month:         intPtr(3),
			Year:          intPtr(2026),
		},
	}
	for _, item := range items {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to insert item: %v", err)
		}
	}

	// Insert inventory record so reconciliation finds it
	invData := `{"HARGA":10000,"Nama Barang":"OK Inventory Item"}`
	invRecord := models.InventoryRecord{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		Data:          invData,
		KeyValue:      sku,
		KeyColumnName: "SKU",
	}
	if err := db.Create(&invRecord).Error; err != nil {
		t.Fatalf("failed to insert inventory record: %v", err)
	}

	result, err := svc.GetReconciliation(ctx, tenantID, 3, 2026)
	if err != nil {
		t.Fatalf("GetReconciliation returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.SkuGroups) != 1 {
		t.Fatalf("expected 1 sku group, got %d", len(result.SkuGroups))
	}
	detail := result.SkuGroups[0]
	if detail.Status != "OK" {
		t.Errorf("expected Status=OK, got %s", detail.Status)
	}
	if detail.InventoryPrice == nil {
		t.Fatal("expected InventoryPrice to be set")
	}
	if *detail.InventoryPrice != inventoryPrice {
		t.Errorf("expected InventoryPrice=10000, got %f", *detail.InventoryPrice)
	}
	if detail.ExpectedIncome == nil {
		t.Fatal("expected ExpectedIncome to be set")
	}
	// Expected: (10000 - 1500) * 0.84 = 7140
	expected := (10000 - 1500) * 0.84
	if *detail.ExpectedIncome != expected {
		t.Errorf("expected ExpectedIncome=%f, got %f", expected, *detail.ExpectedIncome)
	}
	if detail.HasMultiplePrices {
		t.Error("expected HasMultiplePrices=false for single item")
	}
	if detail.ModelSku != sku {
		t.Errorf("expected ModelSku=%s, got %s", sku, detail.ModelSku)
	}
	if detail.VariantName != "" {
		t.Errorf("expected VariantName=\"\" (no variant info), got %s", detail.VariantName)
	}
	if detail.HasPriceDifference {
		t.Error("expected HasPriceDifference=false (unit price matches inventory)")
	}
}

func TestGetReconciliation_PRICE_DIFF_Status_WithMismatch(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	order := models.ShopeeEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		OrderSN:  "ORD-DIFF-001",
		Month:    3,
		Year:     2026,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to insert order: %v", err)
	}

	sku := "SKU-DIFF-TEST"
	// Item with unit price 8000, but inventory says 10000 → PRICE_DIFF
	item := models.ShopeeEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		EscrowOrderID: order.ID,
		Sku:           &sku,
		ModelSku:      &sku,
		ItemName:      stringPtr("Price Diff Item"),
		Quantity:      1,
		OriginalPrice: 8000,
		Month:         intPtr(3),
		Year:          intPtr(2026),
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to insert item: %v", err)
	}

	// Inventory has different price
	invData := `{"HARGA":10000,"Nama Barang":"Diff Inventory"}`
	invRecord := models.InventoryRecord{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		Data:          invData,
		KeyValue:      sku,
		KeyColumnName: "SKU",
	}
	if err := db.Create(&invRecord).Error; err != nil {
		t.Fatalf("failed to insert inventory record: %v", err)
	}

	result, err := svc.GetReconciliation(ctx, tenantID, 3, 2026)
	if err != nil {
		t.Fatalf("GetReconciliation returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.SkuGroups) != 1 {
		t.Fatalf("expected 1 sku group, got %d", len(result.SkuGroups))
	}
	detail := result.SkuGroups[0]
	if detail.Status != "PRICE_DIFF" {
		t.Errorf("expected Status=PRICE_DIFF, got %s", detail.Status)
	}
	if !detail.HasPriceDifference {
		t.Error("expected HasPriceDifference=true when unit price differs from inventory")
	}
}

func TestGetReconciliation_MultipleActualIncomes_TracksCorrectly(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Two separate single-item orders to get escrow amounts
	order1 := models.ShopeeEscrowOrder{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		OrderSN:      "ORD-INC-001",
		Month:        6,
		Year:         2026,
		EscrowAmount: 9500, // actual income for this order
	}
	if err := db.Create(&order1).Error; err != nil {
		t.Fatalf("failed to insert order1: %v", err)
	}
	order2 := models.ShopeeEscrowOrder{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		OrderSN:      "ORD-INC-002",
		Month:        6,
		Year:         2026,
		EscrowAmount: 10500, // different actual income
	}
	if err := db.Create(&order2).Error; err != nil {
		t.Fatalf("failed to insert order2: %v", err)
	}

	sku := "SKU-INC-TEST"
	modelSku := "SKU-INC-MODEL"
	items := []models.ShopeeEscrowItem{
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order1.ID,
			Sku:           &sku,
			ModelSku:      &modelSku,
			ItemName:      stringPtr("Income Item"),
			Quantity:      1,
			OriginalPrice: 10000,
			Month:         intPtr(6),
			Year:          intPtr(2026),
		},
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order2.ID,
			Sku:           &sku,
			ModelSku:      &modelSku,
			ItemName:      stringPtr("Income Item"),
			Quantity:      1,
			OriginalPrice: 10000,
			Month:         intPtr(6),
			Year:          intPtr(2026),
		},
	}
	for _, item := range items {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to insert item: %v", err)
		}
	}

	result, err := svc.GetReconciliation(ctx, tenantID, 6, 2026)
	if err != nil {
		t.Fatalf("GetReconciliation returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if len(result.SkuGroups) != 1 {
		t.Fatalf("expected 1 sku group, got %d", len(result.SkuGroups))
	}
	detail := result.SkuGroups[0]
	// Both orders are single-item, so actual incomes should be tracked
	if len(detail.UniqueActualIncomes) != 2 {
		t.Errorf("expected 2 unique actual incomes, got %d: %v", len(detail.UniqueActualIncomes), detail.UniqueActualIncomes)
	}
}

func TestGetReconciliation_TenantIsolation_DoesNotLeak(t *testing.T) {
	db := setupTestDB(t)
	svcA := analytics.NewShopeeAnalyticsService(db, db, "tenant-a")
	svcB := analytics.NewShopeeAnalyticsService(db, db, "tenant-b")
	ctx := context.Background()

	// Insert order + items for tenant A
	orderA := models.ShopeeEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: "tenant-a",
		OrderSN:  "ORD-A-001",
		Month:    6,
		Year:     2026,
	}
	if err := db.Create(&orderA).Error; err != nil {
		t.Fatalf("failed to insert order A: %v", err)
	}
	skuA := "SKU-A-TEST"
	itemA := models.ShopeeEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      "tenant-a",
		EscrowOrderID: orderA.ID,
		ModelSku:      &skuA,
		Quantity:      1,
		OriginalPrice: 10000,
		Month:         intPtr(6),
		Year:          intPtr(2026),
	}
	if err := db.Create(&itemA).Error; err != nil {
		t.Fatalf("failed to insert item A: %v", err)
	}

	// Query as tenant B — should see no data
	result, err := svcB.GetReconciliation(ctx, "tenant-b", 6, 2026)
	if err != nil {
		t.Fatalf("GetReconciliation for tenant B returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Summary.TotalSku != 0 {
		t.Errorf("expected TotalSku=0 for tenant B, got %d", result.Summary.TotalSku)
	}
	if len(result.SkuGroups) != 0 {
		t.Errorf("expected 0 sku groups for tenant B, got %d", len(result.SkuGroups))
	}

	// Verify tenant A still has its data
	resultA, err := svcA.GetReconciliation(ctx, "tenant-a", 6, 2026)
	if err != nil {
		t.Fatalf("GetReconciliation for tenant A returned error: %v", err)
	}
	if resultA.Summary.TotalSku != 1 {
		t.Errorf("expected TotalSku=1 for tenant A, got %d", resultA.Summary.TotalSku)
	}
}

func TestGetShippingFeeAnalysis_WithShopeeRebate_IncludesRebateInDiff(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	orderDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	order := models.ShopeeEscrowOrder{
		ID:                   uuid.New().String(),
		TenantID:             tenantID,
		OrderSN:              "ORD-REBATE-001",
		Month:                3,
		Year:                 2026,
		BuyerPaidShippingFee: 10000,
		ActualShippingFee:    7000,
		ShopeeShippingRebate: 1500,
		OrderDate:            &orderDate,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to insert order: %v", err)
	}

	result, err := svc.GetShippingFeeAnalysis(ctx, tenantID, 3, 2026)
	if err != nil {
		t.Fatalf("GetShippingFeeAnalysis returned error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.Details) != 1 {
		t.Fatalf("expected 1 detail, got %d", len(result.Details))
	}
	detail := result.Details[0]
	// 10000 - 7000 + 1500 = 4500
	if detail.Difference != 4500 {
		t.Errorf("expected Difference=4500 (10000-7000+1500), got %f", detail.Difference)
	}
	if detail.ShopeeRebate != 1500 {
		t.Errorf("expected ShopeeRebate=1500, got %f", detail.ShopeeRebate)
	}
	if detail.BuyerPaid != 10000 {
		t.Errorf("expected BuyerPaid=10000, got %f", detail.BuyerPaid)
	}
	if detail.ActualFee != 7000 {
		t.Errorf("expected ActualFee=7000, got %f", detail.ActualFee)
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

// ---------------------------------------------------------------------------
// Drill-down: GetSkuOrders & GetOrderItems
// ---------------------------------------------------------------------------

func TestGetSkuOrders_ReturnsOrders(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	orderID := uuid.New().String()
	order := models.ShopeeEscrowOrder{
		ID:            orderID,
		TenantID:      tenantID,
		OrderSN:       "SHOPEE-ORD-001",
		Month:         1,
		Year:          2025,
		EscrowAmount:  75000,
		BuyerUserName: stringPtr("Buyer A"),
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	item := models.ShopeeEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		EscrowOrderID: orderID,
		ModelSku:      stringPtr("SKU-TEST-001"),
		ItemName:      stringPtr("Test Item"),
		Quantity:      2,
		OriginalPrice: 50000,
		Month:         intPtr(1),
		Year:          intPtr(2025),
	}
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("failed to create item: %v", err)
	}

	result, err := svc.GetSkuOrders(ctx, tenantID, "SKU-TEST-001", 1, 2025)
	if err != nil {
		t.Fatalf("GetSkuOrders returned error: %v", err)
	}
	if len(result.Orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(result.Orders))
	}
	if result.Orders[0].OrderSN != "SHOPEE-ORD-001" {
		t.Errorf("expected OrderSN=SHOPEE-ORD-001, got %s", result.Orders[0].OrderSN)
	}
	if result.Orders[0].ModelSku != "SKU-TEST-001" {
		t.Errorf("expected ModelSku=SKU-TEST-001, got %s", result.Orders[0].ModelSku)
	}
	if result.Orders[0].EscrowAmount != 75000 {
		t.Errorf("expected EscrowAmount=75000, got %f", result.Orders[0].EscrowAmount)
	}
	if result.Orders[0].BuyerName != "Buyer A" {
		t.Errorf("expected BuyerName=Buyer A, got %s", result.Orders[0].BuyerName)
	}
}

func TestGetSkuOrders_NoMatch_ReturnsEmpty(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()

	result, err := svc.GetSkuOrders(ctx, "test-tenant", "NONEXISTENT-SKU", 1, 2025)
	if err != nil {
		t.Fatalf("GetSkuOrders returned error: %v", err)
	}
	if result == nil || len(result.Orders) != 0 {
		t.Errorf("expected empty orders for nonexistent SKU, got %d", len(result.Orders))
	}
}

func TestGetSkuOrders_TenantIsolation(t *testing.T) {
	db := setupTestDB(t)
	ctx := context.Background()

	// Create tenant A data
	orderA := models.ShopeeEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: "tenant-a",
		OrderSN:  "ORD-A",
		Month:    1,
		Year:     2025,
	}
	if err := db.Create(&orderA).Error; err != nil {
		t.Fatalf("failed to create order A: %v", err)
	}
	itemA := models.ShopeeEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      "tenant-a",
		EscrowOrderID: orderA.ID,
		ModelSku:      stringPtr("SKU-SHARED"),
		Month:         intPtr(1),
		Year:          intPtr(2025),
	}
	if err := db.Create(&itemA).Error; err != nil {
		t.Fatalf("failed to create item A: %v", err)
	}

	// Tenant B queries the same SKU - should get empty
	svcB := analytics.NewShopeeAnalyticsService(db, db, "tenant-b")
	result, err := svcB.GetSkuOrders(ctx, "tenant-b", "SKU-SHARED", 1, 2025)
	if err != nil {
		t.Fatalf("GetSkuOrders returned error: %v", err)
	}
	if len(result.Orders) != 0 {
		t.Errorf("expected tenant isolation: tenant-b should not see tenant-a orders, got %d", len(result.Orders))
	}
}

func TestGetOrderItems_ReturnsItems(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	orderID := uuid.New().String()
	order := models.ShopeeEscrowOrder{
		ID:       orderID,
		TenantID: tenantID,
		OrderSN:  "SHOPEE-ORD-002",
		Month:    2,
		Year:     2025,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to create order: %v", err)
	}

	items := []models.ShopeeEscrowItem{
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: orderID,
			ModelSku:      stringPtr("SKU-001"),
			ItemName:      stringPtr("Item 1"),
			Quantity:      1,
			OriginalPrice: 25000,
		},
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: orderID,
			ModelSku:      stringPtr("SKU-002"),
			ItemName:      stringPtr("Item 2"),
			Quantity:      3,
			OriginalPrice: 30000,
		},
	}
	for _, item := range items {
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to create item: %v", err)
		}
	}

	result, err := svc.GetOrderItems(ctx, tenantID, "SHOPEE-ORD-002", 2, 2025)
	if err != nil {
		t.Fatalf("GetOrderItems returned error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result.Items))
	}
	if result.Items[0].ItemName != "Item 1" {
		t.Errorf("expected ItemName=Item 1, got %s", result.Items[0].ItemName)
	}
	if result.Items[1].ItemName != "Item 2" {
		t.Errorf("expected ItemName=Item 2, got %s", result.Items[1].ItemName)
	}
}

func TestGetOrderItems_NoMatch_ReturnsEmpty(t *testing.T) {
	db := setupTestDB(t)
	svc := setupService(t, db)
	ctx := context.Background()

	result, err := svc.GetOrderItems(ctx, "test-tenant", "NONEXISTENT-ORDER", 1, 2025)
	if err != nil {
		t.Fatalf("GetOrderItems returned error: %v", err)
	}
	if result == nil || len(result.Items) != 0 {
		t.Errorf("expected empty items for nonexistent order, got %d", len(result.Items))
	}
}
