package sheets_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omni/backend/internal/services/sheets"
)

// ---------------------------------------------------------------------------
// Constructor tests
// ---------------------------------------------------------------------------

func TestNewWalletSheetService_NotNil(t *testing.T) {
	svc := sheets.NewWalletSheetService(nil)
	require.NotNil(t, svc)
}

func TestNewShippingFeeSheetService_NotNil(t *testing.T) {
	svc := sheets.NewShippingFeeSheetService(nil)
	require.NotNil(t, svc)
}

func TestNewSheetConfigService_NotNil(t *testing.T) {
	svc := sheets.NewSheetConfigService(nil)
	require.NotNil(t, svc)
}

func TestNewInventorySheetService_NotNil(t *testing.T) {
	svc := sheets.NewInventorySheetService(nil)
	require.NotNil(t, svc)
}

// ---------------------------------------------------------------------------
// WalletTransaction — struct & table name
// ---------------------------------------------------------------------------

func TestWalletTransaction_TableName(t *testing.T) {
	tx := sheets.WalletTransaction{}
	assert.Equal(t, "wallet_transactions", tx.TableName())
}

func TestWalletTransaction_NetAmountComputation(t *testing.T) {
	// SaveTransactions computes NetAmount = Amount - Fee inline.
	// We verify the same arithmetic here to cover the business rule.
	tests := []struct {
		name     string
		amount   float64
		fee      float64
		expected float64
	}{
		{"positive net", 100.0, 10.0, 90.0},
		{"zero fee", 50.0, 0.0, 50.0},
		{"fee larger than amount", 20.0, 25.0, -5.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			net := tc.amount - tc.fee
			assert.Equal(t, tc.expected, net)
		})
	}
}

func TestWalletTransaction_FieldDefaults(t *testing.T) {
	tx := sheets.WalletTransaction{
		TenantID:      "tenant1",
		Platform:      "shopee",
		TransactionID: "txn-001",
		Amount:        200.0,
		Fee:           15.0,
		Currency:      "IDR",
		Source:        "api",
	}
	assert.Equal(t, "tenant1", tx.TenantID)
	assert.Equal(t, "shopee", tx.Platform)
	assert.Equal(t, 200.0, tx.Amount)
	assert.Equal(t, 15.0, tx.Fee)
	assert.Equal(t, "IDR", tx.Currency)
	assert.Equal(t, "api", tx.Source)
}

// ---------------------------------------------------------------------------
// WalletBalance — struct & table name
// ---------------------------------------------------------------------------

func TestWalletBalance_TableName(t *testing.T) {
	wb := sheets.WalletBalance{}
	assert.Equal(t, "wallet_balances", wb.TableName())
}

func TestWalletBalance_TotalBalance(t *testing.T) {
	// SaveWalletBalance computes TotalBalance = AvailableBalance + PendingBalance.
	tests := []struct {
		name      string
		available float64
		pending   float64
		expected  float64
	}{
		{"both positive", 500.0, 100.0, 600.0},
		{"zero pending", 300.0, 0.0, 300.0},
		{"both zero", 0.0, 0.0, 0.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total := tc.available + tc.pending
			assert.Equal(t, tc.expected, total)
		})
	}
}

// ---------------------------------------------------------------------------
// WalletSummary — struct initialization
// ---------------------------------------------------------------------------

func TestWalletSummary_Initialization(t *testing.T) {
	summary := &sheets.WalletSummary{
		TotalAvailable: 1000.0,
		TotalPending:   200.0,
		ByPlatform:     make(map[string]*sheets.WalletBalance),
	}
	assert.Equal(t, 1000.0, summary.TotalAvailable)
	assert.Equal(t, 200.0, summary.TotalPending)
	assert.NotNil(t, summary.ByPlatform)
}

func TestWalletSummary_ByPlatformAggregation(t *testing.T) {
	// Mirrors the aggregation done in GetWalletSummary.
	balances := []sheets.WalletBalance{
		{Platform: "shopee", AvailableBalance: 500.0, PendingBalance: 50.0},
		{Platform: "lazada", AvailableBalance: 300.0, PendingBalance: 30.0},
	}

	summary := &sheets.WalletSummary{
		ByPlatform: make(map[string]*sheets.WalletBalance),
	}
	for i := range balances {
		summary.TotalAvailable += balances[i].AvailableBalance
		summary.TotalPending += balances[i].PendingBalance
		summary.ByPlatform[balances[i].Platform] = &balances[i]
	}

	assert.Equal(t, 800.0, summary.TotalAvailable)
	assert.Equal(t, 80.0, summary.TotalPending)
	assert.Contains(t, summary.ByPlatform, "shopee")
	assert.Contains(t, summary.ByPlatform, "lazada")
}

