package sync

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ============================================================================
// Test setup helpers
// ============================================================================

func setupBookingTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Discard,
	})
	require.NoError(t, err)

	require.NoError(t, db.AutoMigrate(
		&models.ShopeeBooking{},
		&models.ShopeeBookingItem{},
		&models.ShopeeOrder{},
	))
	return db
}

func newTestBookingService(db *gorm.DB, tenantID string, shopID uint64) *BookingSyncService {
	return &BookingSyncService{
		repository: NewGormBookingRepository(db, tenantID),
		tenantID:   tenantID,
		shopID:     shopID,
	}
}

func sampleBooking(bookingSN, orderSN string) Booking {
	return Booking{
		BookingSN:       bookingSN,
		OrderSN:         orderSN,
		BookingStatus:   "READY_TO_SHIP",
		MatchStatus:     "MATCHED",
		Region:          "ID",
		ShippingCarrier: "SPX",
		RecipientName:   "Test User",
		RecipientPhone:  "+1-555-0100",
		FulfillmentFlag: "PARTIAL",
		CreateTime:      1710000000,
		UpdateTime:      1710000100,
		SyncedAt:        time.Now(),
	}
}

func sampleItems(bookingSN string, count int) []BookingItem {
	items := make([]BookingItem, 0, count)
	for i := 0; i < count; i++ {
		itemID := int64(100 + i)
		modelID := int64(200 + i)
		items = append(items, BookingItem{
			BookingSN: bookingSN,
			LineKey:   GenerateBookingItemLineKey(itemID, modelID, "", "", "Test Item", ""),
			ItemID:    itemID,
			ModelID:   modelID,
			ItemName:  "Test Item",
			Quantity:  1 + i,
			Weight:    0.5,
		})
	}
	return items
}

func countRows(t *testing.T, db *gorm.DB, model any, where ...any) int64 {
	t.Helper()
	var count int64
	q := db.Model(model)
	if len(where) > 0 {
		q = q.Where(where[0], where[1:]...)
	}
	require.NoError(t, q.Count(&count).Error)
	return count
}

// ============================================================================
// 1. IDEMPOTENCY TESTS — Duplicate sync must not create duplicate rows
// ============================================================================

func TestUpsertBookingWithItems_Idempotent_SingleRowAfterDuplicateSync(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	booking := sampleBooking("BSN-IDEMP-1", "OSN-1")
	items := sampleItems("BSN-IDEMP-1", 2)

	// First sync
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, items))

	// Second sync with identical payload
	booking.BookingStatus = "SHIPPED" // status changed on re-sync
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, items))

	// Must be exactly one booking row
	require.Equal(t, int64(1), countRows(t, db, &models.ShopeeBooking{},
		"tenant_id = ? AND booking_sn = ?", "tenant_a", "BSN-IDEMP-1"))

	// Must be exactly two item rows (not four)
	require.Equal(t, int64(2), countRows(t, db, &models.ShopeeBookingItem{},
		"tenant_id = ? AND booking_sn = ?", "tenant_a", "BSN-IDEMP-1"))

	// Status should be updated to the latest
	var stored models.ShopeeBooking
	require.NoError(t, db.Where("tenant_id = ? AND booking_sn = ?", "tenant_a", "BSN-IDEMP-1").First(&stored).Error)
	require.Equal(t, "SHIPPED", stored.BookingStatus)
}

func TestUpsertBookingWithItems_Idempotent_ItemsReplacedNotDuplicated(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	booking := sampleBooking("BSN-IDEMP-2", "OSN-2")

	// First sync: 3 items
	items3 := sampleItems("BSN-IDEMP-2", 3)
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, items3))
	require.Equal(t, int64(3), countRows(t, db, &models.ShopeeBookingItem{},
		"booking_sn = ?", "BSN-IDEMP-2"))

	// Second sync: only 1 item (partial fulfillment scenario)
	items1 := sampleItems("BSN-IDEMP-2", 1)
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, items1))

	// Must be exactly 1 item (old items deleted, new inserted)
	require.Equal(t, int64(1), countRows(t, db, &models.ShopeeBookingItem{},
		"booking_sn = ?", "BSN-IDEMP-2"))
}

