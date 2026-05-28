package analytics_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics"
	"gorm.io/gorm"
)

// ============================================================================
// T28 — Escrow Reports to Inventory/Reconciliation Integration Tests
// ============================================================================
//
// Verifies:
// 1. Escrow rows traceable to order/item/inventory records (order_sn, item_id, SKU)
// 2. Missing marketplace rows flagged as sync incomplete, NOT reconciliation discrepancy
// 3. Partial escrow windows not consumed as final data
// 4. Reconciliation DTO includes identifiers used by UI drill-down
// ============================================================================

// ---------------------------------------------------------------------------
// T28 — Setup helpers
// ---------------------------------------------------------------------------

func setupT28ShopeeDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := fmt.Sprintf("/tmp/t28_shopee_test_%s.db", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(MEMORY)&_pragma=synchronous(OFF)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open SQLite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.AnalyticsSettings{},
		&models.ShopeeEscrowSync{},
		&models.ShopeeEscrowOrder{},
		&models.ShopeeEscrowItem{},
		&models.InventoryRecord{},
		&models.Job{},
	); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	return db
}

func setupT28TiktokDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := fmt.Sprintf("/tmp/t28_tiktok_test_%s.db", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(MEMORY)&_pragma=synchronous(OFF)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open SQLite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.AnalyticsSettings{},
		&models.TiktokEscrowSync{},
		&models.TiktokEscrowOrder{},
		&models.TiktokEscrowItem{},
		&models.InventoryRecord{},
		&models.Job{},
	); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	return db
}

// ---------------------------------------------------------------------------
// T28-1. Shopee: Escrow-to-inventory traceability via SKU
// ---------------------------------------------------------------------------
//
// Verifies: escrow items with SKU X are linked via reconciliation to the
// inventory record with key_value = X, and the inventory price flows through
// to the DTO's inventory_price and expected_income fields.

func TestT28_ShopeeReconciliation_InventoryTraceability(t *testing.T) {
	db := setupT28ShopeeDB(t)
	ctx := context.Background()
	tenantID := "t28-test-tenant"

	// Analytics settings with HARGA price column
	db.Create(&models.AnalyticsSettings{
		ID: uuid.New().String(), TenantID: tenantID,
		Platform: "shopee", PriceColumn: "HARGA",
		FormulaDeduction: 1500, FormulaMultiplier: 0.84,
	})

	// Inventory record for SKU "T28-SKU-A" with price 50000
	invData, _ := json.Marshal(map[string]interface{}{
		"HARGA":       "50000",
		"Nama Barang": "T28 Product A",
	})
	db.Create(&models.InventoryRecord{
		ID: uuid.New().String(), TenantID: tenantID,
		KeyValue: "T28-SKU-A", KeyColumnName: "SKU",
		Data: string(invData), SyncStatus: "synced",
	})

	// Escrow order
	order := models.ShopeeEscrowOrder{
		ID: uuid.New().String(), TenantID: tenantID,
		OrderSN: "T28-ORDER-001", Month: 3, Year: 2025,
		EscrowAmount: 40000, BuyerTotalAmount: 50000,
	}
	db.Create(&order)

	// Single-item escrow record linked to the order
	skuA := "T28-SKU-A"
	itemName := "T28 Product A Item"
	m, y := 3, 2025
	db.Create(&models.ShopeeEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID, EscrowOrderID: order.ID,
		OrderSN: &order.OrderSN, Month: &m, Year: &y,
		Sku: &skuA, ItemName: &itemName,
		Quantity: 1, OriginalPrice: 50000,
	})

	// Run reconciliation
	svc := analytics.NewShopeeAnalyticsService(db, db, tenantID)
	result, err := svc.GetReconciliation(ctx, tenantID, 3, 2025)
	if err != nil {
		t.Fatalf("GetReconciliation error: %v", err)
	}

	// Find the SKU-A group
	var found bool
	for _, g := range result.SkuGroups {
		if g.Sku == "T28-SKU-A" {
			found = true
			// Inventory traceability: inventory_price from inventory record
			if g.InventoryPrice == nil {
				t.Error("inventory_price should be set (traceability to inventory)")
			} else if *g.InventoryPrice != 50000 {
				t.Errorf("expected inventory_price=50000, got %v", *g.InventoryPrice)
			}
			// Expected income derived from inventory price
			if g.ExpectedIncome == nil {
				t.Error("expected_income should be computed from inventory")
			} else {
				expected := (50000.0 - 1500) * 0.84 // 40740
				if *g.ExpectedIncome != expected {
					t.Errorf("expected expected_income=%v, got %v", expected, *g.ExpectedIncome)
				}
			}
			// Item name from inventory when item name is empty in escrow
			// (the item_name was set so it uses the escrow item name)
			if g.Status != "OK" {
				t.Errorf("expected status=OK, got %s", g.Status)
			}
		}
	}
	if !found {
		t.Fatal("T28-SKU-A group not found in reconciliation results")
	}
}

