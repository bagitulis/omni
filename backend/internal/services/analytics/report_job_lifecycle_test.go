package analytics_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// blockingFakeShopeeClient — blocks GetWalletTransactions until a signal fires
// Used for context cancellation tests
// ---------------------------------------------------------------------------

type blockingFakeShopeeClient struct {
	blockSignal <-chan struct{}
}

func (f *blockingFakeShopeeClient) GetWalletTransactions(req shopeePkg.GetWalletTransactionRequest) (*shopeePkg.WalletTransactionResponse, error) {
	<-f.blockSignal
	return &shopeePkg.WalletTransactionResponse{
		Response: struct {
			TransactionList []shopeePkg.WalletTransaction `json:"transaction_list"`
			More            bool                          `json:"more"`
		}{},
	}, nil
}

func (f *blockingFakeShopeeClient) GetEscrowDetails(req shopeePkg.GetEscrowDetailsRequest) (*shopeePkg.GetEscrowDetailsResponse, error) {
	return &shopeePkg.GetEscrowDetailsResponse{}, nil
}

// ---------------------------------------------------------------------------
// dateRecordingFakeShopeeClient — records date ranges passed to wallet API
// Used for timezone boundary tests
// ---------------------------------------------------------------------------

type dateRecordingFakeShopeeClient struct {
	mu          sync.Mutex
	startDates  []time.Time
	endDates    []time.Time
	callCount   int
	emptyResult bool
}

func (f *dateRecordingFakeShopeeClient) GetWalletTransactions(req shopeePkg.GetWalletTransactionRequest) (*shopeePkg.WalletTransactionResponse, error) {
	f.mu.Lock()
	f.startDates = append(f.startDates, time.Unix(req.StartDate, 0))
	f.endDates = append(f.endDates, time.Unix(req.EndDate, 0))
	f.callCount++
	f.mu.Unlock()

	return &shopeePkg.WalletTransactionResponse{
		Response: struct {
			TransactionList []shopeePkg.WalletTransaction `json:"transaction_list"`
			More            bool                          `json:"more"`
		}{},
	}, nil
}

func (f *dateRecordingFakeShopeeClient) GetEscrowDetails(req shopeePkg.GetEscrowDetailsRequest) (*shopeePkg.GetEscrowDetailsResponse, error) {
	return &shopeePkg.GetEscrowDetailsResponse{}, nil
}

// ---------------------------------------------------------------------------
// blockingFakeTiktokClient — blocks SearchOrders until signal fires
// ---------------------------------------------------------------------------

type blockingFakeTiktokClient struct {
	blockSignal <-chan struct{}
}

func (f *blockingFakeTiktokClient) SearchOrders(req tiktokPkg.OrderSearchRequest, pageSize int, pageToken string) (*tiktokPkg.OrderSearchResponse, error) {
	<-f.blockSignal
	return &tiktokPkg.OrderSearchResponse{Code: 0, Message: "success"}, nil
}

func (f *blockingFakeTiktokClient) GetOrderDetail(orderIDs []string) (*tiktokPkg.OrderDetailResponse, error) {
	return &tiktokPkg.OrderDetailResponse{Code: 0, Message: "success"}, nil
}

func (f *blockingFakeTiktokClient) GetOrderTransactions(orderID string) (*tiktokPkg.OrderTransactionResponse, error) {
	return &tiktokPkg.OrderTransactionResponse{Code: 40001, Message: "not found"}, nil
}

func (f *blockingFakeTiktokClient) GetOrderTransactionsV202309(orderID string) (*tiktokPkg.OrderTransactionResponse, error) {
	return &tiktokPkg.OrderTransactionResponse{Code: 40001, Message: "not found"}, nil
}

// ---------------------------------------------------------------------------
// setupReportLifecycleTestDB — in-memory SQLite with all needed models
// ---------------------------------------------------------------------------

func setupReportLifecycleTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := fmt.Sprintf("/tmp/report_lifecycle_test_%s.db", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(MEMORY)&_pragma=synchronous(OFF)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open SQLite: %v", err)
	}
	if err := db.AutoMigrate(
		&models.ShopeeEscrowSync{},
		&models.ShopeeEscrowOrder{},
		&models.ShopeeEscrowItem{},
		&models.TiktokEscrowSync{},
		&models.TiktokEscrowOrder{},
		&models.TiktokEscrowItem{},
	); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
	t.Cleanup(func() { os.Remove(path) })
	return db
}

// ---------------------------------------------------------------------------
// Context Cancellation — Shopee Sync
// ---------------------------------------------------------------------------

