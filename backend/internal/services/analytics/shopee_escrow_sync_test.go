package analytics_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// fakeShopeeClient returns deterministic test data for Shopee escrow sync tests
// ---------------------------------------------------------------------------

type fakeShopeeClient struct {
	t               *testing.T
	walletTxCount   int           // how many wallet transactions to return
	escrowOrderFn   func(orderSN string) *shopeePkg.EscrowOrder // customize per order
}

func (f *fakeShopeeClient) GetWalletTransactions(req shopeePkg.GetWalletTransactionRequest) (*shopeePkg.WalletTransactionResponse, error) {
	f.t.Helper()
	if req.TransactionTabType != "wallet_order_income" {
		f.t.Fatalf("expected wallet_order_income tab, got %s", req.TransactionTabType)
	}
	if req.MoneyFlow != "MONEY_IN" {
		f.t.Fatalf("expected MONEY_IN money_flow, got %s", req.MoneyFlow)
	}

	var list []shopeePkg.WalletTransaction
	for i := 0; i < f.walletTxCount; i++ {
		list = append(list, shopeePkg.WalletTransaction{
			TransactionID:   int64(1000 + i),
			Status:          "SUCCESS",
			TransactionType: "wallet_order_income",
			Amount:          100000,
			CreateTime:      time.Now().Unix(),
			OrderSN:         fmt.Sprintf("FAKE-ORDER-%03d", i+1),
			BuyerUsername:   "Fake Buyer",
		})
	}
	return &shopeePkg.WalletTransactionResponse{
		Response: struct {
			TransactionList []shopeePkg.WalletTransaction `json:"transaction_list"`
			More            bool                           `json:"more"`
		}{
			TransactionList: list,
			More:            false,
		},
	}, nil
}

func (f *fakeShopeeClient) GetEscrowDetails(req shopeePkg.GetEscrowDetailsRequest) (*shopeePkg.GetEscrowDetailsResponse, error) {
	f.t.Helper()
	wrappers := make([]shopeePkg.EscrowDetailWrapper, 0, len(req.OrderSNList))
	for _, sn := range req.OrderSNList {
		order := f.defaultEscrowOrder(sn)
		if f.escrowOrderFn != nil {
			if custom := f.escrowOrderFn(sn); custom != nil {
				order = *custom
			}
		}
		wrappers = append(wrappers, shopeePkg.EscrowDetailWrapper{
			EscrowDetail: &order,
		})
	}
	return &shopeePkg.GetEscrowDetailsResponse{
		Response: wrappers,
	}, nil
}

func (f *fakeShopeeClient) defaultEscrowOrder(orderSN string) shopeePkg.EscrowOrder {
	return shopeePkg.EscrowOrder{
		OrderSN:       orderSN,
		BuyerUsername: "Fake Buyer",
		TotalAmount:   100000,
		OrderIncome: shopeePkg.EscrowOrderData{
			EscrowAmount:              100000,
			CommissionFee:             5000,
			ServiceFee:                1000,
			SellerOrderProcessingFee:  500,
			BuyerPaidShippingFee:      10000,
			ActualShippingFee:         7000,
			ShopeeShippingRebate:      1500,
			EstimatedShippingFee:      8000,
			BuyerTotalAmount:          120000,
			BuyerPaymentMethod:        "credit_card",
			Items: []shopeePkg.EscrowItemData{
				{
					ItemID:            1,
					ModelID:           10,
					ItemName:          "Test Item",
					ModelName:         "Test Model",
					ItemSKU:           "SKU-TEST",
					ModelSKU:          "MODEL-SKU-TEST",
					QuantityPurchased: 2,
					OriginalPrice:     50000,
					SellingPrice:      50000,
					DiscountedPrice:   45000,
				},
			},
		},
		BuyerPaymentInfo: map[string]any{
			"payment_method": "credit_card",
		},
	}
}

// ---------------------------------------------------------------------------
// setupSyncTestDB creates an in-memory SQLite DB for sync service tests
// ---------------------------------------------------------------------------

func setupSyncTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open in-memory SQLite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.ShopeeEscrowSync{},
		&models.ShopeeEscrowOrder{},
		&models.ShopeeEscrowItem{},
	); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
	return db
}

func resetSyncTestDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	db.Exec("DELETE FROM " + (&models.ShopeeEscrowSync{}).TableName())
	db.Exec("DELETE FROM " + (&models.ShopeeEscrowOrder{}).TableName())
	db.Exec("DELETE FROM " + (&models.ShopeeEscrowItem{}).TableName())
}

// ---------------------------------------------------------------------------
// Persistence test — verifies real data makes it into the DB
// ---------------------------------------------------------------------------

func TestShopeeSync_PersistsOrdersItemsAndRawJSON(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "sync-persist-tenant")
	svc.SetShopeeClient(&fakeShopeeClient{
		t:             t,
		walletTxCount: 3,
	})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	if err := svc.SyncMonthWithProgress(ctx, 4, 2026, false, onProgress); err != nil {
		t.Fatalf("SyncMonthWithProgress returned error: %v", err)
	}

	tenantID := "sync-persist-tenant"
	month := 4
	year := 2026

	t.Run("sync_record_has_total_orders", func(t *testing.T) {
		var sync models.ShopeeEscrowSync
		if err := db.Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
			First(&sync).Error; err != nil {
			t.Fatalf("sync record not found: %v", err)
		}
		if sync.TotalOrders == 0 {
			t.Errorf("sync record TotalOrders = 0, want > 0")
		}
		if sync.SyncedAt.IsZero() {
			t.Errorf("sync record SyncedAt is zero")
		}
	})

	t.Run("orders_persisted", func(t *testing.T) {
		var count int64
		db.Model(&models.ShopeeEscrowOrder{}).
			Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
			Count(&count)
		if count == 0 {
			t.Error("no orders persisted, want > 0")
		}
	})

	t.Run("items_persisted", func(t *testing.T) {
		var count int64
		db.Model(&models.ShopeeEscrowItem{}).
			Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
			Count(&count)
		if count == 0 {
			t.Error("no items persisted, want > 0")
		}
	})

	t.Run("raw_json_persisted", func(t *testing.T) {
		var rawCount int64
		db.Model(&models.ShopeeEscrowOrder{}).
			Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
			Where("(raw_order_income IS NOT NULL AND raw_order_income != '') OR (raw_buyer_payment_info IS NOT NULL AND raw_buyer_payment_info != '')").
			Count(&rawCount)
		if rawCount == 0 {
			t.Error("no raw JSON persisted (raw_order_income or raw_buyer_payment_info)")
		}
	})

	t.Run("order_fields_populated", func(t *testing.T) {
		var orders []models.ShopeeEscrowOrder
		db.Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
			Find(&orders)
		for _, o := range orders {
			if o.EscrowAmount == 0 {
				t.Errorf("order %s: EscrowAmount is 0", o.OrderSN)
			}
			if o.CommissionFee == 0 {
				t.Errorf("order %s: CommissionFee is 0", o.OrderSN)
			}
			if o.BuyerUserName == nil || *o.BuyerUserName == "" {
				t.Errorf("order %s: BuyerUserName is nil/empty", o.OrderSN)
			}
		}
	})

	t.Run("item_fields_populated", func(t *testing.T) {
		var items []models.ShopeeEscrowItem
		db.Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
			Find(&items)
		for _, it := range items {
			if it.ItemName == nil || *it.ItemName == "" {
				t.Errorf("item %s: ItemName is nil/empty", it.ID)
			}
			if it.Quantity == 0 {
				t.Errorf("item %s: Quantity is 0", it.ID)
			}
			if it.RawItemData == nil || *it.RawItemData == "" {
				t.Errorf("item %s: RawItemData is nil/empty", it.ID)
			}
		}
	})
}

// ---------------------------------------------------------------------------
// Idempotency test — running same sync twice does not duplicate data
// ---------------------------------------------------------------------------