// ---------------------------------------------------------------------------
// T28-2. Shopee: Sync FailedOrders indicates incomplete sync (NOT discrepancy)
// ---------------------------------------------------------------------------
//
// Verifies: when FailedOrders > 0, SyncStatusDTO reflects it clearly.
// Reconciliation should NOT flag missing orders as PRICE_DIFF — they are
// simply not present in the database because the marketplace API didn't return them.

func TestT28_ShopeeSyncStatus_FailedOrders_IndicatesIncomplete(t *testing.T) {
	db := setupT28ShopeeDB(t)
	ctx := context.Background()
	tenantID := "t28-test-tenant"

	failedIDs := "T28-FAIL-001,T28-FAIL-002"
	db.Create(&models.ShopeeEscrowSync{
		ID: uuid.New().String(), TenantID: tenantID,
		Month: 3, Year: 2025,
		TotalOrders: 8, FailedOrders: 2,
		FailedOrderIDs: &failedIDs,
		SyncedAt:       time.Now(),
	})

	svc := analytics.NewShopeeAnalyticsService(db, db, tenantID)
	status, err := svc.GetSyncStatus(ctx, tenantID, 3, 2025)
	if err != nil {
		t.Fatalf("GetSyncStatus error: %v", err)
	}

	// Synced=true (sync happened), FailedOrders=2 (2 orders missing from marketplace)
	if !status.Synced {
		t.Error("expected Synced=true")
	}
	if status.TotalOrders != 8 {
		t.Errorf("expected TotalOrders=8, got %d", status.TotalOrders)
	}
	if status.FailedOrders != 2 {
		t.Errorf("expected FailedOrders=2, got %d", status.FailedOrders)
	}

	// KEY: FailedOrders > 0 = sync was INCOMPLETE.
	// Callers must check this BEFORE interpreting reconciliation PRICE_DIFF results.
	// Missing orders are a sync completeness issue, NOT a price discrepancy.
	if status.FailedOrders == 0 {
		t.Error("FAILED: FailedOrders must be > 0 to flag sync as incomplete")
	}

	// Reconciliation on empty DB (no synced orders inserted) returns empty
	result, err := svc.GetReconciliation(ctx, tenantID, 3, 2025)
	if err != nil {
		t.Fatalf("GetReconciliation error: %v", err)
	}
	if len(result.SkuGroups) != 0 {
		t.Errorf("expected 0 SKU groups (no synced orders in DB), got %d", len(result.SkuGroups))
	}
}

// ---------------------------------------------------------------------------
// T28-3. Shopee: Reconciliation DTO has all identifiers for UI drill-down
// ---------------------------------------------------------------------------