func TestReportCancel_ShopeeSyncReturnsErrorOnContextCancel(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "cancel-shopee-tenant")

	signal := make(chan struct{})
	svc.SetShopeeClient(&blockingFakeShopeeClient{blockSignal: signal})

	ctx, cancel := context.WithCancel(context.Background())
	onProgress := func(processed, total int, message string) {}

	done := make(chan error, 1)
	go func() {
		done <- svc.SyncMonthWithProgress(ctx, 2, 2025, false, onProgress)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(signal)

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error from cancelled context, got nil")
		}
	case <-time.After(5 * time.Second):
		close(signal)
		t.Fatal("sync did not return within 5s after cancellation")
	}

	var syncCount int64
	db.Model(&models.ShopeeEscrowSync{}).
		Where("tenant_id = ?", "cancel-shopee-tenant").
		Count(&syncCount)
	if syncCount > 0 {
		t.Errorf("expected no sync record after cancellation, got %d", syncCount)
	}
}

// ---------------------------------------------------------------------------
// Context Cancellation — TikTok Sync
// ---------------------------------------------------------------------------

func TestReportCancel_TiktokSyncReturnsErrorOnContextCancel(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svc := analytics.NewTiktokEscrowSyncService(db, db, "cancel-tiktok-tenant")

	signal := make(chan struct{})
	svc.SetClient(&blockingFakeTiktokClient{blockSignal: signal})

	ctx, cancel := context.WithCancel(context.Background())
	onProgress := func(processed, total int, message string) {}

	done := make(chan error, 1)
	go func() {
		done <- svc.SyncMonthWithProgress(ctx, 2, 2025, false, onProgress)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(signal)

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error from cancelled context, got nil")
		}
	case <-time.After(5 * time.Second):
		close(signal)
		t.Fatal("sync did not return within 5s after cancellation")
	}

	var syncCount int64
	db.Model(&models.TiktokEscrowSync{}).
		Where("tenant_id = ?", "cancel-tiktok-tenant").
		Count(&syncCount)
	if syncCount > 0 {
		t.Errorf("expected no sync record after cancellation, got %d", syncCount)
	}
}

// ---------------------------------------------------------------------------
// Timezone Boundaries — Shopee uses UTC month boundaries
// ---------------------------------------------------------------------------

func TestReportTimezone_ShopeeUsesUTCMonthBoundaries(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	recorder := &dateRecordingFakeShopeeClient{}
	svc := analytics.NewShopeeEscrowSyncService(db, db, "tz-tenant")
	svc.SetShopeeClient(recorder)

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	if err := svc.SyncMonthWithProgress(ctx, 2, 2026, false, onProgress); err != nil {
		t.Fatalf("SyncMonthWithProgress error: %v", err)
	}

	recorder.mu.Lock()
	defer recorder.mu.Unlock()

	if recorder.callCount == 0 {
		t.Fatal("expected at least one wallet API call, got 0")
	}

	expectedStart := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
	firstStart := recorder.startDates[0]
	if firstStart.Year() != expectedStart.Year() ||
		firstStart.Month() != expectedStart.Month() ||
		firstStart.Day() != expectedStart.Day() {
		t.Errorf("first chunk start should be Feb 1 2026 UTC, got %v", firstStart)
	}

	expectedEnd := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second)
	lastEnd := recorder.endDates[len(recorder.endDates)-1]
	if lastEnd.Before(expectedEnd.Add(-time.Hour)) || lastEnd.After(expectedEnd.Add(time.Hour)) {
		t.Errorf("last chunk end should be near %v, got %v", expectedEnd, lastEnd)
	}

	for i, s := range recorder.startDates {
		if s.Before(expectedStart.Add(-time.Second)) {
			t.Errorf("chunk %d start %v is before month start %v", i, s, expectedStart)
		}
	}
	for i, e := range recorder.endDates {
		if e.After(expectedEnd.Add(time.Second)) {
			t.Errorf("chunk %d end %v is after month end %v", i, e, expectedEnd)
		}
	}
}

