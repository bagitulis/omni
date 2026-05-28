package sync

import (
	"context"
	"fmt"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/require"
)

// ============================================================================
// BOOKING ↔ ORDER MANAGEMENT INTEGRATION TESTS
// Verifies: filter correctness, parent linking, list pagination,
// booking detail with items, parent status resolution, bulk op exclusion
// ============================================================================

// --- ListBookings filter correctness ---

func TestListBookings_FilterByBookingStatus(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Seed bookings with different statuses
	for i, status := range []string{"READY_TO_SHIP", "COMPLETED", "READY_TO_SHIP"} {
		b := sampleBooking(fmt.Sprintf("BSN-FS-%d", i), fmt.Sprintf("OSN-FS-%d", i))
		b.BookingStatus = status
		require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b, nil))
	}

	// Filter by READY_TO_SHIP
	result, total, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50, BookingStatus: "READY_TO_SHIP",
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)
	require.Len(t, result, 2)
	for _, b := range result {
		require.Equal(t, "READY_TO_SHIP", b.BookingStatus)
	}
}

func TestListBookings_FilterByMatchStatus(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Seed with different match statuses
	matched := sampleBooking("BSN-MS-1", "OSN-MS-1")
	matched.MatchStatus = "MATCHED"
	unmatched := sampleBooking("BSN-MS-2", "")
	unmatched.MatchStatus = "UNMATCHED"
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, matched, nil))
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, unmatched, nil))

	result, total, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50, MatchStatus: "MATCHED",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, result, 1)
	require.Equal(t, "MATCHED", result[0].MatchStatus)
}

func TestListBookings_SearchByBookingSN(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100,
		sampleBooking("BSN-ALPHA-1", "OSN-1"), nil))
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100,
		sampleBooking("BSN-BETA-2", "OSN-2"), nil))

	result, total, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50, Search: "ALPHA",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, result, 1)
	require.Equal(t, "BSN-ALPHA-1", result[0].BookingSN)
}

func TestListBookings_SearchByOrderSN(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100,
		sampleBooking("BSN-X", "OSN-TARGET-99"), nil))
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100,
		sampleBooking("BSN-Y", "OSN-OTHER"), nil))

	result, _, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50, Search: "TARGET-99",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "OSN-TARGET-99", result[0].OrderSN)
}

func TestListBookings_SearchByRecipientName(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	b1 := sampleBooking("BSN-R1", "OSN-R1")
	b1.RecipientName = "Alice Wonderland"
	b2 := sampleBooking("BSN-R2", "OSN-R2")
	b2.RecipientName = "Bob Builder"
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b1, nil))
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b2, nil))

	result, _, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50, Search: "Alice",
	})
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, "Alice Wonderland", result[0].RecipientName)
}

// --- ListBookings pagination ---

func TestListBookings_Pagination(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Seed 5 bookings
	for i := 0; i < 5; i++ {
		b := sampleBooking("BSN-PG-"+string(rune('A'+i)), "OSN-PG-"+string(rune('A'+i)))
		require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b, nil))
	}

	// Page 1 of 2
	result, total, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(5), total, "total should count all bookings before pagination")
	require.Len(t, result, 2, "page 1 should return 2 items")

	// Page 2 of 2
	result2, total2, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 2, PageSize: 2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(5), total2)
	require.Len(t, result2, 2, "page 2 should return 2 items")

	// Page 3 of 2 (last page, 1 item)
	result3, total3, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 3, PageSize: 2,
	})
	require.NoError(t, err)
	require.Equal(t, int64(5), total3)
	require.Len(t, result3, 1, "page 3 should return 1 item")
}

func TestListBookings_EmptyResult(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	result, total, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50,
	})
	require.NoError(t, err)
	require.Equal(t, int64(0), total)
	require.Empty(t, result)
}

// --- ListBookings has_parent_order flag ---

func TestListBookings_HasParentOrderFlag(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Booking with order_sn → has_parent_order = true
	withParent := sampleBooking("BSN-WP", "OSN-WP")
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, withParent, nil))

	// Booking without order_sn → has_parent_order = false
	withoutParent := sampleBooking("BSN-NP", "")
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, withoutParent, nil))

	result, total, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), total)

	resultMap := make(map[string]Booking)
	for _, b := range result {
		resultMap[b.BookingSN] = b
	}

	require.True(t, resultMap["BSN-WP"].HasParentOrder, "booking with order_sn should have has_parent_order=true")
	require.False(t, resultMap["BSN-NP"].HasParentOrder, "booking without order_sn should have has_parent_order=false")
}

// --- ListBookings item count batch query ---