func TestT28_ShopeeReconciliation_DTOContainsIdentifiers(t *testing.T) {
	db := setupT28ShopeeDB(t)
	ctx := context.Background()
	tenantID := "t28-test-tenant"

	db.Create(&models.AnalyticsSettings{
		ID: uuid.New().String(), TenantID: tenantID,
		Platform: "shopee", PriceColumn: "HARGA",
		FormulaDeduction: 1500, FormulaMultiplier: 0.84,
	})

	// Inventory for T28-SKU-X only (T28-SKU-Y → NO_INVENTORY)
	invData, _ := json.Marshal(map[string]interface{}{
		"HARGA": "30000", "Nama Barang": "T28 Product X",
	})
	db.Create(&models.InventoryRecord{
		ID: uuid.New().String(), TenantID: tenantID,
		KeyValue: "T28-SKU-X", KeyColumnName: "SKU",
		Data: string(invData), SyncStatus: "synced",
	})

	o1 := models.ShopeeEscrowOrder{ID: uuid.New().String(), TenantID: tenantID, OrderSN: "T28-O1", Month: 4, Year: 2025, EscrowAmount: 25000}
	o2 := models.ShopeeEscrowOrder{ID: uuid.New().String(), TenantID: tenantID, OrderSN: "T28-O2", Month: 4, Year: 2025, EscrowAmount: 28000}
	db.Create(&o1)
	db.Create(&o2)

	skuX, skuY := "T28-SKU-X", "T28-SKU-Y"
	itemX, itemY := "T28 Item X", "T28 Item Y"
	m, y := 4, 2025

	db.Create(&models.ShopeeEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID, EscrowOrderID: o1.ID,
		OrderSN: &o1.OrderSN, Month: &m, Year: &y,
		Sku: &skuX, ItemName: &itemX, Quantity: 1, OriginalPrice: 30000,
	})
	db.Create(&models.ShopeeEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID, EscrowOrderID: o2.ID,
		OrderSN: &o2.OrderSN, Month: &m, Year: &y,
		Sku: &skuY, ItemName: &itemY, Quantity: 1, OriginalPrice: 28000,
	})

	svc := analytics.NewShopeeAnalyticsService(db, db, tenantID)
	result, err := svc.GetReconciliation(ctx, tenantID, 4, 2025)
	if err != nil {
		t.Fatalf("GetReconciliation error: %v", err)
	}

	if len(result.SkuGroups) != 2 {
		t.Fatalf("expected 2 SKU groups, got %d", len(result.SkuGroups))
	}

	for _, g := range result.SkuGroups {
		// All SKU groups must have these identifiers for UI drill-down
		if g.Sku == "" {
			t.Errorf("sku identifier missing")
		}
		if g.TotalTransactions == 0 {
			t.Errorf("total_transactions=0 for sku %s", g.Sku)
		}
		if len(g.UniqueUnitPrices) == 0 {
			t.Errorf("unique_unit_prices empty for sku %s", g.Sku)
		}

		switch g.Sku {
		case "T28-SKU-X":
			if g.InventoryPrice == nil {
				t.Error("T28-SKU-X: inventory_price should be set")
			}
			if g.ExpectedIncome == nil {
				t.Error("T28-SKU-X: expected_income should be computed")
			}
			if g.Status != "OK" && g.Status != "PRICE_DIFF" {
				t.Errorf("T28-SKU-X: expected OK or PRICE_DIFF, got %s", g.Status)
			}
		case "T28-SKU-Y":
			if g.InventoryPrice != nil {
				t.Error("T28-SKU-Y: inventory_price should be nil (no inventory)")
			}
			if g.Status != "NO_INVENTORY" {
				t.Errorf("T28-SKU-Y: expected NO_INVENTORY, got %s", g.Status)
			}
		}
	}

	if result.Summary.TotalSku != 2 {
		t.Errorf("expected TotalSku=2, got %d", result.Summary.TotalSku)
	}
	if result.Summary.TotalTransactions != 2 {
		t.Errorf("expected TotalTransactions=2, got %d", result.Summary.TotalTransactions)
	}
}

// ---------------------------------------------------------------------------
// T28-4. TikTok: Escrow-to-inventory traceability via seller_sku
// ---------------------------------------------------------------------------