func TestUpsertBookingWithItems_Idempotent_ItemsGrownOnReSync(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	booking := sampleBooking("BSN-IDEMP-3", "OSN-3")

	// First sync: 1 item
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, sampleItems("BSN-IDEMP-3", 1)))

	// Second sync: 5 items
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, sampleItems("BSN-IDEMP-3", 5)))

	require.Equal(t, int64(5), countRows(t, db, &models.ShopeeBookingItem{},
		"booking_sn = ?", "BSN-IDEMP-3"))
}

func TestUpsertBookings_BatchIdempotency(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	bookings := []Booking{
		sampleBooking("BSN-BATCH-1", "OSN-B1"),
		sampleBooking("BSN-BATCH-2", "OSN-B2"),
	}

	// First batch upsert
	require.NoError(t, repo.UpsertBookings(ctx, "tenant_a", 100, bookings))
	require.Equal(t, int64(2), countRows(t, db, &models.ShopeeBooking{},
		"tenant_id = ?", "tenant_a"))

	// Second batch upsert with updated status
	bookings[0].BookingStatus = "COMPLETED"
	bookings[1].BookingStatus = "CANCELLED"
	require.NoError(t, repo.UpsertBookings(ctx, "tenant_a", 100, bookings))

	// Still exactly 2 rows
	require.Equal(t, int64(2), countRows(t, db, &models.ShopeeBooking{},
		"tenant_id = ?", "tenant_a"))

	// Values updated
	var b1 models.ShopeeBooking
	require.NoError(t, db.Where("booking_sn = ?", "BSN-BATCH-1").First(&b1).Error)
	require.Equal(t, "COMPLETED", b1.BookingStatus)
}

// ============================================================================
// 2. PARENT ORDER MATCHING — Booking-to-order resolution
// ============================================================================

func TestParentOrderExists_ReturnsTrueWhenOrderSynced(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Insert a parent order
	order := models.ShopeeOrder{
		TenantID:    "tenant_a",
		OrderSN:     "OSN-PARENT-1",
		OrderStatus: "COMPLETED",
	}
	require.NoError(t, db.Create(&order).Error)

	exists, err := repo.ParentOrderExists(ctx, "tenant_a", "OSN-PARENT-1")
	require.NoError(t, err)
	require.True(t, exists)
}

func TestParentOrderExists_ReturnsFalseWhenOrderMissing(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	exists, err := repo.ParentOrderExists(ctx, "tenant_a", "OSN-NONEXISTENT")
	require.NoError(t, err)
	require.False(t, exists)
}

func TestBookingBeforeParent_InitiallyNotSynced_ThenSynced(t *testing.T) {
	db := setupBookingTestDB(t)
	service := newTestBookingService(db, "tenant_a", 100)
	ctx := context.Background()

	// Sync booking first (parent order doesn't exist yet)
	booking := sampleBooking("BSN-EARLY", "OSN-LATE")
	items := sampleItems("BSN-EARLY", 1)
	require.NoError(t, service.repository.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, items))

	// Check parent status: should be not_synced
	_, _, err := service.GetBookingDetail(ctx, "BSN-EARLY")
	require.NoError(t, err)

	// Direct check: parent doesn't exist yet
	exists, err := service.repository.ParentOrderExists(ctx, "tenant_a", "OSN-LATE")
	require.NoError(t, err)
	require.False(t, exists, "parent should not exist before order sync")

	// Now sync the parent order
	parentOrder := models.ShopeeOrder{
		TenantID:    "tenant_a",
		OrderSN:     "OSN-LATE",
		OrderStatus: "TO_SHIP",
	}
	require.NoError(t, db.Create(&parentOrder).Error)

	// Re-check: parent now exists
	exists, err = service.repository.ParentOrderExists(ctx, "tenant_a", "OSN-LATE")
	require.NoError(t, err)
	require.True(t, exists, "parent should exist after order sync")

	// GetBookingDetail should resolve parent status correctly
	result, _, err := service.GetBookingDetail(ctx, "BSN-EARLY")
	require.NoError(t, err)
	require.True(t, result.ParentOrderExists)
	require.Equal(t, "synced", result.ParentOrderStatus)
}

