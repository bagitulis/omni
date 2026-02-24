package orders

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupLockedOrderTestDB creates an in-memory SQLite DB with the required schema
func setupLockedOrderTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	err = db.AutoMigrate(&LockedOrderItem{})
	require.NoError(t, err, "failed to migrate locked_orders table")

	return db
}

// --- Constructor ---

func TestNewLockedOrderService_ReturnsNonNil(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	svc := NewLockedOrderService(db)
	assert.NotNil(t, svc)
}

// --- LockedOrderItem.TableName ---

func TestLockedOrderItem_TableName(t *testing.T) {
	item := LockedOrderItem{}
	assert.Equal(t, "locked_orders", item.TableName())
}

// --- SaveLockedOrders ---

func TestLockedOrderService_SaveLockedOrders_EmptySlice(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	n, err := svc.SaveLockedOrders(ctx, "tenant1", []LockedOrderItem{})
	assert.NoError(t, err)
	assert.Equal(t, 0, n)
}

func TestLockedOrderService_SaveLockedOrders_SingleItem(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	items := []LockedOrderItem{
		{SKU: "SKU-A", ProductName: "Product A", VariationName: "Red", Qty: 3},
	}

	n, err := svc.SaveLockedOrders(ctx, "tenant1", items)
	assert.NoError(t, err)
	assert.Equal(t, 1, n)
}

func TestLockedOrderService_SaveLockedOrders_AssignsUUIDAndTenantID(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	items := []LockedOrderItem{
		{SKU: "SKU-A", ProductName: "Product A", Qty: 2},
		{SKU: "SKU-B", ProductName: "Product B", Qty: 5},
	}

	n, err := svc.SaveLockedOrders(ctx, "tenant1", items)
	require.NoError(t, err)
	assert.Equal(t, 2, n)

	// Verify IDs and TenantIDs are set
	saved, err := svc.GetLockedOrders(ctx, "tenant1")
	require.NoError(t, err)
	require.Len(t, saved, 2)
	for _, item := range saved {
		assert.NotEmpty(t, item.ID, "ID must be set (UUID)")
		assert.Equal(t, "tenant1", item.TenantID)
	}

	// IDs must be unique UUIDs
	assert.NotEqual(t, saved[0].ID, saved[1].ID)
}

func TestLockedOrderService_SaveLockedOrders_ReplacesExistingForTenant(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	first := []LockedOrderItem{
		{SKU: "SKU-A", ProductName: "Product A", Qty: 2},
		{SKU: "SKU-B", ProductName: "Product B", Qty: 3},
	}
	n1, err := svc.SaveLockedOrders(ctx, "tenant1", first)
	require.NoError(t, err)
	assert.Equal(t, 2, n1)

	// Second save replaces
	second := []LockedOrderItem{
		{SKU: "SKU-C", ProductName: "Product C", Qty: 10},
	}
	n2, err := svc.SaveLockedOrders(ctx, "tenant1", second)
	require.NoError(t, err)
	assert.Equal(t, 1, n2)

	all, err := svc.GetLockedOrders(ctx, "tenant1")
	require.NoError(t, err)
	assert.Len(t, all, 1)
	assert.Equal(t, "SKU-C", all[0].SKU)
}

func TestLockedOrderService_SaveLockedOrders_TenantIsolation(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	itemsT1 := []LockedOrderItem{{SKU: "SKU-T1", ProductName: "P1", Qty: 1}}
	itemsT2 := []LockedOrderItem{{SKU: "SKU-T2", ProductName: "P2", Qty: 2}}

	_, err := svc.SaveLockedOrders(ctx, "tenant1", itemsT1)
	require.NoError(t, err)
	_, err = svc.SaveLockedOrders(ctx, "tenant2", itemsT2)
	require.NoError(t, err)

	// Replace tenant1 with empty — tenant2 stays
	_, err = svc.SaveLockedOrders(ctx, "tenant1", []LockedOrderItem{})
	require.NoError(t, err)

	t2Items, err := svc.GetLockedOrders(ctx, "tenant2")
	require.NoError(t, err)
	assert.Len(t, t2Items, 1)
	assert.Equal(t, "SKU-T2", t2Items[0].SKU)
}

// --- GetLockedOrders ---