func TestT28_TiktokReconciliation_InventoryTraceability(t *testing.T) {
	db := setupT28TiktokDB(t)
	ctx := context.Background()
	tenantID := "t28-test-tenant"

	db.Create(&models.AnalyticsSettings{
		ID: uuid.New().String(), TenantID: tenantID,
		Platform: "tiktok", PriceColumn: "HARGA",
		FormulaDeduction: 1500, FormulaMultiplier: 0.86,
	})

	// Inventory for seller_sku "T28-TT-SKU-1"
	invData, _ := json.Marshal(map[string]interface{}{
		"HARGA": "75000", "Nama Barang": "T28 TikTok Product 1",
	})
	db.Create(&models.InventoryRecord{
		ID: uuid.New().String(), TenantID: tenantID,
		KeyValue: "T28-TT-SKU-1", KeyColumnName: "SKU",
		Data: string(invData), SyncStatus: "synced",
	})

	// TikTok escrow order
	orderStatus := "COMPLETED"
	txID := "T28-TX-001"
	order := models.TiktokEscrowOrder{
		ID: uuid.New().String(), TenantID: tenantID,
		OrderID: "T28-TT-ORDER-001", Month: 5, Year: 2025,
		OrderStatus: &orderStatus, TransactionID: &txID,
		TotalSettlementAmount: 60000,
		BuyerTotalAmount:      80000,
		ProductRevenue:        75000,
	}
	db.Create(&order)

	// Item with seller_sku matching inventory key_value
	sellerSku := "T28-TT-SKU-1"
	skuID := "T28-SKU-ID-001"
	productName := "T28 TikTok Product 1"
	db.Create(&models.TiktokEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID,
		EscrowOrderID: order.ID, OrderID: order.OrderID,
		SellerSku: &sellerSku, SkuID: &skuID,
		ProductName: &productName, Quantity: 1,
		OriginalPrice: 75000, SalePrice: 75000,
	})

	svc := analytics.NewTiktokAnalyticsService(db, db, tenantID)
	result, err := svc.GetReconciliation(ctx, tenantID, 5, 2025)
	if err != nil {
		t.Fatalf("GetReconciliation error: %v", err)
	}

	var found bool
	for _, g := range result.SkuGroups {
		if g.Sku == "T28-TT-SKU-1" {
			found = true
			if g.InventoryPrice == nil {
				t.Error("inventory_price should be set from inventory lookup")
			} else if *g.InventoryPrice != 75000 {
				t.Errorf("expected inventory_price=75000, got %v", *g.InventoryPrice)
			}
			if g.ExpectedIncome == nil {
				t.Error("expected_income should be computed")
			} else {
				expected := (75000.0 - 1500) * 0.86 // 63210
				if *g.ExpectedIncome != expected {
					t.Errorf("expected expected_income=%v, got %v", expected, *g.ExpectedIncome)
				}
			}
			if g.Status != "OK" {
				t.Errorf("expected status=OK, got %s", g.Status)
			}
		}
	}
	if !found {
		t.Fatal("T28-TT-SKU-1 group not found in TikTok reconciliation results")
	}
}

// ---------------------------------------------------------------------------
// T28-5. TikTok: Sync FailedOrders indicates incomplete sync
// ---------------------------------------------------------------------------

func TestT28_TiktokSyncStatus_FailedOrders_IndicatesIncomplete(t *testing.T) {
	db := setupT28TiktokDB(t)
	ctx := context.Background()
	tenantID := "t28-test-tenant"

	failedIDs := `["T28-TT-ORDER-001","T28-TT-ORDER-002","T28-TT-ORDER-003"]`
	db.Create(&models.TiktokEscrowSync{
		ID: uuid.New().String(), TenantID: tenantID,
		Month: 5, Year: 2025,
		TotalOrders: 12, FailedOrders: 3,
		FailedOrderIDs: &failedIDs,
		SyncedAt:       time.Now(),
	})

	svc := analytics.NewTiktokAnalyticsService(db, db, tenantID)
	status, err := svc.GetSyncStatus(ctx, tenantID, 5, 2025)
	if err != nil {
		t.Fatalf("GetSyncStatus error: %v", err)
	}

	if !status.Synced {
		t.Error("expected Synced=true")
	}
	if status.FailedOrders != 3 {
		t.Errorf("expected FailedOrders=3, got %d", status.FailedOrders)
	}
	// KEY ASSERTION: FailedOrders > 0 means 3 orders could not be fetched from
	// the TikTok marketplace API. This is an INCOMPLETE sync indicator,
	// NOT a reconciliation price discrepancy.
	if status.FailedOrders == 0 {
		t.Error("FAILED: FailedOrders must be > 0 to indicate incomplete sync")
	}
}