func TestShopeeSync_IdempotentRerun(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "idempotent-tenant")
	svc.SetShopeeClient(&fakeShopeeClient{
		t:             t,
		walletTxCount: 2,
	})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}
	tenantID := "idempotent-tenant"
	month := 5
	year := 2026

	if err := svc.SyncMonthWithProgress(ctx, month, year, false, onProgress); err != nil {
		t.Fatalf("first sync returned error: %v", err)
	}

	var orderCount1, itemCount1 int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&orderCount1)
	db.Model(&models.ShopeeEscrowItem{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&itemCount1)

	if orderCount1 == 0 {
		t.Fatal("first sync: no orders persisted")
	}

	// Second sync — with no forceResync, should detect "already synced"
	if err := svc.SyncMonthWithProgress(ctx, month, year, false, onProgress); err != nil {
		t.Fatalf("second sync (no force) returned error: %v", err)
	}

	var orderCount2, itemCount2 int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&orderCount2)
	db.Model(&models.ShopeeEscrowItem{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&itemCount2)

	if orderCount2 != orderCount1 {
		t.Errorf("order count changed after idempotent rerun: before=%d after=%d", orderCount1, orderCount2)
	}
	if itemCount2 != itemCount1 {
		t.Errorf("item count changed after idempotent rerun: before=%d after=%d", itemCount1, itemCount2)
	}

	// Only one sync record should exist
	var syncCount int64
	db.Model(&models.ShopeeEscrowSync{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&syncCount)
	if syncCount != 1 {
		t.Errorf("expected 1 sync record, got %d", syncCount)
	}
}

// ---------------------------------------------------------------------------
// Force resync — replaces data without duplicating
// ---------------------------------------------------------------------------

func TestShopeeSync_ForceResyncReplacesData(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "force-resync-tenant")
	svc.SetShopeeClient(&fakeShopeeClient{
		t:             t,
		walletTxCount: 2,
	})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}
	tenantID := "force-resync-tenant"
	month := 6
	year := 2026

	// First sync
	if err := svc.SyncMonthWithProgress(ctx, month, year, false, onProgress); err != nil {
		t.Fatalf("first sync returned error: %v", err)
	}

	// Force resync
	if err := svc.SyncMonthWithProgress(ctx, month, year, true, onProgress); err != nil {
		t.Fatalf("force resync returned error: %v", err)
	}

	var orderCount, itemCount int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&orderCount)
	db.Model(&models.ShopeeEscrowItem{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&itemCount)

	if orderCount == 0 {
		t.Error("force resync: no orders persisted")
	}
	if itemCount == 0 {
		t.Error("force resync: no items persisted")
	}

	// After force resync, only one sync record should exist
	var syncCount int64
	db.Model(&models.ShopeeEscrowSync{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&syncCount)
	if syncCount != 1 {
		t.Errorf("expected 1 sync record after force resync, got %d", syncCount)
	}
}

// ---------------------------------------------------------------------------
// Tenant isolation — cross-tenant data is not visible
// ---------------------------------------------------------------------------

func TestShopeeSync_TenantIsolation(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svcA := analytics.NewShopeeEscrowSyncService(db, db, "tenant-a")
	svcA.SetShopeeClient(&fakeShopeeClient{
		t:             t,
		walletTxCount: 2,
	})
	svcB := analytics.NewShopeeEscrowSyncService(db, db, "tenant-b")
	svcB.SetShopeeClient(&fakeShopeeClient{
		t:             t,
		walletTxCount: 3,
	})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	if err := svcA.SyncMonthWithProgress(ctx, 7, 2026, false, onProgress); err != nil {
		t.Fatalf("tenant A sync error: %v", err)
	}
	if err := svcB.SyncMonthWithProgress(ctx, 7, 2026, false, onProgress); err != nil {
		t.Fatalf("tenant B sync error: %v", err)
	}

	var countA, countB int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ?", "tenant-a").
		Count(&countA)
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ?", "tenant-b").
		Count(&countB)

	if countA == 0 {
		t.Error("tenant A has no orders")
	}
	if countB == 0 {
		t.Error("tenant B has no orders")
	}

	var crossTenantCount int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ?", "tenant-a").
		Where("order_sn LIKE ?", "FAKE-ORDER-%%").
		Not("order_sn LIKE ?", "FAKE-ORDER-00%").
		Count(&crossTenantCount)
	// tenant A had 2 orders (FAKE-ORDER-001, FAKE-ORDER-002)
	// tenant B had 3 orders (FAKE-ORDER-001, FAKE-ORDER-002, FAKE-ORDER-003)
	// verifying tenant A cannot see tenant B's third order

	var tenantBOnlyOrder int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ?", "tenant-a").
		Where("order_sn = ?", "FAKE-ORDER-003").
		Count(&tenantBOnlyOrder)
	if tenantBOnlyOrder > 0 {
		t.Errorf("tenant A can see tenant B's order (FAKE-ORDER-003)")
	}
}

// ---------------------------------------------------------------------------
// Wallet deduplication — same OrderSN+Amount consumed once
// ---------------------------------------------------------------------------

type fakeShopeeClientWithDupes struct {
	t *testing.T
}

func (f *fakeShopeeClientWithDupes) GetWalletTransactions(req shopeePkg.GetWalletTransactionRequest) (*shopeePkg.WalletTransactionResponse, error) {
	f.t.Helper()
	// Return 4 transactions, but 2 have duplicate (OrderSN+Amount)
	return &shopeePkg.WalletTransactionResponse{
		Response: struct {
			TransactionList []shopeePkg.WalletTransaction `json:"transaction_list"`
			More            bool                           `json:"more"`
		}{
			TransactionList: []shopeePkg.WalletTransaction{
				{OrderSN: "DUP-ORDER-001", Amount: 100000, TransactionType: "wallet_order_income", BuyerUsername: "Buyer"},
				{OrderSN: "DUP-ORDER-001", Amount: 100000, TransactionType: "wallet_order_income", BuyerUsername: "Buyer"},
				{OrderSN: "DUP-ORDER-002", Amount: 50000, TransactionType: "wallet_order_income", BuyerUsername: "Buyer"},
				{OrderSN: "DUP-ORDER-002", Amount: 50000, TransactionType: "wallet_order_income", BuyerUsername: "Buyer"},
			},
			More: false,
		},
	}, nil
}

func (f *fakeShopeeClientWithDupes) GetEscrowDetails(req shopeePkg.GetEscrowDetailsRequest) (*shopeePkg.GetEscrowDetailsResponse, error) {
	f.t.Helper()
	wrappers := make([]shopeePkg.EscrowDetailWrapper, 0, len(req.OrderSNList))
	for _, sn := range req.OrderSNList {
		wrappers = append(wrappers, shopeePkg.EscrowDetailWrapper{
			EscrowDetail: &shopeePkg.EscrowOrder{
				OrderSN:       sn,
				BuyerUsername: "Buyer",
				TotalAmount:   100000,
				OrderIncome: shopeePkg.EscrowOrderData{
					EscrowAmount:   100000,
					Items: []shopeePkg.EscrowItemData{
						{ItemID: 1, ItemName: "Dup Test Item", QuantityPurchased: 1, SellingPrice: 100000},
					},
				},
				BuyerPaymentInfo: map[string]any{},
			},
		})
	}
	return &shopeePkg.GetEscrowDetailsResponse{Response: wrappers}, nil
}

func TestShopeeSync_WalletDedupe_ConsumesOnce(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "dedupe-tenant")
	svc.SetShopeeClient(&fakeShopeeClientWithDupes{t: t})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	if err := svc.SyncMonthWithProgress(ctx, 8, 2026, false, onProgress); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	// Should have 2 unique orders (not 4 — wallet deduped by OrderSN+Amount)
	var orderCount int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ?", "dedupe-tenant").
		Count(&orderCount)
	if orderCount != 2 {
		t.Errorf("expected 2 orders (deduped from 4 wallet entries), got %d", orderCount)
	}
}