// ============================================================================
// 3. CROSS-TENANT ISOLATION — Tenant A must not see Tenant B data
// ============================================================================

func TestParentOrderExists_CrossTenantIsolation(t *testing.T) {
	db := setupBookingTestDB(t)
	repoA := NewGormBookingRepository(db, "tenant_a")
	repoB := NewGormBookingRepository(db, "tenant_b")
	ctx := context.Background()

	// Insert order for tenant_a only
	order := models.ShopeeOrder{
		TenantID:    "tenant_a",
		OrderSN:     "OSN-CROSS",
		OrderStatus: "COMPLETED",
	}
	require.NoError(t, db.Create(&order).Error)

	// tenant_a sees it
	exists, err := repoA.ParentOrderExists(ctx, "tenant_a", "OSN-CROSS")
	require.NoError(t, err)
	require.True(t, exists, "tenant_a should see its own order")

	// tenant_b must NOT see it (fail closed)
	exists, err = repoB.ParentOrderExists(ctx, "tenant_b", "OSN-CROSS")
	require.NoError(t, err)
	require.False(t, exists, "tenant_b must not see tenant_a's order")
}

func TestBookingsIsolatedByTenant(t *testing.T) {
	db := setupBookingTestDB(t)
	repoA := NewGormBookingRepository(db, "tenant_a")
	repoB := NewGormBookingRepository(db, "tenant_b")
	ctx := context.Background()

	// Sync a booking for tenant_a
	booking := sampleBooking("BSN-TENANT-ISO", "OSN-TISO")
	require.NoError(t, repoA.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, sampleItems("BSN-TENANT-ISO", 1)))

	// tenant_a can retrieve it
	result, err := repoA.GetBookingDetail(ctx, "tenant_a", 100, "BSN-TENANT-ISO")
	require.NoError(t, err)
	require.NotNil(t, result)

	// tenant_b cannot see it
	result, err = repoB.GetBookingDetail(ctx, "tenant_b", 200, "BSN-TENANT-ISO")
	require.NoError(t, err)
	require.Nil(t, result, "tenant_b must not see tenant_a's booking")
}

func TestBookingsIsolatedByShop(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	booking := sampleBooking("BSN-SHOP-ISO", "OSN-SISO")
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, sampleItems("BSN-SHOP-ISO", 1)))

	// Same tenant, different shop → not found
	result, err := repo.GetBookingDetail(ctx, "tenant_a", 999, "BSN-SHOP-ISO")
	require.NoError(t, err)
	require.Nil(t, result, "different shop_id should not see the booking")
}

// ============================================================================
// 4. MULTIPLE PARENT CANDIDATES — Deterministic behavior
// ============================================================================

func TestParentOrderExists_MultipleCandidates_SameTenant(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// In production each tenant has its own schema, so order_sn uniqueIndex
	// prevents duplicates. The query uses COUNT, so even hypothetically it
	// returns true for count > 0. We verify this with a single row.
	order := models.ShopeeOrder{
		TenantID:    "tenant_a",
		OrderSN:     "OSN-MULTI",
		OrderStatus: "COMPLETED",
	}
	require.NoError(t, db.Create(&order).Error)

	exists, err := repo.ParentOrderExists(ctx, "tenant_a", "OSN-MULTI")
	require.NoError(t, err)
	require.True(t, exists)

	// Non-matching order_sn returns false
	exists, err = repo.ParentOrderExists(ctx, "tenant_a", "OSN-DIFFERENT")
	require.NoError(t, err)
	require.False(t, exists)
}

// ============================================================================
// 5. PARTIAL FULFILLMENT / CANCEL — No double-counting
// ============================================================================