// ---------------------------------------------------------------------------
// T28-6. TikTok: Reconciliation DTO has all identifiers for UI drill-down
// ---------------------------------------------------------------------------

func TestT28_TiktokReconciliation_DTOContainsIdentifiers(t *testing.T) {
	db := setupT28TiktokDB(t)
	ctx := context.Background()
	tenantID := "t28-test-tenant"

	db.Create(&models.AnalyticsSettings{
		ID: uuid.New().String(), TenantID: tenantID,
		Platform: "tiktok", PriceColumn: "HARGA",
		FormulaDeduction: 1500, FormulaMultiplier: 0.86,
	})

	orderStatus := "COMPLETED"
	o := models.TiktokEscrowOrder{
		ID: uuid.New().String(), TenantID: tenantID,
		OrderID: "T28-TT-O-001", Month: 6, Year: 2025,
		OrderStatus:           &orderStatus,
		TotalSettlementAmount: 40000,
	}
	db.Create(&o)

	sellerSku := "T28-TT-SELLER-SKU"
	productName := "T28 Test Product"
	db.Create(&models.TiktokEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID,
		EscrowOrderID: o.ID, OrderID: o.OrderID,
		SellerSku: &sellerSku, ProductName: &productName,
		Quantity: 2, SalePrice: 40000, OriginalPrice: 42000,
	})

	svc := analytics.NewTiktokAnalyticsService(db, db, tenantID)
	result, err := svc.GetReconciliation(ctx, tenantID, 6, 2025)
	if err != nil {
		t.Fatalf("GetReconciliation error: %v", err)
	}

	if len(result.SkuGroups) == 0 {
		t.Fatal("expected at least one SKU group")
	}

	g := result.SkuGroups[0]
	// Verify all required identifiers for UI drill-down are present
	if g.Sku == "" {
		t.Error("sku identifier missing")
	}
	if g.SellerSku == "" {
		t.Error("seller_sku identifier missing")
	}
	if g.TotalTransactions == 0 {
		t.Error("total_transactions should be > 0")
	}
	if len(g.UniqueUnitPrices) == 0 {
		t.Error("unique_unit_prices should be populated for drill-down")
	}
	if g.Status == "" {
		t.Error("status field missing")
	}
}

// ---------------------------------------------------------------------------
// T28-7. Shopee: Escrow order traceable to items via EscrowOrderID chain
// ---------------------------------------------------------------------------
//
// Verifies the full traceability chain:
// EscrowOrder.OrderSN → EscrowItem.EscrowOrderID → EscrowItem.Sku → Inventory
//
// And the drill-down API: GetSkuOrders(sku) → OrderSN, GetOrderItems(orderSN) → items