// ---------------------------------------------------------------------------
// No transactions — gracefully handles empty period
// ---------------------------------------------------------------------------

type fakeShopeeClientEmpty struct{ t *testing.T }

func (f *fakeShopeeClientEmpty) GetWalletTransactions(req shopeePkg.GetWalletTransactionRequest) (*shopeePkg.WalletTransactionResponse, error) {
	return &shopeePkg.WalletTransactionResponse{
		Response: struct {
			TransactionList []shopeePkg.WalletTransaction `json:"transaction_list"`
			More            bool                           `json:"more"`
		}{
			TransactionList: []shopeePkg.WalletTransaction{},
			More:            false,
		},
	}, nil
}

func (f *fakeShopeeClientEmpty) GetEscrowDetails(req shopeePkg.GetEscrowDetailsRequest) (*shopeePkg.GetEscrowDetailsResponse, error) {
	return &shopeePkg.GetEscrowDetailsResponse{}, nil
}

func TestShopeeSync_NoTransactions_SucceedsWithZero(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "empty-tenant")
	svc.SetShopeeClient(&fakeShopeeClientEmpty{t: t})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	if err := svc.SyncMonthWithProgress(ctx, 9, 2026, false, onProgress); err != nil {
		t.Fatalf("sync error with no transactions: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Client not configured — returns error, not silent success
// ---------------------------------------------------------------------------

func TestShopeeSync_NoClient_ReturnsError(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "no-client-tenant")
	// No SetShopeeClient — should error

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	err := svc.SyncMonthWithProgress(ctx, 10, 2026, false, onProgress)
	if err == nil {
		t.Fatal("expected error when no Shopee client is configured, got nil")
	}
}

// ---------------------------------------------------------------------------
// Raw JSON round-trip — verify serialized order_income and buyer_payment_info
// ---------------------------------------------------------------------------

func TestShopeeSync_RawJSONRoundTrip(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "raw-json-tenant")
	svc.SetShopeeClient(&fakeShopeeClient{
		t:             t,
		walletTxCount: 1,
	})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	if err := svc.SyncMonthWithProgress(ctx, 11, 2026, false, onProgress); err != nil {
		t.Fatalf("sync error: %v", err)
	}

	var order models.ShopeeEscrowOrder
	if err := db.Where("tenant_id = ?", "raw-json-tenant").First(&order).Error; err != nil {
		t.Fatalf("order not found: %v", err)
	}

	if order.RawOrderIncome == nil || *order.RawOrderIncome == "" {
		t.Error("raw_order_income is nil or empty")
	} else {
		// Verify it's valid JSON
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(*order.RawOrderIncome), &parsed); err != nil {
			t.Errorf("raw_order_income is not valid JSON: %v", err)
		}
		if _, ok := parsed["escrow_amount"]; !ok {
			t.Error("raw_order_income missing escrow_amount field")
		}
	}

	if order.RawBuyerPaymentInfo == nil || *order.RawBuyerPaymentInfo == "" {
		t.Error("raw_buyer_payment_info is nil or empty")
	} else {
		var parsed map[string]interface{}
		if err := json.Unmarshal([]byte(*order.RawBuyerPaymentInfo), &parsed); err != nil {
			t.Errorf("raw_buyer_payment_info is not valid JSON: %v", err)
		}
	}

	var item models.ShopeeEscrowItem
	if err := db.Where("tenant_id = ? AND escrow_order_id = ?", "raw-json-tenant", order.ID).First(&item).Error; err != nil {
		t.Fatalf("item not found: %v", err)
	}
	if item.RawItemData == nil || *item.RawItemData == "" {
		t.Error("raw_item_data is nil or empty")
	}
}