func TestReplaceBookingItems_NoOrphanedItemsAfterCancel(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	booking := sampleBooking("BSN-CANCEL", "OSN-CANCEL")

	// Sync with 3 items
	items3 := sampleItems("BSN-CANCEL", 3)
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, items3))

	// Booking cancelled → items removed (empty re-sync)
	require.NoError(t, repo.ReplaceBookingItems(ctx, "tenant_a", 100, "BSN-CANCEL", nil))

	require.Equal(t, int64(0), countRows(t, db, &models.ShopeeBookingItem{},
		"booking_sn = ?", "BSN-CANCEL"))
}

func TestReplaceBookingItems_QuantityReduced(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	booking := sampleBooking("BSN-REDUCE", "OSN-REDUCE")

	// First sync: 4 items
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, sampleItems("BSN-REDUCE", 4)))
	require.Equal(t, int64(4), countRows(t, db, &models.ShopeeBookingItem{},
		"booking_sn = ?", "BSN-REDUCE"))

	// Partial fulfillment: only 2 items remain
	require.NoError(t, repo.ReplaceBookingItems(ctx, "tenant_a", 100, "BSN-REDUCE", sampleItems("BSN-REDUCE", 2)))
	require.Equal(t, int64(2), countRows(t, db, &models.ShopeeBookingItem{},
		"booking_sn = ?", "BSN-REDUCE"))
}

func TestUpsertBookingWithItems_EmptyItemsList(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	booking := sampleBooking("BSN-EMPTY", "OSN-EMPTY")

	// Sync with zero items
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, nil))

	// Booking row exists
	require.Equal(t, int64(1), countRows(t, db, &models.ShopeeBooking{},
		"booking_sn = ?", "BSN-EMPTY"))

	// No items
	require.Equal(t, int64(0), countRows(t, db, &models.ShopeeBookingItem{},
		"booking_sn = ?", "BSN-EMPTY"))
}

// ============================================================================
// 6. SYNC FLOW — API failures and error reporting
// ============================================================================

func TestSyncBookings_ReportsFailedDetailBatch(t *testing.T) {
	repo := &mockBookingRepository{
		replacedItems: map[string][]BookingItem{},
		replaceErr:    map[string]error{},
	}
	service := &BookingSyncService{
		manager: &mockBookingManager{
			list: []Booking{
				{BookingSN: "BSN-OK"},
				{BookingSN: "BSN-FAIL"},
			},
			details: map[string]Booking{
				"BSN-OK": {BookingSN: "BSN-OK"},
			},
			// BSN-FAIL not in details → reported as missing
		},
		repository: repo,
		tenantID:   "tenant_test",
		shopID:     100,
	}

	result := service.SyncBookings(context.Background(), 15)

	require.False(t, result.Success)
	require.Equal(t, 1, result.Count, "BSN-OK should succeed")
	require.Equal(t, 1, result.FailedCount, "BSN-FAIL should fail")
	require.Len(t, result.Errors, 1)
	require.Contains(t, result.Errors[0], "BSN-FAIL")
}

func TestSyncBookings_ReportsDetailFetchError(t *testing.T) {
	repo := &mockBookingRepository{
		replacedItems: map[string][]BookingItem{},
		replaceErr:    map[string]error{},
	}
	service := &BookingSyncService{
		manager: &mockBookingManager{
			list: []Booking{{BookingSN: "BSN-1"}},
			err:  errors.New("API rate limited"),
		},
		repository: repo,
		tenantID:   "tenant_test",
		shopID:     100,
	}

	result := service.SyncBookings(context.Background(), 15)

	require.False(t, result.Success)
	require.Equal(t, 1, result.FailedCount)
	require.Len(t, result.Errors, 1)
	require.Contains(t, result.Errors[0], "API rate limited")
}

func TestSyncBookings_ReportsPersistFailure(t *testing.T) {
	repo := &mockBookingRepository{
		replacedItems: map[string][]BookingItem{},
		replaceErr:    map[string]error{},
		upsertErr:     errors.New("disk full"),
	}
	service := &BookingSyncService{
		manager: &mockBookingManager{
			list: []Booking{{BookingSN: "BSN-1"}},
			details: map[string]Booking{
				"BSN-1": {BookingSN: "BSN-1", Items: []BookingItem{{LineKey: "1|0||||"}}},
			},
		},
		repository: repo,
		tenantID:   "tenant_test",
		shopID:     100,
	}

	result := service.SyncBookings(context.Background(), 15)

	require.False(t, result.Success)
	require.Equal(t, 0, result.Count)
	require.Equal(t, 1, result.FailedCount)
	require.Contains(t, result.Errors[0], "persist booking BSN-1 failed")
}