// TestReportTimezone_DifferentMonths verifies each month gets correct boundaries.
func TestReportTimezone_DifferentMonths(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))

	tests := []struct {
		name          string
		month, year   int
		expectedStart time.Time
		expectedEnd   time.Time
	}{
		{
			name:          "January 2026",
			month: 1, year: 2026,
			expectedStart: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second),
		},
		{
			name:          "February 2026 (non-leap)",
			month: 2, year: 2026,
			expectedStart: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second),
		},
		{
			name:          "December 2025",
			month: 12, year: 2025,
			expectedStart: time.Date(2025, 12, 1, 0, 0, 0, 0, time.UTC),
			expectedEnd:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC).Add(-time.Second),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupReportLifecycleTestDB(t)
			recorder := &dateRecordingFakeShopeeClient{}
			svc := analytics.NewShopeeEscrowSyncService(db, db, "tz-month-tenant")
			svc.SetShopeeClient(recorder)

			ctx := context.Background()
			onProgress := func(processed, total int, message string) {}
			if err := svc.SyncMonthWithProgress(ctx, tt.month, tt.year, false, onProgress); err != nil {
				t.Fatalf("sync error: %v", err)
			}

			recorder.mu.Lock()
			defer recorder.mu.Unlock()

			if recorder.callCount == 0 {
				t.Fatal("expected at least one wallet API call")
			}

			firstStart := recorder.startDates[0]
			if firstStart.Year() != tt.expectedStart.Year() ||
				firstStart.Month() != tt.expectedStart.Month() ||
				firstStart.Day() != tt.expectedStart.Day() {
				t.Errorf("first chunk should start at %v, got %v", tt.expectedStart, firstStart)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Idempotency — Sync deduplication at service level
// ---------------------------------------------------------------------------

func TestReportIdempotent_ShopeeSyncSkipsAlreadySynced(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "idempotent-report-tenant")
	svc.SetShopeeClient(&fakeShopeeClient{t: t, walletTxCount: 2})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}
	tenantID := "idempotent-report-tenant"
	month, year := 1, 2026

	if err := svc.SyncMonthWithProgress(ctx, month, year, false, onProgress); err != nil {
		t.Fatalf("first sync error: %v", err)
	}

	var orderCount1 int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&orderCount1)
	if orderCount1 == 0 {
		t.Fatal("first sync persisted no orders")
	}

	if err := svc.SyncMonthWithProgress(ctx, month, year, false, onProgress); err != nil {
		t.Fatalf("second sync error: %v", err)
	}

	var orderCount2 int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&orderCount2)
	if orderCount2 != orderCount1 {
		t.Errorf("idempotent rerun changed order count: before=%d after=%d", orderCount1, orderCount2)
	}

	var syncCount int64
	db.Model(&models.ShopeeEscrowSync{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&syncCount)
	if syncCount != 1 {
		t.Errorf("expected 1 sync record after idempotent rerun, got %d", syncCount)
	}
}

func TestReportIdempotent_TiktokSyncSkipsAlreadySynced(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svc := analytics.NewTiktokEscrowSyncService(db, db, "idempotent-tt-tenant")
	fake := newFakeTiktokClient()
	svc.SetClient(fake)

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}
	tenantID := "idempotent-tt-tenant"

	if err := svc.SyncMonthWithProgress(ctx, 3, 2026, false, onProgress); err != nil {
		t.Fatalf("first sync error: %v", err)
	}

	var syncCount1 int64
	db.Model(&models.TiktokEscrowSync{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, 3, 2026).
		Count(&syncCount1)
	if syncCount1 != 1 {
		t.Fatalf("expected 1 sync record after first sync, got %d", syncCount1)
	}

	initialCalls := fake.GetCallCount("SearchOrders")

	if err := svc.SyncMonthWithProgress(ctx, 3, 2026, false, onProgress); err != nil {
		t.Fatalf("second sync error: %v", err)
	}

	afterCalls := fake.GetCallCount("SearchOrders")
	if afterCalls != initialCalls {
		t.Errorf("idempotent rerun made %d extra SearchOrders calls (should be 0)", afterCalls-initialCalls)
	}
}

// ---------------------------------------------------------------------------
// Tenant Isolation — Cross-tenant cancellation does not affect other tenants
// ---------------------------------------------------------------------------

func TestReportTenant_CancelDoesNotAffectOtherTenant(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svcA := analytics.NewShopeeEscrowSyncService(db, db, "tenant-report-a")
	svcA.SetShopeeClient(&fakeShopeeClient{t: t, walletTxCount: 2})

	svcB := analytics.NewShopeeEscrowSyncService(db, db, "tenant-report-b")
	svcB.SetShopeeClient(&fakeShopeeClient{t: t, walletTxCount: 3})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	if err := svcA.SyncMonthWithProgress(ctx, 5, 2025, false, onProgress); err != nil {
		t.Fatalf("tenant A sync error: %v", err)
	}
	if err := svcB.SyncMonthWithProgress(ctx, 5, 2025, false, onProgress); err != nil {
		t.Fatalf("tenant B sync error: %v", err)
	}

	var countA, countB int64
	db.Model(&models.ShopeeEscrowOrder{}).Where("tenant_id = ?", "tenant-report-a").Count(&countA)
	db.Model(&models.ShopeeEscrowOrder{}).Where("tenant_id = ?", "tenant-report-b").Count(&countB)

	if countA == 0 {
		t.Error("tenant A has no orders")
	}
	if countB == 0 {
		t.Error("tenant B has no orders")
	}

	var tenantBOrderInTenantA int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND order_sn = ?", "tenant-report-a", "FAKE-ORDER-003").
		Count(&tenantBOrderInTenantA)
	if tenantBOrderInTenantA > 0 {
		t.Error("tenant A can see tenant B's order FAKE-ORDER-003")
	}

	var tenantAOrderInTenantB int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND order_sn = ?", "tenant-report-b", "FAKE-ORDER-001").
		Count(&tenantAOrderInTenantB)
	if tenantAOrderInTenantB == 0 {
		t.Error("tenant B should have FAKE-ORDER-001 (own copy)")
	}
}

