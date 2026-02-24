package orders

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupOrderTodayTestDB creates an in-memory SQLite DB with the required schema
func setupOrderTodayTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	err = db.AutoMigrate(&OrderTodayItem{})
	require.NoError(t, err, "failed to migrate order_today_items table")

	return db
}

// --- Constructor ---

func TestNewOrderTodayService_ReturnsNonNil(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	svc := NewOrderTodayService(db)
	assert.NotNil(t, svc)
}

// --- OrderTodayItem.TableName ---

func TestOrderTodayItem_TableName(t *testing.T) {
	item := OrderTodayItem{}
	assert.Equal(t, "order_today_items", item.TableName())
}

// --- SaveOrderTodayItems ---

func TestOrderTodayService_SaveOrderTodayItems_EmptySlice(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	n, err := svc.SaveOrderTodayItems(ctx, "tenant1", []OrderTodayData{})
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestOrderTodayService_SaveOrderTodayItems_SingleItem(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	items := []OrderTodayData{
		{
			Platform:      "shopee",
			OrderSN:       "SN001",
			TrackingNo:    "TRACK001",
			Courier:       "J&T",
			SellerSku:     "SKU-A",
			ProductName:   "Product A",
			VariationName: "Red",
			Quantity:      2,
		},
	}

	n, err := svc.SaveOrderTodayItems(ctx, "tenant1", items)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestOrderTodayService_SaveOrderTodayItems_DeduplicatesDuplicateKeys(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	// Two items with same platform+order_sn+seller_sku → should be deduplicated and quantities aggregated
	items := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 2},
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 3},
	}

	n, err := svc.SaveOrderTodayItems(ctx, "tenant1", items)
	assert.NoError(t, err)
	assert.Equal(t, 1, n) // deduplicated to 1

	// Verify the quantity was aggregated (2 + 3 = 5)
	saved, err := svc.GetOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)
	require.Len(t, saved, 1)
	assert.Equal(t, 5, saved[0].Quantity)
}

func TestOrderTodayService_SaveOrderTodayItems_DifferentKeys_NotDeduplicated(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	items := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1},
		{Platform: "lazada", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1}, // different platform
		{Platform: "shopee", OrderSN: "SN002", SellerSku: "SKU-A", Quantity: 1}, // different order
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-B", Quantity: 1}, // different sku
	}

	n, err := svc.SaveOrderTodayItems(ctx, "tenant1", items)
	assert.NoError(t, err)
	assert.Equal(t, 4, n)
}

func TestOrderTodayService_SaveOrderTodayItems_ZeroQuantityDefaultsToOne(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	items := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 0},
	}

	n, err := svc.SaveOrderTodayItems(ctx, "tenant1", items)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)

	saved, err := svc.GetOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)
	require.Len(t, saved, 1)
	assert.Equal(t, 1, saved[0].Quantity) // defaulted to 1
}

func TestOrderTodayService_SaveOrderTodayItems_NegativeQuantityDefaultsToOne(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	items := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: -5},
	}

	n, err := svc.SaveOrderTodayItems(ctx, "tenant1", items)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)

	saved, err := svc.GetOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)
	require.Len(t, saved, 1)
	assert.Equal(t, 1, saved[0].Quantity)
}

func TestOrderTodayService_SaveOrderTodayItems_ReplacesExistingForTenant(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	// First save: 3 items
	first := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1},
		{Platform: "shopee", OrderSN: "SN002", SellerSku: "SKU-B", Quantity: 1},
		{Platform: "lazada", OrderSN: "LZ001", SellerSku: "SKU-C", Quantity: 1},
	}
	n1, err := svc.SaveOrderTodayItems(ctx, "tenant1", first)
	assert.NoError(t, err)
	assert.Equal(t, 3, n1)

	// Second save: 1 item — must replace, not append
	second := []OrderTodayData{
		{Platform: "tiktok", OrderSN: "TT001", SellerSku: "SKU-D", Quantity: 2},
	}
	n2, err := svc.SaveOrderTodayItems(ctx, "tenant1", second)
	assert.NoError(t, err)
	assert.Equal(t, 1, n2)

	// Only the 1 new item should remain
	all, err := svc.GetOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, "tiktok", all[0].Platform)
}

func TestOrderTodayService_SaveOrderTodayItems_TenantIsolation(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	// Save for tenant1
	itemsT1 := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN-T1", SellerSku: "SKU-A", Quantity: 1},
	}
	_, err := svc.SaveOrderTodayItems(ctx, "tenant1", itemsT1)
	require.NoError(t, err)

	// Save for tenant2
	itemsT2 := []OrderTodayData{
		{Platform: "lazada", OrderSN: "SN-T2", SellerSku: "SKU-B", Quantity: 2},
	}
	_, err = svc.SaveOrderTodayItems(ctx, "tenant2", itemsT2)
	require.NoError(t, err)

	// Replace tenant1 items with empty — tenant2 should be unaffected
	_, err = svc.SaveOrderTodayItems(ctx, "tenant1", []OrderTodayData{})
	require.NoError(t, err)

	t2Items, err := svc.GetOrderTodayItems(ctx, "tenant2")
	assert.NoError(t, err)
	assert.Len(t, t2Items, 1)
}