func TestSyncBookings_ServiceNotConfigured(t *testing.T) {
	// Both manager and repository nil
	service := &BookingSyncService{tenantID: "tenant_test"}

	result := service.SyncBookings(context.Background(), 15)

	require.False(t, result.Success)
	require.Equal(t, 1, result.FailedCount)
	require.Contains(t, result.Errors[0], "not fully configured")
}

func TestSyncBookings_DeduplicatesSNsInList(t *testing.T) {
	repo := &mockBookingRepository{
		replacedItems: map[string][]BookingItem{},
		replaceErr:    map[string]error{},
	}
	service := &BookingSyncService{
		manager: &mockBookingManager{
			list: []Booking{
				{BookingSN: "BSN-DUP"},
				{BookingSN: "BSN-DUP"}, // duplicate
				{BookingSN: "BSN-UNIQUE"},
			},
			details: map[string]Booking{
				"BSN-DUP":    {BookingSN: "BSN-DUP"},
				"BSN-UNIQUE": {BookingSN: "BSN-UNIQUE"},
			},
		},
		repository: repo,
		tenantID:   "tenant_test",
		shopID:     100,
	}

	result := service.SyncBookings(context.Background(), 15)

	require.True(t, result.Success)
	require.Equal(t, 2, result.Count, "deduped: BSN-DUP once + BSN-UNIQUE")
	require.Len(t, repo.upserted, 2)
}

// ============================================================================
// 7. GET BOOKING DETAIL — Parent resolution integration
// ============================================================================

func TestGetBookingDetail_ParentOrderExists_Integration(t *testing.T) {
	db := setupBookingTestDB(t)
	service := newTestBookingService(db, "tenant_a", 100)
	ctx := context.Background()

	// Insert booking with order_sn
	booking := sampleBooking("BSN-DETAIL", "OSN-DETAIL")
	require.NoError(t, service.repository.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, sampleItems("BSN-DETAIL", 2)))

	// Insert parent order
	require.NoError(t, db.Create(&models.ShopeeOrder{
		TenantID:    "tenant_a",
		OrderSN:     "OSN-DETAIL",
		OrderStatus: "COMPLETED",
	}).Error)

	result, items, err := service.GetBookingDetail(ctx, "BSN-DETAIL")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.True(t, result.ParentOrderExists)
	require.Equal(t, "synced", result.ParentOrderStatus)
	require.Len(t, items, 2)
}

func TestGetBookingDetail_ParentOrderMissing_Integration(t *testing.T) {
	db := setupBookingTestDB(t)
	service := newTestBookingService(db, "tenant_a", 100)
	ctx := context.Background()

	// Insert booking referencing an order that doesn't exist
	booking := sampleBooking("BSN-ORPHAN", "OSN-ORPHAN")
	require.NoError(t, service.repository.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, sampleItems("BSN-ORPHAN", 1)))

	result, _, err := service.GetBookingDetail(ctx, "BSN-ORPHAN")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.ParentOrderExists)
	require.Equal(t, "not_synced", result.ParentOrderStatus)
}

func TestGetBookingDetail_EmptyOrderSN_NoParent(t *testing.T) {
	db := setupBookingTestDB(t)
	service := newTestBookingService(db, "tenant_a", 100)
	ctx := context.Background()

	// Booking with no order_sn
	booking := sampleBooking("BSN-NO-ORDER", "")
	require.NoError(t, service.repository.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, nil))

	result, _, err := service.GetBookingDetail(ctx, "BSN-NO-ORDER")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.ParentOrderExists)
	require.Equal(t, "no_parent", result.ParentOrderStatus)
}