// ---------------------------------------------------------------------------
// Guard test — updated to use fake client (was: placeholder guard)
// ---------------------------------------------------------------------------

func TestShopeeSyncGuard_RejectsPlaceholderNoData(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "test-tenant")
	svc.SetShopeeClient(&fakeShopeeClient{
		t:             t,
		walletTxCount: 3,
	})

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
	var sync models.ShopeeEscrowSync
	if err := db.Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		First(&sync).Error; err != nil {
		t.Fatalf("failed to find sync record: %v", err)
	}
	if sync.TotalOrders == 0 {
		t.Error("GUARD FAILED: SyncMonthWithProgress created a sync record with TotalOrders=0. ",
			"Real sync must set TotalOrders > 0 after persisting orders.")
	}

	// GUARD 2: Sync MUST persist order records
	var orderCount int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&orderCount)
	if orderCount == 0 {
		t.Error("GUARD FAILED: SyncMonthWithProgress completed without persisting any ShopeeEscrowOrder records. ",
			"Implement real sync logic that populates ShopeeEscrowOrder.")
	}

	// GUARD 3: Sync MUST persist item records
	var itemCount int64
	db.Model(&models.ShopeeEscrowItem{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&itemCount)
	if itemCount == 0 {
		t.Error("GUARD FAILED: SyncMonthWithProgress completed without persisting any ShopeeEscrowItem records. ",
			"Implement real sync logic that populates ShopeeEscrowItem.")
	}

	// GUARD 4: Sync MUST store raw JSON data (raw_order_income or raw_buyer_payment_info)
	var rawCount int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Where("(raw_order_income IS NOT NULL AND raw_order_income != '') OR (raw_buyer_payment_info IS NOT NULL AND raw_buyer_payment_info != '')").
		Count(&rawCount)
	if rawCount == 0 {
		t.Error("GUARD FAILED: SyncMonthWithProgress completed without storing raw JSON data. ",
			"Sync must persist raw_order_income or raw_buyer_payment_info for traceability.")
	}
}