func TestLockedOrderService_GetLockedOrders_Empty(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	orders, err := svc.GetLockedOrders(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Empty(t, orders)
}

func TestLockedOrderService_GetLockedOrders_ReturnsSavedData(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	items := []LockedOrderItem{
		{SKU: "SKU-A", ProductName: "Product A", Qty: 5},
		{SKU: "SKU-B", ProductName: "Product B", Qty: 2},
	}
	_, err := svc.SaveLockedOrders(ctx, "tenant1", items)
	require.NoError(t, err)

	orders, err := svc.GetLockedOrders(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Len(t, orders, 2)

	for _, o := range orders {
		assert.Equal(t, "tenant1", o.TenantID)
	}
}

func TestLockedOrderService_GetLockedOrders_OrderedByQtyDesc(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	items := []LockedOrderItem{
		{SKU: "SKU-A", ProductName: "Low Qty", Qty: 1},
		{SKU: "SKU-B", ProductName: "High Qty", Qty: 100},
		{SKU: "SKU-C", ProductName: "Mid Qty", Qty: 10},
	}
	_, err := svc.SaveLockedOrders(ctx, "tenant1", items)
	require.NoError(t, err)

	orders, err := svc.GetLockedOrders(ctx, "tenant1")
	require.NoError(t, err)
	require.Len(t, orders, 3)

	// Should be ordered by qty DESC: 100, 10, 1
	assert.Equal(t, 100, orders[0].Qty)
	assert.Equal(t, 10, orders[1].Qty)
	assert.Equal(t, 1, orders[2].Qty)
}

// --- GetTotalQty ---

func TestLockedOrderService_GetTotalQty_Empty(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	total, err := svc.GetTotalQty(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

func TestLockedOrderService_GetTotalQty_SumsAllQty(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	items := []LockedOrderItem{
		{SKU: "SKU-A", ProductName: "Product A", Qty: 5},
		{SKU: "SKU-B", ProductName: "Product B", Qty: 3},
		{SKU: "SKU-C", ProductName: "Product C", Qty: 7},
	}
	_, err := svc.SaveLockedOrders(ctx, "tenant1", items)
	require.NoError(t, err)

	total, err := svc.GetTotalQty(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, int64(15), total) // 5+3+7
}

func TestLockedOrderService_GetTotalQty_TenantIsolation(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	itemsT1 := []LockedOrderItem{{SKU: "SKU-A", ProductName: "P1", Qty: 10}}
	itemsT2 := []LockedOrderItem{{SKU: "SKU-B", ProductName: "P2", Qty: 999}}

	_, err := svc.SaveLockedOrders(ctx, "tenant1", itemsT1)
	require.NoError(t, err)
	_, err = svc.SaveLockedOrders(ctx, "tenant2", itemsT2)
	require.NoError(t, err)

	total, err := svc.GetTotalQty(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, int64(10), total) // tenant2's 999 must not bleed in
}

// --- ClearLockedOrders ---

func TestLockedOrderService_ClearLockedOrders_EmptyIsNoError(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	err := svc.ClearLockedOrders(ctx, "tenant1")
	assert.NoError(t, err)
}

func TestLockedOrderService_ClearLockedOrders_RemovesItems(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	items := []LockedOrderItem{
		{SKU: "SKU-A", ProductName: "Product A", Qty: 5},
	}
	_, err := svc.SaveLockedOrders(ctx, "tenant1", items)
	require.NoError(t, err)

	err = svc.ClearLockedOrders(ctx, "tenant1")
	assert.NoError(t, err)

	total, err := svc.GetTotalQty(ctx, "tenant1")
	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
}

func TestLockedOrderService_ClearLockedOrders_TenantIsolation(t *testing.T) {
	db := setupLockedOrderTestDB(t)
	ctx := context.Background()
	svc := NewLockedOrderService(db)

	itemsT1 := []LockedOrderItem{{SKU: "SKU-A", ProductName: "P1", Qty: 1}}
	itemsT2 := []LockedOrderItem{{SKU: "SKU-B", ProductName: "P2", Qty: 2}}

	_, err := svc.SaveLockedOrders(ctx, "tenant1", itemsT1)
	require.NoError(t, err)
	_, err = svc.SaveLockedOrders(ctx, "tenant2", itemsT2)
	require.NoError(t, err)

	err = svc.ClearLockedOrders(ctx, "tenant1")
	assert.NoError(t, err)

	// tenant2 must be untouched
	t2orders, err := svc.GetLockedOrders(ctx, "tenant2")
	require.NoError(t, err)
	assert.Len(t, t2orders, 1)
}