// --- GetOrderTodayItems ---

func TestOrderTodayService_GetOrderTodayItems_Empty(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	items, err := svc.GetOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Empty(t, items)
}

func TestOrderTodayService_GetOrderTodayItems_ReturnsSavedData(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	input := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", ProductName: "Prod A", Quantity: 3},
		{Platform: "lazada", OrderSN: "LZ001", SellerSku: "SKU-B", ProductName: "Prod B", Quantity: 1},
	}
	_, err := svc.SaveOrderTodayItems(ctx, "tenant1", input)
	require.NoError(t, err)

	items, err := svc.GetOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Len(t, items, 2)

	// Verify tenant ID is set
	for _, it := range items {
		assert.Equal(t, "tenant1", it.TenantID)
	}
}

// --- GetOrderTodayCount ---

func TestOrderTodayService_GetOrderTodayCount_Empty(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	count, err := svc.GetOrderTodayCount(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestOrderTodayService_GetOrderTodayCount_CorrectCount(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	input := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1},
		{Platform: "shopee", OrderSN: "SN002", SellerSku: "SKU-B", Quantity: 1},
		{Platform: "lazada", OrderSN: "LZ001", SellerSku: "SKU-C", Quantity: 1},
	}
	_, err := svc.SaveOrderTodayItems(ctx, "tenant1", input)
	require.NoError(t, err)

	count, err := svc.GetOrderTodayCount(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, int64(3), count)
}

func TestOrderTodayService_GetOrderTodayCount_TenantIsolation(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	input := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1},
	}
	_, err := svc.SaveOrderTodayItems(ctx, "tenant1", input)
	require.NoError(t, err)

	count, err := svc.GetOrderTodayCount(ctx, "tenant2")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// --- GetPlatformCounts ---

func TestOrderTodayService_GetPlatformCounts_Empty(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	counts, err := svc.GetPlatformCounts(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Empty(t, counts)
}

func TestOrderTodayService_GetPlatformCounts_MultiPlatform(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	// 2 shopee orders (distinct SNs), 1 lazada order
	input := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1},
		{Platform: "shopee", OrderSN: "SN002", SellerSku: "SKU-B", Quantity: 1},
		{Platform: "lazada", OrderSN: "LZ001", SellerSku: "SKU-C", Quantity: 1},
	}
	_, err := svc.SaveOrderTodayItems(ctx, "tenant1", input)
	require.NoError(t, err)

	counts, err := svc.GetPlatformCounts(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, int64(2), counts["shopee"])
	assert.Equal(t, int64(1), counts["lazada"])
}

func TestOrderTodayService_GetPlatformCounts_SameOrderSN_CountedOnce(t *testing.T) {
	// Two rows with same platform+order_sn but different SKUs should count as 1 distinct order
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	// Insert rows directly (bypassing dedup) to test COUNT(DISTINCT order_sn) behavior
	rows := []OrderTodayItem{
		{TenantID: "tenant1", Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1},
		{TenantID: "tenant1", Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-B", Quantity: 2},
	}
	err := db.Create(&rows).Error
	require.NoError(t, err)

	counts, err := svc.GetPlatformCounts(ctx, "tenant1")
	assert.NoError(t, err)
	// COUNT(DISTINCT order_sn) = 1, not 2
	assert.Equal(t, int64(1), counts["shopee"])
}

// --- ClearOrderTodayItems ---

func TestOrderTodayService_ClearOrderTodayItems_EmptyIsNoError(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	err := svc.ClearOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)
}

func TestOrderTodayService_ClearOrderTodayItems_RemovesItems(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	input := []OrderTodayData{
		{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1},
	}
	_, err := svc.SaveOrderTodayItems(ctx, "tenant1", input)
	require.NoError(t, err)

	err = svc.ClearOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)

	count, err := svc.GetOrderTodayCount(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestOrderTodayService_ClearOrderTodayItems_TenantIsolation(t *testing.T) {
	db := setupOrderTodayTestDB(t)
	ctx := context.Background()
	svc := NewOrderTodayService(db)

	itemsT1 := []OrderTodayData{{Platform: "shopee", OrderSN: "SN001", SellerSku: "SKU-A", Quantity: 1}}
	itemsT2 := []OrderTodayData{{Platform: "shopee", OrderSN: "SN002", SellerSku: "SKU-B", Quantity: 1}}
	_, err := svc.SaveOrderTodayItems(ctx, "tenant1", itemsT1)
	require.NoError(t, err)
	_, err = svc.SaveOrderTodayItems(ctx, "tenant2", itemsT2)
	require.NoError(t, err)

	err = svc.ClearOrderTodayItems(ctx, "tenant1")
	assert.NoError(t, err)

	// tenant2 must be untouched
	count, err := svc.GetOrderTodayCount(ctx, "tenant2")
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count)
}