func TestShopeeRepopulateGuard_RejectsNoopOnItemRestoration(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	svc := analytics.NewShopeeAnalyticsService(db, db, "test-tenant")
	ctx := context.Background()
	tenantID := "test-tenant"

	rawOrderData := `{"order_sn":"ORD-REPOP-001","items":[{"item_id":123}]}`
	rawPaymentData := `{"payment_method":"credit_card"}`
	order := models.ShopeeEscrowOrder{
		ID:                  uuid.New().String(),
		TenantID:            tenantID,
		OrderSN:             "ORD-REPOP-001",
		Month:               3,
		Year:                2026,
		RawOrderIncome:      &rawOrderData,
		RawBuyerPaymentInfo: &rawPaymentData,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	rawItemData := `{"item_name":"Guard Test Item","sku":"SKU-GUARD"}`
	itemCount := 3
	for i := 0; i < itemCount; i++ {
		item := models.ShopeeEscrowItem{
			ID:            uuid.New().String(),
			TenantID:      tenantID,
			EscrowOrderID: order.ID,
			Month:         intPtr(3),
			Year:          intPtr(2026),
			Quantity:      1,
			SellingPrice:  10000,
			OriginalPrice: 10000,
			RawItemData:   &rawItemData,
		}
		if err := db.Create(&item).Error; err != nil {
			t.Fatalf("failed to seed item %d: %v", i, err)
		}
	}

	var initialCount int64
	db.Model(&models.ShopeeEscrowItem{}).
		Where("tenant_id = ? AND escrow_order_id = ?", tenantID, order.ID).
		Count(&initialCount)
	if initialCount != int64(itemCount) {
		t.Fatalf("expected %d items initially, got %d", itemCount, initialCount)
	}

	if err := db.Where("tenant_id = ?", tenantID).Delete(&models.ShopeeEscrowItem{}).Error; err != nil {
		t.Fatalf("failed to clear items: %v", err)
	}
	var afterDelete int64
	db.Model(&models.ShopeeEscrowItem{}).Where("tenant_id = ?", tenantID).Count(&afterDelete)
	if afterDelete != 0 {
		t.Fatalf("expected 0 items after delete, got %d", afterDelete)
	}

	if err := svc.RepopulateItems(ctx, tenantID, "2026-03"); err != nil {
		t.Fatalf("RepopulateItems returned error: %v", err)
	}

	var finalCount int64
	db.Model(&models.ShopeeEscrowItem{}).
		Where("tenant_id = ? AND escrow_order_id = ?", tenantID, order.ID).
		Count(&finalCount)
	if finalCount != initialCount {
		t.Errorf("GUARD FAILED: RepopulateItems did not restore items from raw JSON data. "+
			"Expected %d items (matching pre-delete count), got %d. "+
			"Current RepopulateItems is a no-op — implement item reconstruction from raw JSON.",
			initialCount, finalCount)
	}
}