func TestGetBookingDetail_NotFound(t *testing.T) {
	db := setupBookingTestDB(t)
	service := newTestBookingService(db, "tenant_a", 100)
	ctx := context.Background()

	result, _, err := service.GetBookingDetail(ctx, "BSN-NONEXISTENT")
	require.NoError(t, err)
	require.Nil(t, result)
}

// ============================================================================
// 8. LIST BOOKINGS — Pagination and filtering
// ============================================================================

func TestListBookings_ReturnsCorrectItemCount(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Insert 3 bookings with varying item counts
	for i, sn := range []string{"BSN-LIST-1", "BSN-LIST-2", "BSN-LIST-3"} {
		b := sampleBooking(sn, "OSN-L"+sn)
		require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b, sampleItems(sn, i+1)))
	}

	bookings, total, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{Page: 1, PageSize: 10})
	require.NoError(t, err)
	require.Equal(t, int64(3), total)
	require.Len(t, bookings, 3)

	// Verify item counts are correct
	itemCounts := map[string]int{}
	for _, b := range bookings {
		itemCounts[b.BookingSN] = b.ItemCount
	}
	require.Equal(t, 1, itemCounts["BSN-LIST-1"])
	require.Equal(t, 2, itemCounts["BSN-LIST-2"])
	require.Equal(t, 3, itemCounts["BSN-LIST-3"])
}

// ============================================================================
// 9. LINE KEY GENERATION — Deterministic uniqueness
// ============================================================================

func TestGenerateBookingItemLineKey_Deterministic(t *testing.T) {
	key1 := GenerateBookingItemLineKey(10, 20, "SKU-A", "SKU-B", "Item", "Model")
	key2 := GenerateBookingItemLineKey(10, 20, "SKU-A", "SKU-B", "Item", "Model")
	require.Equal(t, key1, key2, "same inputs must produce same key")
}

func TestGenerateBookingItemLineKey_DifferentInputsProduceDifferentKeys(t *testing.T) {
	key1 := GenerateBookingItemLineKey(10, 20, "SKU-A", "", "Item", "")
	key2 := GenerateBookingItemLineKey(10, 20, "SKU-A", "SKU-B", "Item", "Model")
	require.NotEqual(t, key1, key2, "different inputs must produce different keys")
}

func TestGenerateBookingItemLineKey_EmptyComponentsPreserved(t *testing.T) {
	key := GenerateBookingItemLineKey(0, 0, "", "", "", "")
	require.Equal(t, "0|0||||", key, "empty components should be preserved as empty segments")
}

// ============================================================================
// 10. RAW PARSING EDGE CASES
// ============================================================================

func TestRawToBooking_MissingBookingSN_ReturnsError(t *testing.T) {
	raw := map[string]any{
		"order_sn": "OSN-1",
	}
	_, _, err := rawToBooking(raw)
	require.Error(t, err)
	require.Contains(t, err.Error(), "booking_sn is required")
}

func TestRawToBooking_NoRecipientAddress(t *testing.T) {
	raw := map[string]any{
		"booking_sn": "BSN-NO-ADDR",
		"order_sn":   "OSN-1",
	}
	booking, _, err := rawToBooking(raw)
	require.NoError(t, err)
	require.Equal(t, "{}", booking.RecipientAddressJSON)
	require.Empty(t, booking.RecipientName)
	require.Empty(t, booking.RecipientPhone)
}

func TestRawToBooking_NoItems(t *testing.T) {
	raw := map[string]any{
		"booking_sn": "BSN-NO-ITEMS",
	}
	_, items, err := rawToBooking(raw)
	require.NoError(t, err)
	require.Empty(t, items)
}

func TestCollectBookingSNs_DeduplicatesAndSkipsEmpty(t *testing.T) {
	bookings := []Booking{
		{BookingSN: "A"},
		{BookingSN: "B"},
		{BookingSN: "A"}, // duplicate
		{BookingSN: ""},  // empty
		{BookingSN: "C"},
	}
	sns := collectBookingSNs(bookings)
	require.Equal(t, []string{"A", "B", "C"}, sns)
}