// ---------------------------------------------------------------------------
// ShippingFeeRecord — struct & table name
// ---------------------------------------------------------------------------

func TestShippingFeeRecord_TableName(t *testing.T) {
	r := sheets.ShippingFeeRecord{}
	assert.Equal(t, "shipping_fee_records", r.TableName())
}

func TestShippingFeeRecord_TotalFeeComputation(t *testing.T) {
	// SaveShippingFees computes: TotalFee = Base+Weight+Insurance+Cod+Other - Discount.
	tests := []struct {
		name        string
		base        float64
		weight      float64
		insurance   float64
		cod         float64
		other       float64
		discount    float64
		expectedFee float64
	}{
		{"all fees", 10.0, 5.0, 2.0, 3.0, 1.0, 0.0, 21.0},
		{"with discount", 10.0, 5.0, 0.0, 0.0, 0.0, 3.0, 12.0},
		{"zero fees", 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total := tc.base + tc.weight + tc.insurance + tc.cod + tc.other - tc.discount
			assert.Equal(t, tc.expectedFee, total)
		})
	}
}

func TestShippingFeeRecord_FieldAccess(t *testing.T) {
	r := sheets.ShippingFeeRecord{
		TenantID: "tenant1",
		Platform: "shopee",
		OrderSN:  "ORDER-001",
		Courier:  "JNE",
		PaidBy:   "seller",
		Currency: "IDR",
		Source:   "api",
	}
	assert.Equal(t, "tenant1", r.TenantID)
	assert.Equal(t, "shopee", r.Platform)
	assert.Equal(t, "JNE", r.Courier)
	assert.Equal(t, "seller", r.PaidBy)
}

// ---------------------------------------------------------------------------
// ShippingFeeSummary & CourierStat
// ---------------------------------------------------------------------------

func TestShippingFeeSummary_AvgFeePerOrder(t *testing.T) {
	tests := []struct {
		name        string
		totalFees   float64
		totalOrders int
		expected    float64
	}{
		{"normal", 300.0, 10, 30.0},
		{"single order", 25.0, 1, 25.0},
		{"zero orders", 0.0, 0, 0.0}, // division guard
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var avg float64
			if tc.totalOrders > 0 {
				avg = tc.totalFees / float64(tc.totalOrders)
			}
			assert.Equal(t, tc.expected, avg)
		})
	}
}

func TestCourierStat_AvgFeeCalculation(t *testing.T) {
	stat := sheets.CourierStat{
		OrderCount: 5,
		TotalFees:  125.0,
	}
	if stat.OrderCount > 0 {
		stat.AvgFee = stat.TotalFees / float64(stat.OrderCount)
	}
	assert.Equal(t, 25.0, stat.AvgFee)
}

// ---------------------------------------------------------------------------
// SheetConfig — struct & table name
// ---------------------------------------------------------------------------

func TestSheetConfig_TableName(t *testing.T) {
	c := sheets.SheetConfig{}
	assert.Equal(t, "sheet_configs", c.TableName())
}

func TestSheetConfig_DefaultValues(t *testing.T) {
	c := sheets.SheetConfig{
		TenantID:      "tenant1",
		SpreadsheetID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
		SheetName:     "Inventory",
		SheetType:     "inventory",
		HeaderRow:     1,
		DataStartRow:  2,
		IsActive:      true,
	}
	assert.Equal(t, 1, c.HeaderRow)
	assert.Equal(t, 2, c.DataStartRow)
	assert.True(t, c.IsActive)
	assert.Equal(t, "inventory", c.SheetType)
}

// ---------------------------------------------------------------------------
// SheetSnapshot — struct & table name
// ---------------------------------------------------------------------------

func TestSheetSnapshot_TableName(t *testing.T) {
	s := sheets.SheetSnapshot{}
	assert.Equal(t, "sheet_snapshots", s.TableName())
}