func TestListBookings_ItemCountBatchQuery(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Booking with 3 items
	b1 := sampleBooking("BSN-IC-1", "OSN-IC-1")
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b1, sampleItems("BSN-IC-1", 3)))

	// Booking with 0 items
	b2 := sampleBooking("BSN-IC-2", "OSN-IC-2")
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b2, nil))

	// Booking with 1 item
	b3 := sampleBooking("BSN-IC-3", "OSN-IC-3")
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b3, sampleItems("BSN-IC-3", 1)))

	result, _, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50,
	})
	require.NoError(t, err)
	require.Len(t, result, 3)

	countMap := make(map[string]int)
	for _, b := range result {
		countMap[b.BookingSN] = b.ItemCount
	}

	require.Equal(t, 3, countMap["BSN-IC-1"], "first booking should have 3 items")
	require.Equal(t, 0, countMap["BSN-IC-2"], "second booking should have 0 items")
	require.Equal(t, 1, countMap["BSN-IC-3"], "third booking should have 1 item")
}

// --- GetBookingDetail with items ---

func TestGetBookingDetail_ReturnsItemsWithCorrectData(t *testing.T) {
	db := setupBookingTestDB(t)
	service := newTestBookingService(db, "tenant_a", 100)
	ctx := context.Background()

	booking := sampleBooking("BSN-DETAIL", "OSN-DETAIL")
	items := sampleItems("BSN-DETAIL", 3)
	require.NoError(t, service.repository.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, items))

	result, resultItems, err := service.GetBookingDetail(ctx, "BSN-DETAIL")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, "BSN-DETAIL", result.BookingSN)
	require.Equal(t, "OSN-DETAIL", result.OrderSN)
	require.Len(t, resultItems, 3)
	for _, item := range resultItems {
		require.NotEmpty(t, item.LineKey)
		require.Equal(t, "BSN-DETAIL", item.BookingSN)
	}
}

func TestGetBookingDetail_NotFound_Integration(t *testing.T) {
	db := setupBookingTestDB(t)
	service := newTestBookingService(db, "tenant_a", 100)
	ctx := context.Background()

	result, _, err := service.GetBookingDetail(ctx, "BSN-NONEXISTENT")
	require.NoError(t, err)
	require.Nil(t, result, "should return nil for non-existent booking")
}

// --- GetBookingDetail parent_order status resolution ---

func TestGetBookingDetail_ParentOrderNotSyncedStatus(t *testing.T) {
	db := setupBookingTestDB(t)
	service := newTestBookingService(db, "tenant_a", 100)
	ctx := context.Background()

	// Booking has order_sn but parent order not in shopee_orders table
	booking := sampleBooking("BSN-NSYNC", "OSN-NOT-IN-DB")
	require.NoError(t, service.repository.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, nil))

	result, _, err := service.GetBookingDetail(ctx, "BSN-NSYNC")
	require.NoError(t, err)
	require.NotNil(t, result)
	require.False(t, result.ParentOrderExists, "parent should not exist when order not synced")
	require.Equal(t, "not_synced", result.ParentOrderStatus, "status should be not_synced")
}

// --- Booking ↔ Order separation audit (no double-counting) ---

func TestBookingOrdersNotInRegularOrderTable(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Sync a booking
	booking := sampleBooking("BSN-SEP-1", "OSN-SEP-1")
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, sampleItems("BSN-SEP-1", 2)))

	// Verify booking is NOT in the shopee_orders table
	var orderCount int64
	require.NoError(t, db.Model(&models.ShopeeOrder{}).
		Where("order_sn = ?", "BSN-SEP-1").
		Count(&orderCount).Error)
	require.Equal(t, int64(0), orderCount, "booking SN must not appear in shopee_orders table")

	// Verify booking items are NOT in any order items table
	var bookingItemCount int64
	require.NoError(t, db.Model(&models.ShopeeBookingItem{}).
		Where("booking_sn = ?", "BSN-SEP-1").
		Count(&bookingItemCount).Error)
	require.Equal(t, int64(2), bookingItemCount, "items should be in shopee_booking_items table")
}

func TestBookingItemsIsolatedFromOrderItems(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Create a booking with items
	booking := sampleBooking("BSN-ISO-ITEMS", "OSN-ISO-ITEMS")
	items := sampleItems("BSN-ISO-ITEMS", 3)
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, items))

	// Create a regular order
	order := models.ShopeeOrder{
		TenantID:    "tenant_a",
		OrderSN:     "OSN-ISO-ITEMS",
		OrderStatus: "COMPLETED",
	}
	require.NoError(t, db.Create(&order).Error)

	// Booking item count stays at 3 (in shopee_booking_items)
	var bookingItems int64
	require.NoError(t, db.Model(&models.ShopeeBookingItem{}).
		Where("booking_sn = ?", "BSN-ISO-ITEMS").
		Count(&bookingItems).Error)
	require.Equal(t, int64(3), bookingItems, "booking items count should be stable")

	// Parent order exists check is separate from booking items
	exists, err := repo.ParentOrderExists(ctx, "tenant_a", "OSN-ISO-ITEMS")
	require.NoError(t, err)
	require.True(t, exists, "parent order should exist independently")
}

