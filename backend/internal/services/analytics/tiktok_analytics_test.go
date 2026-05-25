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

func setupTiktokTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite: %v", err)
	}

	// Auto-migrate the tables we need
	if err := db.AutoMigrate(
		&models.AnalyticsSettings{},
		&models.TiktokEscrowSync{},
		&models.TiktokEscrowOrder{},
		&models.TiktokEscrowItem{},
		&models.Job{},
	); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}

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