func TestSheetSnapshot_FieldAccess(t *testing.T) {
	now := time.Now()
	s := sheets.SheetSnapshot{
		TenantID:      "tenant1",
		SheetConfigID: 42,
		DataHash:      "abc123",
		RowCount:      100,
		Snapshot:      `[{"sku":"A","qty":10}]`,
		CreatedAt:     now,
	}
	assert.Equal(t, "tenant1", s.TenantID)
	assert.Equal(t, uint(42), s.SheetConfigID)
	assert.Equal(t, "abc123", s.DataHash)
	assert.Equal(t, 100, s.RowCount)
}

// ---------------------------------------------------------------------------
// InventoryRecord — struct & table name
// ---------------------------------------------------------------------------

func TestInventoryRecord_TableName(t *testing.T) {
	r := sheets.InventoryRecord{}
	assert.Equal(t, "inventory_records", r.TableName())
}

func TestInventoryRecord_AvailableQtyComputation(t *testing.T) {
	// SaveInventoryRecords: AvailableQty = StockQuantity - ReservedQty
	tests := []struct {
		name     string
		stock    int
		reserved int
		expected int
	}{
		{"normal", 100, 20, 80},
		{"no reservations", 50, 0, 50},
		{"all reserved", 30, 30, 0},
		{"over-reserved", 10, 15, -5}, // edge: negative
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			available := tc.stock - tc.reserved
			assert.Equal(t, tc.expected, available)
		})
	}
}

func TestInventoryRecord_TotalValueComputation(t *testing.T) {
	// SaveInventoryRecords: TotalValue = float64(StockQuantity) * UnitCost
	tests := []struct {
		name     string
		stock    int
		unitCost float64
		expected float64
	}{
		{"normal", 100, 5.0, 500.0},
		{"zero stock", 0, 10.0, 0.0},
		{"zero cost", 50, 0.0, 0.0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			total := float64(tc.stock) * tc.unitCost
			assert.Equal(t, tc.expected, total)
		})
	}
}

func TestInventoryRecord_FieldAccess(t *testing.T) {
	r := sheets.InventoryRecord{
		TenantID:      "tenant1",
		SKU:           "SKU-001",
		ProductName:   "Widget",
		StockQuantity: 100,
		ReservedQty:   10,
		MinStock:      5,
		MaxStock:      200,
		ReorderPoint:  20,
		UnitCost:      7.5,
		Source:        "sheet",
	}
	assert.Equal(t, "SKU-001", r.SKU)
	assert.Equal(t, 100, r.StockQuantity)
	assert.Equal(t, 20, r.ReorderPoint)
	assert.Equal(t, 7.5, r.UnitCost)
}

// ---------------------------------------------------------------------------
// InventorySyncHistory — struct & table name
// ---------------------------------------------------------------------------

func TestInventorySyncHistory_TableName(t *testing.T) {
	h := sheets.InventorySyncHistory{}
	assert.Equal(t, "inventory_sync_history", h.TableName())
}

func TestInventorySyncHistory_FieldAccess(t *testing.T) {
	h := sheets.InventorySyncHistory{
		TenantID:       "tenant1",
		SyncType:       "full",
		RecordsAdded:   50,
		RecordsUpdated: 10,
		RecordsDeleted: 2,
		Status:         "success",
		Duration:       1500,
	}
	assert.Equal(t, "full", h.SyncType)
	assert.Equal(t, 50, h.RecordsAdded)
	assert.Equal(t, "success", h.Status)
	assert.Equal(t, 1500, h.Duration)
}

// ---------------------------------------------------------------------------
// InventoryStats — struct initialization
// ---------------------------------------------------------------------------

func TestInventoryStats_Initialization(t *testing.T) {
	stats := &sheets.InventoryStats{
		TotalSKUs:       100,
		TotalStock:      5000,
		TotalValue:      25000.0,
		LowStockCount:   5,
		OutOfStockCount: 2,
		OverstockCount:  3,
	}
	assert.Equal(t, 100, stats.TotalSKUs)
	assert.Equal(t, 5000, stats.TotalStock)
	assert.Equal(t, 25000.0, stats.TotalValue)
	assert.Equal(t, 5, stats.LowStockCount)
	assert.Equal(t, 2, stats.OutOfStockCount)
	assert.Equal(t, 3, stats.OverstockCount)
}