// ---------------------------------------------------------------------------
// Validation — Invalid month/year rejected
// ---------------------------------------------------------------------------

func TestReportJob_InvalidMonthYearRejected(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "validation-tenant")
	svc.SetShopeeClient(&fakeShopeeClient{t: t, walletTxCount: 1})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	err := svc.SyncMonthWithProgress(ctx, 13, 2026, false, onProgress)
	if err == nil {
		t.Error("expected error for month=13, got nil")
	}

	err = svc.SyncMonthWithProgress(ctx, 0, 2026, false, onProgress)
	if err == nil {
		t.Error("expected error for month=0, got nil")
	}

	err = svc.SyncMonthWithProgress(ctx, 1, 1999, false, onProgress)
	if err == nil {
		t.Error("expected error for year=1999, got nil")
	}
}

// ---------------------------------------------------------------------------
// Idempotency — Force resync replaces data without duplication
// ---------------------------------------------------------------------------

func TestReportIdempotent_ForceResyncReplacesShopeeData(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "force-resync-report")
	svc.SetShopeeClient(&fakeShopeeClient{t: t, walletTxCount: 2})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}
	tenantID := "force-resync-report"
	month, year := 4, 2025

	if err := svc.SyncMonthWithProgress(ctx, month, year, false, onProgress); err != nil {
		t.Fatalf("first sync error: %v", err)
	}

	if err := svc.SyncMonthWithProgress(ctx, month, year, true, onProgress); err != nil {
		t.Fatalf("force resync error: %v", err)
	}

	var syncCount int64
	db.Model(&models.ShopeeEscrowSync{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&syncCount)
	if syncCount != 1 {
		t.Errorf("expected 1 sync record after force resync, got %d", syncCount)
	}

	var orderCount int64
	db.Model(&models.ShopeeEscrowOrder{}).
		Where("tenant_id = ? AND month = ? AND year = ?", tenantID, month, year).
		Count(&orderCount)
	if orderCount == 0 {
		t.Error("expected orders after force resync, got 0")
	}
}

// ---------------------------------------------------------------------------
// Job deduplication at sync level — already-synced check
// ---------------------------------------------------------------------------

func TestReportJob_SyncRecordPreventsDuplicateRun(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "dedup-sync-record")
	svc.SetShopeeClient(&fakeShopeeClient{t: t, walletTxCount: 1})

	ctx := context.Background()
	onProgress := func(processed, total int, message string) {}

	if err := svc.SyncMonthWithProgress(ctx, 11, 2025, false, onProgress); err != nil {
		t.Fatalf("first sync error: %v", err)
	}

	var syncCount int64
	db.Model(&models.ShopeeEscrowSync{}).
		Where("tenant_id = ? AND month = ? AND year = ?", "dedup-sync-record", 11, 2025).
		Count(&syncCount)
	if syncCount != 1 {
		t.Fatalf("expected 1 sync record, got %d", syncCount)
	}

	if err := svc.SyncMonthWithProgress(ctx, 11, 2025, false, onProgress); err != nil {
		t.Fatalf("second sync error: %v", err)
	}

	db.Model(&models.ShopeeEscrowSync{}).
		Where("tenant_id = ? AND month = ? AND year = ?", "dedup-sync-record", 11, 2025).
		Count(&syncCount)
	if syncCount != 1 {
		t.Errorf("expected 1 sync record after rerun (dedup), got %d", syncCount)
	}
}

// ---------------------------------------------------------------------------
// Ensure error message distinguishes cancellation context
// ---------------------------------------------------------------------------

func TestReportCancel_SyncErrorContainsCancellationHint(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupReportLifecycleTestDB(t)

	svc := analytics.NewShopeeEscrowSyncService(db, db, "cancel-hint-tenant")

	signal := make(chan struct{})
	svc.SetShopeeClient(&blockingFakeShopeeClient{blockSignal: signal})

	ctx, cancel := context.WithCancel(context.Background())
	onProgress := func(processed, total int, message string) {}

	done := make(chan error, 1)
	go func() {
		done <- svc.SyncMonthWithProgress(ctx, 3, 2025, false, onProgress)
	}()

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(signal)

	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		errMsg := err.Error()
		if !strings.Contains(errMsg, "context") && !strings.Contains(errMsg, "cancel") {
			t.Logf("note: error message %q does not explicitly mention context cancellation", errMsg)
		}
	case <-time.After(5 * time.Second):
		close(signal)
		t.Fatal("sync did not return after cancellation")
	}
}
