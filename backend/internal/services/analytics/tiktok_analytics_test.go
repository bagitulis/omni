package analytics_test

import (
	"context"
	"fmt"
	"os"
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

func setupTiktokTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := fmt.Sprintf("/tmp/tiktok_test_%s.db", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(MEMORY)&_pragma=synchronous(OFF)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open SQLite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.AnalyticsSettings{},
		&models.TiktokEscrowSync{},
		&models.TiktokEscrowOrder{},
		&models.TiktokEscrowItem{},
		&models.Job{},
	); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	return db
}

func setupTiktokService(t *testing.T, db *gorm.DB) *analytics.TiktokAnalyticsService {
	t.Helper()
	return analytics.NewTiktokAnalyticsService(db, db, "test-tenant")
}

// ---------------------------------------------------------------------------
// Settings
// ---------------------------------------------------------------------------

func TestTiktokGetSettings_WithData_ReturnsDTO(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Insert test settings
	settings := models.AnalyticsSettings{
		ID:                uuid.New().String(),
		TenantID:          tenantID,
		Platform:          "tiktok",
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

func TestTiktokGetSettings_NoData_ReturnsDefaults(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
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

func TestTiktokSaveSettings_ValidInput_Saves(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
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

func TestTiktokGetSyncStatus_ExistingSync_ReturnsStatus(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	now := time.Now().UTC()
	sync := models.TiktokEscrowSync{
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

func TestTiktokGetSyncStatus_NoSync_ReturnsUnsynced(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
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

func TestTiktokDeleteSyncData_ClearsOrdersAndItems(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Insert sync record
	sync := models.TiktokEscrowSync{
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
	order := models.TiktokEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		OrderID:  "ORD-001",
		Month:    3,
		Year:     2026,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to insert order: %v", err)
	}

	// Insert item
	item := models.TiktokEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		EscrowOrderID: order.ID,
		OrderID:       "ORD-001",
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
	db.Model(&models.TiktokEscrowSync{}).Where("month = 3 AND year = 2026").Count(&syncCount)
	if syncCount != 0 {
		t.Errorf("expected 0 sync records, got %d", syncCount)
	}

	var orderCount int64
	db.Model(&models.TiktokEscrowOrder{}).Where("month = 3 AND year = 2026").Count(&orderCount)
	if orderCount != 0 {
		t.Errorf("expected 0 orders, got %d", orderCount)
	}

	var itemCount int64
	db.Model(&models.TiktokEscrowItem{}).Where("tenant_id = ?", tenantID).Count(&itemCount)
	if itemCount != 0 {
		t.Errorf("expected 0 items, got %d", itemCount)
	}
}

// ---------------------------------------------------------------------------
// Reconciliation
// ---------------------------------------------------------------------------

func TestTiktokGetReconciliation_WithOrders_ReturnsSummaryAndDetails(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Insert order
	order := models.TiktokEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: tenantID,
		OrderID:  "ORD-RECON-001",
		Month:    3,
		Year:     2026,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to insert order: %v", err)
	}

	// Insert items with known prices
	sku := "SKU-TEST-001"
	items := []models.TiktokEscrowItem{
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order.ID,
			OrderID:       "ORD-RECON-001",
			SellerSku:     &sku,
			ProductName:   stringPtr("Test Item"),
			Quantity:      2,
			SalePrice:     10000, // total = 20000
			OriginalPrice: 9500,  // total = 19000
		},
		{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order.ID,
			OrderID:       "ORD-RECON-001",
			SellerSku:     &sku,
			ProductName:   stringPtr("Test Item"),
			Quantity:      1,
			SalePrice:     10000, // total = 10000
			OriginalPrice: 10000, // total = 10000
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
	if detail.SKU != sku {
		t.Errorf("expected SKU=%s, got %s", sku, detail.SKU)
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

func TestTiktokGetReconciliation_NoData_ReturnsEmpty(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
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

func TestTiktokGetShippingFeeAnalysis_WithOrders_ReturnsFeeSummary(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	orderDate := time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC)
	orders := []models.TiktokEscrowOrder{
		{
			ID:                      uuid.New().String(),
			TenantID:                tenantID,
			OrderID:                 "ORD-SHIP-001",
			Month:                   3,
			Year:                    2026,
			ShippingFeeCustomerPaid: 10000, // platform fee
			ShippingFeeActual:       8000,  // actual fee -> profit 2000
			OrderDate:               &orderDate,
		},
		{
			ID:                      uuid.New().String(),
			TenantID:                tenantID,
			OrderID:                 "ORD-SHIP-002",
			Month:                   3,
			Year:                    2026,
			ShippingFeeCustomerPaid: 5000,
			ShippingFeeActual:       7000, // actual fee > platform -> loss -2000
			OrderDate:               &orderDate,
		},
		{
			ID:                      uuid.New().String(),
			TenantID:                tenantID,
			OrderID:                 "ORD-SHIP-003",
			Month:                   3,
			Year:                    2026,
			ShippingFeeCustomerPaid: 10000,
			ShippingFeeActual:       10000, // no diff
			OrderDate:               &orderDate,
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

func TestTiktokGetShippingFeeAnalysis_NoOrders_ReturnsEmpty(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
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
// Sync Escrow
// ---------------------------------------------------------------------------

func TestTiktokSyncEscrow_CreatesJob(t *testing.T) {
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	jobID, err := svc.SyncEscrow(ctx, tenantID, 3, 2026, false)
	if err != nil {
		t.Fatalf("SyncEscrow returned error: %v", err)
	}
	if jobID == "" {
		t.Fatal("expected non-empty job ID")
	}

	// Verify job was created
	var job models.Job
	if err := db.Where("id = ?", jobID).First(&job).Error; err != nil {
		t.Fatalf("failed to find created job: %v", err)
	}
	if job.Type != models.JobTypeTiktokEscrowSync {
		t.Errorf("expected job type %s, got %s", models.JobTypeTiktokEscrowSync, job.Type)
	}
	if job.Status != models.JobStatusPending {
		t.Errorf("expected job status pending, got %s", job.Status)
	}
}

// ---------------------------------------------------------------------------
// Formula Deterministic Tests
// ---------------------------------------------------------------------------

func TestComputeTiktokShippingDiff_WithFixture_ReturnsCorrectDifference(t *testing.T) {
	// Fixture: customer_paid=12000, actual=8000, discount=1000
	// Expected: 12000 - 8000 + 1000 = 5000
	diff := analytics.ComputeTiktokShippingDiff(12000, 8000, 1000)
	if diff != 5000 {
		t.Errorf("expected 5000, got %f", diff)
	}
}

func TestComputeTiktokShippingDiff_NoDiscount_Works(t *testing.T) {
	// When discount is 0, formula reduces to customerPaid - actual
	diff := analytics.ComputeTiktokShippingDiff(10000, 8000, 0)
	if diff != 2000 {
		t.Errorf("expected 2000, got %f", diff)
	}
}

func TestComputeTiktokShippingDiff_NegativeDifference_Works(t *testing.T) {
	// When actual > customerPaid, result is negative
	diff := analytics.ComputeTiktokShippingDiff(5000, 7000, 300)
	// 5000 - 7000 + 300 = -1700
	if diff != -1700 {
		t.Errorf("expected -1700, got %f", diff)
	}
}

// ---------------------------------------------------------------------------
// Sync Guard Tests — fail against placeholder, pass after Wave 2/3
// ---------------------------------------------------------------------------
//
// These guardrail tests assert that SyncMonthWithProgress actually persists
// orders, items, and raw JSON data. They SHOULD FAIL against the current
// no-op/MVP placeholder sync and PASS only after real sync is implemented.
//
func TestTiktokSyncGuard_RejectsPlaceholderNoData(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupTiktokTestDB(t)
	svc := analytics.NewTiktokEscrowSyncService(db, db, "test-tenant")
	fake := newFakeTiktokClient()
	svc.SetClient(fake)
	ctx := context.Background()

	onProgress := func(processed, total int, message string) {}

	err := svc.SyncMonthWithProgress(ctx, 3, 2026, false, onProgress)
	if err != nil {
		t.Fatalf("SyncMonthWithProgress returned error: %v", err)
	}

	tenantID := "test-tenant"
	month := 3
	year := 2026

	// GUARD 1: Sync record must report TotalOrders > 0
	var sync models.TiktokEscrowSync
	if err := db.Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		First(&sync).Error; err != nil {
		t.Fatalf("failed to find sync record: %v", err)
	}
	if sync.TotalOrders == 0 {
		t.Error("GUARD FAILED: Tiktok SyncMonthWithProgress created a sync record with TotalOrders=0. ",
			"Real sync must set TotalOrders > 0 after persisting orders.")
	}

	// GUARD 2: Sync MUST persist order records
	var orderCount int64
	db.Model(&models.TiktokEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&orderCount)
	if orderCount == 0 {
		t.Error("GUARD FAILED: Tiktok SyncMonthWithProgress completed without persisting any TiktokEscrowOrder records. ",
			"Implement real sync logic that populates TiktokEscrowOrder.")
	}

	// GUARD 3: Sync MUST persist item records
	// (TiktokEscrowItem has no Month/Year, so query via order IDs)
	var orderIDsForItems []string
	db.Model(&models.TiktokEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Pluck("id", &orderIDsForItems)

	if len(orderIDsForItems) == 0 {
		t.Error("GUARD FAILED: Tiktok SyncMonthWithProgress completed without persisting any TiktokEscrowItem records ",
			"(no orders were created to associate items with).")
	} else {
		var itemCount int64
		db.Model(&models.TiktokEscrowItem{}).
			Where("tenant_id = ? AND escrow_order_id IN ?", tenantID, orderIDsForItems).
			Count(&itemCount)
		if itemCount == 0 {
			t.Error("GUARD FAILED: Tiktok SyncMonthWithProgress completed without persisting any TiktokEscrowItem records. ",
				"Implement real sync logic that populates TiktokEscrowItem.")
		}
	}

	// GUARD 4: Sync MUST store raw JSON data (raw_transaction_data or raw_order_data)
	var rawCount int64
	db.Model(&models.TiktokEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Where("(raw_transaction_data IS NOT NULL AND raw_transaction_data != '') OR (raw_order_data IS NOT NULL AND raw_order_data != '')").
		Count(&rawCount)
	if rawCount == 0 {
		t.Error("GUARD FAILED: Tiktok SyncMonthWithProgress completed without storing raw JSON data. ",
			"Sync must persist raw_transaction_data or raw_order_data for traceability.")
	}
}

// ---------------------------------------------------------------------------
// Repopulate Guard Tests — fail against no-op placeholder, pass after Wave 2/3
// ---------------------------------------------------------------------------
//
func TestTiktokRepopulateGuard_RejectsNoopOnItemRestoration(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Seed an order with raw JSON data that Wave 2/3 would use to reconstruct items
	rawTransactionData := `{"transaction_id":"TXN-001"}`
	rawOrderData := `{
  "id": "ORD-REPOP-TT-001",
  "status": "COMPLETED",
  "line_items": [
    {"id":"LI-001","product_id":"PROD-001","product_name":"Test Product A","sku_id":"SKU-001","sku_name":"Variant A","seller_sku":"TPA-001","quantity":1,"original_price":"10000","sale_price":"10000","platform_discount":"0","seller_discount":"0"},
    {"id":"LI-002","product_id":"PROD-001","product_name":"Test Product B","sku_id":"SKU-002","sku_name":"Variant B","seller_sku":"TPB-002","quantity":1,"original_price":"10000","sale_price":"10000","platform_discount":"0","seller_discount":"0"},
    {"id":"LI-003","product_id":"PROD-002","product_name":"Test Product C","sku_id":"SKU-003","sku_name":"Variant C","seller_sku":"TPC-003","quantity":1,"original_price":"10000","sale_price":"10000","platform_discount":"0","seller_discount":"0"}
  ]
}`
	order := models.TiktokEscrowOrder{
		ID:                 uuid.New().String(),
		TenantID:           tenantID,
		OrderID:            "ORD-REPOP-TT-001",
		Month:              3,
		Year:               2026,
		RawTransactionData: &rawTransactionData,
		RawOrderData:       &rawOrderData,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	// Seed items with raw JSON data
	rawItemData := `{"product_name":"Guard Test Product"}`
	itemCount := 3
	for i := 0; i < itemCount; i++ {
		item := models.TiktokEscrowItem{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order.ID,
			OrderID:       "ORD-REPOP-TT-001",
			Quantity:      1,
			SalePrice:     10000,
			OriginalPrice: 10000,
			RawItemData:   &rawItemData,
		}
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to seed item %d: %v", i, err)
		}
	}

	// Record initial count
	var initialCount int64
	db.Model(&models.TiktokEscrowItem{}).
		Where("tenant_id = ? AND escrow_order_id = ?", tenantID, order.ID).
		Count(&initialCount)
	if initialCount != int64(itemCount) {
		t.Fatalf("expected %d items initially, got %d", itemCount, initialCount)
	}

	// Clear all items (simulating data loss that repopulate should fix)
	if err := db.Where("tenant_id = ? AND escrow_order_id = ?", tenantID, order.ID).
		Delete(&models.TiktokEscrowItem{}).Error; err != nil {
		t.Fatalf("failed to clear items: %v", err)
	}
	var afterDelete int64
	db.Model(&models.TiktokEscrowItem{}).Where("tenant_id = ? AND escrow_order_id = ?", tenantID, order.ID).Count(&afterDelete)
	if afterDelete != 0 {
		t.Fatalf("expected 0 items after delete, got %d", afterDelete)
	}

	// Call RepopulateItems to restore items from raw JSON data
	if err := svc.RepopulateItems(ctx, tenantID, "2026-03"); err != nil {
		t.Fatalf("RepopulateItems returned error: %v", err)
	}

	// GUARD: Items must be restored from raw JSON data
	var finalCount int64
	db.Model(&models.TiktokEscrowItem{}).
		Where("tenant_id = ? AND escrow_order_id = ?", tenantID, order.ID).
		Count(&finalCount)
	if finalCount != initialCount {
		t.Errorf("GUARD FAILED: Tiktok RepopulateItems did not restore items from raw JSON data. "+
			"Expected %d items (matching pre-delete count), got %d. "+
			"Current RepopulateItems is a no-op — implement item reconstruction from raw JSON.",
			initialCount, finalCount)
	}
}

// ---------------------------------------------------------------------------
// Repopulate Malformed Raw JSON — must return explicit error
// ---------------------------------------------------------------------------

func TestTiktokRepopulate_MalformedRawJSON_ReturnsError(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantID := "test-tenant"

	// Seed an order with malformed raw_order_data
	malformedJSON := `{"id":"ORD-MALFORM-001",`  // truncated JSON
	order := models.TiktokEscrowOrder{
		ID:               uuid.New().String(),
		TenantID:         tenantID,
		OrderID:          "ORD-MALFORM-001",
		Month:            4,
		Year:             2026,
		RawOrderData:     &malformedJSON,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	// Also seed a good order that should NOT be corrupted by the malformed one
	goodRawData := `{"id":"ORD-GOOD-001","status":"COMPLETED","line_items":[{"id":"LI-GOOD","product_id":"PROD-GOOD","product_name":"Good Product","sku_id":"SKU-GOOD","sku_name":"Good Variant","seller_sku":"GP-001","quantity":1,"original_price":"5000","sale_price":"5000","platform_discount":"0","seller_discount":"0"}]}`
	goodOrder := models.TiktokEscrowOrder{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		OrderID:      "ORD-GOOD-001",
		Month:        4,
		Year:         2026,
		RawOrderData: &goodRawData,
	}
	if err := db.Create(&goodOrder).Error; err != nil {
		t.Fatalf("failed to seed good order: %v", err)
	}

	// Seed an item for the good order to verify it survives
	goodItem := models.TiktokEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		EscrowOrderID: goodOrder.ID,
		OrderID:       "ORD-GOOD-001",
		Quantity:      1,
		SalePrice:     5000,
		OriginalPrice: 5000,
	}
	if err := db.Create(&goodItem).Error; err != nil {
		t.Fatalf("failed to seed good item: %v", err)
	}

	// Call RepopulateItems — should fail due to malformed JSON
	err := svc.RepopulateItems(ctx, tenantID, "2026-04")
	if err == nil {
		t.Fatal("expected error for malformed raw_order_data, got nil")
	}
	// Must mention the order ID and "malformed" in the error
	if !containsSubstring(err.Error(), "ORD-MALFORM-001") ||
		!containsSubstring(err.Error(), "malformed") {
		t.Errorf("error should name the order and mention malformed JSON, got: %v", err)
	}

	// Verify the good order's item was NOT deleted (transaction rolled back)
	var goodItemCount int64
	db.Model(&models.TiktokEscrowItem{}).
		Where("tenant_id = ? AND escrow_order_id = ?", tenantID, goodOrder.ID).
		Count(&goodItemCount)
	if goodItemCount != 1 {
		t.Errorf("expected good order item count=1 (transaction rolled back), got %d", goodItemCount)
	}
}

// ---------------------------------------------------------------------------
// Repopulate Tenant Isolation — cross-tenant data must not be affected
// ---------------------------------------------------------------------------

func TestTiktokRepopulate_TenantIsolation_DoesNotAffectOtherTenants(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupTiktokTestDB(t)
	svc := setupTiktokService(t, db)
	ctx := context.Background()
	tenantA := "tenant-a"
	tenantB := "tenant-b"

	// Seed tenant A order with raw data and items
	rawDataA := `{"id":"ORD-TA-001","status":"COMPLETED","line_items":[{"id":"LI-TA","product_id":"PROD-TA","product_name":"Tenant A Product","sku_id":"SKU-TA","sku_name":"TA Variant","seller_sku":"TA-001","quantity":2,"original_price":"20000","sale_price":"20000","platform_discount":"0","seller_discount":"0"}]}`
	orderA := models.TiktokEscrowOrder{
		ID:           uuid.New().String(),
		TenantID:     tenantA,
		OrderID:      "ORD-TA-001",
		Month:        5,
		Year:         2026,
		RawOrderData: &rawDataA,
	}
	if err := db.Create(&orderA).Error; err != nil {
		t.Fatalf("failed to seed tenant A order: %v", err)
	}

	// Seed tenant A items (will be deleted and recreated)
	itemA := models.TiktokEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      tenantA,
		EscrowOrderID: orderA.ID,
		OrderID:       "ORD-TA-001",
		Quantity:      2,
		SalePrice:     20000,
		OriginalPrice: 20000,
	}
	if err := db.Create(&itemA).Error; err != nil {
		t.Fatalf("failed to seed tenant A item: %v", err)
	}

	// Seed tenant B order with DIFFERENT order_id (no raw data overlap)
	orderB := models.TiktokEscrowOrder{
		ID:       uuid.New().String(),
		TenantID: tenantB,
		OrderID:  "ORD-TB-001",
		Month:    5,
		Year:     2026,
		// No raw_order_data — tenant B has no items to restore from
	}
	if err := db.Create(&orderB).Error; err != nil {
		t.Fatalf("failed to seed tenant B order: %v", err)
	}

	// Seed tenant B items
	itemB := models.TiktokEscrowItem{
		ID:            uuid.New().String(),
		TenantID:      tenantB,
		EscrowOrderID: orderB.ID,
		OrderID:       "ORD-TB-001",
		Quantity:      3,
		SalePrice:     30000,
		OriginalPrice: 30000,
	}
	if err := db.Create(&itemB).Error; err != nil {
		t.Fatalf("failed to seed tenant B item: %v", err)
	}

	// Record pre-repopulate counts
	var tenantBCountBefore int64
	db.Model(&models.TiktokEscrowItem{}).Where("tenant_id = ?", tenantB).Count(&tenantBCountBefore)

	// Run repopulate as tenant B — should only affect tenant B's orders with raw data
	// Tenant B has no raw data, so this should be a no-op for tenant B
	if err := svc.RepopulateItems(ctx, tenantB, "2026-05"); err != nil {
		t.Fatalf("RepopulateItems for tenant B returned error: %v", err)
	}

	// Tenant B items must remain unchanged (there's no raw data to reconstruct from)
	var tenantBCountAfter int64
	db.Model(&models.TiktokEscrowItem{}).Where("tenant_id = ?", tenantB).Count(&tenantBCountAfter)
	if tenantBCountAfter != tenantBCountBefore {
		t.Errorf("tenant B item count changed from %d to %d (should be unchanged)", tenantBCountBefore, tenantBCountAfter)
	}

	// Tenant A items must remain completely untouched
	var tenantACount int64
	db.Model(&models.TiktokEscrowItem{}).
		Where("tenant_id = ? AND escrow_order_id = ?", tenantA, orderA.ID).
		Count(&tenantACount)
	if tenantACount != 1 {
		t.Errorf("tenant A item count changed to %d (should be 1, untouched)", tenantACount)
	}
}

// containsSubstring is a helper for error assertions
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}