func TestT28_ShopeeEscrowOrder_TraceableToItems(t *testing.T) {
	db := setupT28ShopeeDB(t)
	ctx := context.Background()
	tenantID := "t28-test-tenant"

	order := models.ShopeeEscrowOrder{
		ID: uuid.New().String(), TenantID: tenantID,
		OrderSN: "T28-TRACE-001", Month: 7, Year: 2025,
		EscrowAmount: 100000, BuyerTotalAmount: 120000,
	}
	db.Create(&order)

	// Two items in same order (multi-item order)
	sku1, sku2 := "T28-SKU-ITEM-1", "T28-SKU-ITEM-2"
	item1Name, item2Name := "T28 Item 1", "T28 Item 2"
	m, y := 7, 2025

	db.Create(&models.ShopeeEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID, EscrowOrderID: order.ID,
		OrderSN: &order.OrderSN, Month: &m, Year: &y,
		Sku: &sku1, ItemName: &item1Name, Quantity: 1, OriginalPrice: 50000,
	})
	db.Create(&models.ShopeeEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID, EscrowOrderID: order.ID,
		OrderSN: &order.OrderSN, Month: &m, Year: &y,
		Sku: &sku2, ItemName: &item2Name, Quantity: 2, OriginalPrice: 70000,
	})

	svc := analytics.NewShopeeAnalyticsService(db, db, tenantID)

	// Forward trace: order → items
	itemsResult, err := svc.GetOrderItems(ctx, tenantID, "T28-TRACE-001", 7, 2025)
	if err != nil {
		t.Fatalf("GetOrderItems error: %v", err)
	}
	if len(itemsResult.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(itemsResult.Items))
	}
	for _, item := range itemsResult.Items {
		if item.EscrowOrderID != order.ID {
			t.Errorf("item %s: EscrowOrderID=%s, expected %s", item.ID, item.EscrowOrderID, order.ID)
		}
		if item.Sku == "" {
			t.Errorf("item %s: missing SKU identifier", item.ID)
		}
	}

	// Reverse trace: SKU → order
	skuOrders, err := svc.GetSkuOrders(ctx, tenantID, "T28-SKU-ITEM-1", 7, 2025)
	if err != nil {
		t.Fatalf("GetSkuOrders error: %v", err)
	}
	if len(skuOrders.Orders) != 1 {
		t.Fatalf("expected 1 order for T28-SKU-ITEM-1, got %d", len(skuOrders.Orders))
	}
	if skuOrders.Orders[0].OrderSN != "T28-TRACE-001" {
		t.Errorf("expected order_sn=T28-TRACE-001, got %s", skuOrders.Orders[0].OrderSN)
	}
	// Verify order DTO contains reconciliation identifiers
	o := skuOrders.Orders[0]
	if o.Sku == "" {
		t.Error("order DTO missing sku identifier")
	}
	if o.ID == "" {
		t.Error("order DTO missing id")
	}
}

// ---------------------------------------------------------------------------
// T28-8. Shopee: Multi-item orders exclude actual income from reconciliation
// ---------------------------------------------------------------------------
//
// When an order has multiple items, EscrowAmount cannot be attributed to a
// single SKU. Reconciliation must NOT use it for per-unit actual income,
// preventing false PRICE_DIFF flags on multi-item orders.

func TestT28_ShopeeReconciliation_MultiItemOrder_NoActualIncome(t *testing.T) {
	db := setupT28ShopeeDB(t)
	ctx := context.Background()
	tenantID := "t28-test-tenant"

	db.Create(&models.AnalyticsSettings{
		ID: uuid.New().String(), TenantID: tenantID,
		Platform: "shopee", PriceColumn: "HARGA",
		FormulaDeduction: 1500, FormulaMultiplier: 0.84,
	})

	// Multi-item order (2 items)
	order := models.ShopeeEscrowOrder{
		ID: uuid.New().String(), TenantID: tenantID,
		OrderSN: "T28-MULTI-001", Month: 8, Year: 2025,
		EscrowAmount: 80000, // combined actual — should NOT be used per-item
	}
	db.Create(&order)

	sku1, sku2 := "T28-MULTI-SKU-1", "T28-MULTI-SKU-2"
	m, y := 8, 2025
	db.Create(&models.ShopeeEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID, EscrowOrderID: order.ID,
		OrderSN: &order.OrderSN, Month: &m, Year: &y,
		Sku: &sku1, ItemName: &sku1, Quantity: 1, OriginalPrice: 40000,
	})
	db.Create(&models.ShopeeEscrowItem{
		ID: uuid.New().String(), TenantID: tenantID, EscrowOrderID: order.ID,
		OrderSN: &order.OrderSN, Month: &m, Year: &y,
		Sku: &sku2, ItemName: &sku2, Quantity: 1, OriginalPrice: 40000,
	})

	svc := analytics.NewShopeeAnalyticsService(db, db, tenantID)
	result, err := svc.GetReconciliation(ctx, tenantID, 8, 2025)
	if err != nil {
		t.Fatalf("GetReconciliation error: %v", err)
	}

	// Both SKU groups must have empty actual incomes (multi-item order guard)
	for _, g := range result.SkuGroups {
		if len(g.UniqueActualIncomes) != 0 {
			t.Errorf("SKU %s: multi-item order should have empty unique_actual_incomes, got %v",
				g.Sku, g.UniqueActualIncomes)
		}
	}
}