// --- Bulk operation exclusion audit ---
// Booking orders are architecturally excluded from bulk operations because:
// 1. BulkShipOrders accepts order_sns (regular orders), not booking_sns
// 2. BulkPrintLabels accepts order_sns (regular orders), not booking_sns
// 3. Booking is a separate domain stored in shopee_bookings table
// 4. Frontend UI separates booking tab from order management tabs

func TestBookingSNsNotAcceptedByOrderValidation(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Sync a booking
	booking := sampleBooking("BSN-BULK-EXCL", "OSN-BULK-EXCL")
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, booking, nil))

	// Verify booking SN is NOT in shopee_orders (where bulk ops look)
	var count int64
	require.NoError(t, db.Model(&models.ShopeeOrder{}).
		Where("order_sn = ?", "BSN-BULK-EXCL").
		Count(&count).Error)
	require.Equal(t, int64(0), count,
		"booking SN must not exist in shopee_orders — bulk ops validate against this table")
}

// --- Cross-tenant list isolation ---

func TestListBookings_CrossTenantIsolation(t *testing.T) {
	db := setupBookingTestDB(t)
	repoA := NewGormBookingRepository(db, "tenant_a")
	repoB := NewGormBookingRepository(db, "tenant_b")
	ctx := context.Background()

	// Seed bookings for tenant_a
	for i := 0; i < 3; i++ {
		b := sampleBooking("BSN-TA-"+string(rune('1'+i)), "OSN-TA-"+string(rune('1'+i)))
		require.NoError(t, repoA.UpsertBookingWithItems(ctx, "tenant_a", 100, b, nil))
	}

	// Seed bookings for tenant_b
	for i := 0; i < 2; i++ {
		b := sampleBooking("BSN-TB-"+string(rune('1'+i)), "OSN-TB-"+string(rune('1'+i)))
		require.NoError(t, repoB.UpsertBookingWithItems(ctx, "tenant_b", 200, b, nil))
	}

	// tenant_a sees only their bookings
	resultA, totalA, err := repoA.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50,
	})
	require.NoError(t, err)
	require.Equal(t, int64(3), totalA)
	require.Len(t, resultA, 3)
	for _, b := range resultA {
		require.Contains(t, b.BookingSN, "BSN-TA-")
	}

	// tenant_b sees only their bookings
	resultB, totalB, err := repoB.ListBookings(ctx, "tenant_b", BookingListParams{
		Page: 1, PageSize: 50,
	})
	require.NoError(t, err)
	require.Equal(t, int64(2), totalB)
	require.Len(t, resultB, 2)
	for _, b := range resultB {
		require.Contains(t, b.BookingSN, "BSN-TB-")
	}
}

// --- ListBookings combined filters ---

func TestListBookings_CombinedFilters(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Seed with varied data
	b1 := sampleBooking("BSN-CF-1", "OSN-CF-1")
	b1.BookingStatus = "READY_TO_SHIP"
	b1.MatchStatus = "MATCHED"
	b1.RecipientName = "Alice"

	b2 := sampleBooking("BSN-CF-2", "OSN-CF-2")
	b2.BookingStatus = "READY_TO_SHIP"
	b2.MatchStatus = "UNMATCHED"
	b2.RecipientName = "Bob"

	b3 := sampleBooking("BSN-CF-3", "OSN-CF-3")
	b3.BookingStatus = "COMPLETED"
	b3.MatchStatus = "MATCHED"
	b3.RecipientName = "Charlie"

	for _, b := range []Booking{b1, b2, b3} {
		require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b, nil))
	}

	// Filter: READY_TO_SHIP + MATCHED → should match only b1
	result, total, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50, BookingStatus: "READY_TO_SHIP", MatchStatus: "MATCHED",
	})
	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, result, 1)
	require.Equal(t, "BSN-CF-1", result[0].BookingSN)
}

// --- ListBookings ordering ---

func TestListBookings_OrderedByCreateTimeDesc(t *testing.T) {
	db := setupBookingTestDB(t)
	repo := NewGormBookingRepository(db, "tenant_a")
	ctx := context.Background()

	// Seed with different create times
	b1 := sampleBooking("BSN-OLD", "OSN-OLD")
	b1.CreateTime = 1710000000
	b2 := sampleBooking("BSN-NEW", "OSN-NEW")
	b2.CreateTime = 1710100000

	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b1, nil))
	require.NoError(t, repo.UpsertBookingWithItems(ctx, "tenant_a", 100, b2, nil))

	result, _, err := repo.ListBookings(ctx, "tenant_a", BookingListParams{
		Page: 1, PageSize: 50,
	})
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, "BSN-NEW", result[0].BookingSN, "newest booking should come first")
	require.Equal(t, "BSN-OLD", result[1].BookingSN, "oldest booking should come last")
}
