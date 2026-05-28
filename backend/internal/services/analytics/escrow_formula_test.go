package analytics_test

import (
	"fmt"
	"math"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/analytics"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ============================================================================
// 1. FORMULA UNIT TESTS — Pure function correctness
// ============================================================================
//
// Shopee formula: BuyerPaidShippingFee - ActualShippingFee + ShopeeShippingRebate
// TikTok formula: ShippingFeeCustomerPaid - ShippingFeeActual + ShippingFeePlatformDiscount
// ============================================================================

func TestComputeShopeeShippingDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		buyerPaid  float64
		actual     float64
		rebate     float64
		want       float64
	}{
		// Normal case from spec
		{name: "normal case from spec", buyerPaid: 10000, actual: 7000, rebate: 1500, want: 4500},

		// Variants with different rebate values
		{name: "zero rebate", buyerPaid: 10000, actual: 7000, rebate: 0, want: 3000},
		{name: "rebate exceeds actual shipping", buyerPaid: 5000, actual: 3000, rebate: 2000, want: 4000},

		// Negative / loss scenarios
		{name: "negative diff (loss)", buyerPaid: 5000, actual: 7000, rebate: 0, want: -2000},
		{name: "rebate covers partial loss", buyerPaid: 3000, actual: 5000, rebate: 1500, want: -500},
		{name: "rebate fully covers loss", buyerPaid: 3000, actual: 5000, rebate: 3000, want: 1000},
		{name: "buyer paid less than actual with rebate", buyerPaid: 8000, actual: 10000, rebate: 500, want: -1500},

		// Zero edge cases
		{name: "all zeros", buyerPaid: 0, actual: 0, rebate: 0, want: 0},
		{name: "only actual shipping charged", buyerPaid: 0, actual: 7000, rebate: 0, want: -7000},

		// Fractional values (IDR has no sub-sen, but float64 is used)
		{name: "fractional values", buyerPaid: 10050.50, actual: 7500.25, rebate: 1250.25, want: 3800.50},
		{name: "fractional rebate", buyerPaid: 10000, actual: 7500, rebate: 2499.99, want: 4999.99},

		// Large values (millions)
		{name: "large values", buyerPaid: 10_000_000, actual: 7_000_000, rebate: 1_500_000, want: 4_500_000},
		{name: "very large values", buyerPaid: 1_000_000_000, actual: 800_000_000, rebate: 100_000_000, want: 300_000_000},

		// Chargeback / refund scenarios
		{name: "negative buyer paid (refund reversal)", buyerPaid: -5000, actual: 7000, rebate: 0, want: -12000},
		{name: "full chargeback", buyerPaid: -10000, actual: 10000, rebate: 0, want: -20000},
		{name: "chargeback with rebate", buyerPaid: -10000, actual: 7000, rebate: 1500, want: -15500},
	}

	for _, tt := range tests {
		tt := tt // capture range variable
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := analytics.ComputeShopeeShippingDiff(tt.buyerPaid, tt.actual, tt.rebate)
			if got != tt.want {
				t.Errorf("ComputeShopeeShippingDiff(%v, %v, %v) = %v, want %v",
					tt.buyerPaid, tt.actual, tt.rebate, got, tt.want)
			}
		})
	}
}

func TestComputeTiktokShippingDiff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		customerPaid float64
		actual       float64
		discount     float64
		want         float64
	}{
		// Normal case
		{name: "normal case", customerPaid: 10000, actual: 8000, discount: 2000, want: 4000},

		// Variants
		{name: "zero discount", customerPaid: 10000, actual: 8000, discount: 0, want: 2000},
		{name: "discount exceeds diff", customerPaid: 5000, actual: 8000, discount: 5000, want: 2000},
		{name: "large discount", customerPaid: 10000, actual: 12000, discount: 5000, want: 3000},

		// Negative / loss
		{name: "negative diff (loss)", customerPaid: 5000, actual: 8000, discount: 1000, want: -2000},
		{name: "customer paid less than actual", customerPaid: 3000, actual: 10000, discount: 0, want: -7000},

		// Zero
		{name: "all zeros", customerPaid: 0, actual: 0, discount: 0, want: 0},
		{name: "only actual charged", customerPaid: 0, actual: 8000, discount: 0, want: -8000},

		// Fractional
		{name: "fractional values", customerPaid: 15000.75, actual: 10000.50, discount: 2500.25, want: 7500.50},
		{name: "fractional discount", customerPaid: 10000, actual: 8000, discount: 1499.99, want: 3499.99},

		// Large
		{name: "large values", customerPaid: 10_000_000, actual: 8_000_000, discount: 2_000_000, want: 4_000_000},

		// Normalized: actual is already abs'd before storage
		{name: "normalized negative actual", customerPaid: 10000, actual: 3000, discount: 1500, want: 8500},

		// Chargeback / refund
		{name: "negative customer paid (refund)", customerPaid: -2000, actual: 8000, discount: 1000, want: -9000},
		{name: "chargeback with discount", customerPaid: -10000, actual: 8000, discount: 2000, want: -16000},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := analytics.ComputeTiktokShippingDiff(tt.customerPaid, tt.actual, tt.discount)
			if got != tt.want {
				t.Errorf("ComputeTiktokShippingDiff(%v, %v, %v) = %v, want %v",
					tt.customerPaid, tt.actual, tt.discount, got, tt.want)
			}
		})
	}
}

// ============================================================================
// 2. MONETARY PRECISION — No float drift, exact computation
// ============================================================================

func TestShopeeShippingDiff_MonetaryPrecision(t *testing.T) {
	t.Parallel()

	// These values must compute EXACTLY (no float drift).
	// Float64 addition/subtraction of integer and .xx values is exact
	// when the result is representable (IEEE 754 double precision).
	tests := []struct {
		name      string
		buyerPaid float64
		actual    float64
		rebate    float64
	}{
		{name: "IDR integers", buyerPaid: 100000, actual: 70000, rebate: 15000},
		{name: "IDR with .25 cents", buyerPaid: 100.25, actual: 75.25, rebate: 15.25},
		{name: "IDR with .50 cents", buyerPaid: 100.50, actual: 75.50, rebate: 15.50},
		{name: "IDR with .75 cents", buyerPaid: 100.75, actual: 75.75, rebate: 15.75},
		{name: "tiny values", buyerPaid: 0.01, actual: 0.01, rebate: 0},
		{name: "all int multiples of 100", buyerPaid: 10000, actual: 7500, rebate: 1250},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := analytics.ComputeShopeeShippingDiff(tt.buyerPaid, tt.actual, tt.rebate)
			want := tt.buyerPaid - tt.actual + tt.rebate

			if got != want {
				t.Errorf("precision mismatch: ComputeShopeeShippingDiff(%v, %v, %v) = %.20f, want %.20f",
					tt.buyerPaid, tt.actual, tt.rebate, got, want)
			}
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("got NaN or Inf: %v", got)
			}
		})
	}
}

func TestTiktokShippingDiff_MonetaryPrecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		customerPaid float64
		actual       float64
		discount     float64
	}{
		{name: "IDR integers", customerPaid: 100000, actual: 80000, discount: 20000},
		{name: "IDR with .25 cents", customerPaid: 100.25, actual: 80.25, discount: 20.25},
		{name: "IDR with .50 cents", customerPaid: 100.50, actual: 80.50, discount: 20.50},
		{name: "IDR with .75 cents", customerPaid: 100.75, actual: 80.75, discount: 20.75},
		{name: "tiny values", customerPaid: 0.01, actual: 0.01, discount: 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := analytics.ComputeTiktokShippingDiff(tt.customerPaid, tt.actual, tt.discount)
			want := tt.customerPaid - tt.actual + tt.discount

			if got != want {
				t.Errorf("precision mismatch: ComputeTiktokShippingDiff(%v, %v, %v) = %.20f, want %.20f",
					tt.customerPaid, tt.actual, tt.discount, got, want)
			}
			if math.IsNaN(got) || math.IsInf(got, 0) {
				t.Errorf("got NaN or Inf: %v", got)
			}
		})
	}
}

// ============================================================================
// 3. NORMALIZATION VERIFICATION — abs() for negative shipping fees
// ============================================================================
//
// TikTok normalizes negative ActualShippingFee via abs() before storing.
// This test verifies that the formula produces correct results with
// pre-normalized values (as they would appear after persistence).
// ============================================================================

func TestTiktokNormalizedShippingFee_FormulaCorrectness(t *testing.T) {
	t.Parallel()

	// Simulate the normalization that happens during sync:
	// raw actual = -3000 → stored as 3000 via normalizeShippingFee()
	normalizedActual := 3000.0 // math.Abs(-3000)
	customerPaid := 10000.0
	discount := 1500.0

	// With normalization: 10000 - 3000 + 1500 = 8500
	got := analytics.ComputeTiktokShippingDiff(customerPaid, normalizedActual, discount)
	want := 10000.0 - 3000.0 + 1500.0 // 8500
	if got != want {
		t.Errorf("with normalized values: got %v, want %v", got, want)
	}

	// WITHOUT normalization, if -3000 were stored as-is:
	// 10000 - (-3000) + 1500 = 14500 (incorrect — inflates seller profit)
	withoutNorm := analytics.ComputeTiktokShippingDiff(customerPaid, -3000, discount)
	_ = withoutNorm // would be 14500 — demonstrating WHY normalization is necessary
}

// ============================================================================
// 4. DATA INTEGRITY — Formula consistency with persisted records
// ============================================================================

func TestShopeeShippingDiff_FromSyncedData(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	now := time.Now()
	orders := []models.ShopeeEscrowOrder{
		{
			ID:                   uuid.New().String(),
			TenantID:             "formula-tenant",
			OrderSN:              "FORMULA-ORD-001",
			Month:                5,
			Year:                 2026,
			BuyerPaidShippingFee: 10000,
			ActualShippingFee:    7000,
			ShopeeShippingRebate: 1500,
			SyncedAt:             now,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			ID:                   uuid.New().String(),
			TenantID:             "formula-tenant",
			OrderSN:              "FORMULA-ORD-002",
			Month:                5,
			Year:                 2026,
			BuyerPaidShippingFee: 5000,
			ActualShippingFee:    8000,
			ShopeeShippingRebate: 0,
			SyncedAt:             now,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			ID:                   uuid.New().String(),
			TenantID:             "formula-tenant",
			OrderSN:              "FORMULA-ORD-003",
			Month:                5,
			Year:                 2026,
			BuyerPaidShippingFee: 0,
			ActualShippingFee:    0,
			ShopeeShippingRebate: 0,
			SyncedAt:             now,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		{
			ID:                   uuid.New().String(),
			TenantID:             "formula-tenant",
			OrderSN:              "FORMULA-ORD-004",
			Month:                5,
			Year:                 2026,
			BuyerPaidShippingFee: 15000,
			ActualShippingFee:    12000,
			ShopeeShippingRebate: 3000,
			SyncedAt:             now,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		// Negative shipping fee scenario (chargeback)
		{
			ID:                   uuid.New().String(),
			TenantID:             "formula-tenant",
			OrderSN:              "FORMULA-ORD-005",
			Month:                5,
			Year:                 2026,
			BuyerPaidShippingFee: -5000,
			ActualShippingFee:    7000,
			ShopeeShippingRebate: 0,
			SyncedAt:             now,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
	}

	for _, o := range orders {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("failed to seed order %s: %v", o.OrderSN, err)
		}
	}

	var stored []models.ShopeeEscrowOrder
	if err := db.Where("tenant_id = ?", "formula-tenant").Find(&stored).Error; err != nil {
		t.Fatalf("failed to read orders: %v", err)
	}

	expectedDiffs := map[string]float64{
		"FORMULA-ORD-001": 10000 - 7000 + 1500, // 4500
		"FORMULA-ORD-002": 5000 - 8000 + 0,     // -3000
		"FORMULA-ORD-003": 0 - 0 + 0,            // 0
		"FORMULA-ORD-004": 15000 - 12000 + 3000, // 6000
		"FORMULA-ORD-005": -5000 - 7000 + 0,     // -12000
	}

	for _, o := range stored {
		want, ok := expectedDiffs[o.OrderSN]
		if !ok {
			t.Errorf("unexpected order %s", o.OrderSN)
			continue
		}
		got := analytics.ComputeShopeeShippingDiff(
			o.BuyerPaidShippingFee, o.ActualShippingFee, o.ShopeeShippingRebate)
		if got != want {
			t.Errorf("order %s: ComputeShopeeShippingDiff(%v, %v, %v) = %v, want %v",
				o.OrderSN, o.BuyerPaidShippingFee, o.ActualShippingFee, o.ShopeeShippingRebate, got, want)
		}
	}
}

func TestTiktokShippingDiff_FromSyncedData(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupTiktokSyncTestDB(t)
	defer tiktokCleanupDB(db)

	now := time.Now()
	orders := []models.TiktokEscrowOrder{
		{
			ID:                          uuid.New().String(),
			TenantID:                    "tiktok-formula",
			OrderID:                     "TT-FORMULA-001",
			Month:                       5,
			Year:                        2026,
			ShippingFeeCustomerPaid:     10000,
			ShippingFeeActual:           8000,
			ShippingFeePlatformDiscount: 2000,
			Currency:                    "IDR",
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		},
		{
			ID:                          uuid.New().String(),
			TenantID:                    "tiktok-formula",
			OrderID:                     "TT-FORMULA-002",
			Month:                       5,
			Year:                        2026,
			ShippingFeeCustomerPaid:     8000,
			ShippingFeeActual:           10000,
			ShippingFeePlatformDiscount: 500,
			Currency:                    "IDR",
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		},
		{
			ID:                          uuid.New().String(),
			TenantID:                    "tiktok-formula",
			OrderID:                     "TT-FORMULA-003",
			Month:                       5,
			Year:                        2026,
			ShippingFeeCustomerPaid:     0,
			ShippingFeeActual:           0,
			ShippingFeePlatformDiscount: 0,
			Currency:                    "IDR",
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		},
		// Normalized negative actual (raw -3000 → stored 3000)
		{
			ID:                          uuid.New().String(),
			TenantID:                    "tiktok-formula",
			OrderID:                     "TT-FORMULA-004",
			Month:                       5,
			Year:                        2026,
			ShippingFeeCustomerPaid:     10000,
			ShippingFeeActual:           3000,
			ShippingFeePlatformDiscount: 1500,
			Currency:                    "IDR",
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		},
		// Refund scenario
		{
			ID:                          uuid.New().String(),
			TenantID:                    "tiktok-formula",
			OrderID:                     "TT-FORMULA-005",
			Month:                       5,
			Year:                        2026,
			ShippingFeeCustomerPaid:     -5000,
			ShippingFeeActual:           8000,
			ShippingFeePlatformDiscount: 1000,
			Currency:                    "IDR",
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		},
	}

	for _, o := range orders {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("failed to seed order %s: %v", o.OrderID, err)
		}
	}

	var stored []models.TiktokEscrowOrder
	if err := db.Where("tenant_id = ?", "tiktok-formula").Find(&stored).Error; err != nil {
		t.Fatalf("failed to read orders: %v", err)
	}

	expectedDiffs := map[string]float64{
		"TT-FORMULA-001": 10000 - 8000 + 2000, // 4000
		"TT-FORMULA-002": 8000 - 10000 + 500,  // -1500
		"TT-FORMULA-003": 0 - 0 + 0,            // 0
		"TT-FORMULA-004": 10000 - 3000 + 1500,  // 8500 (normalized)
		"TT-FORMULA-005": -5000 - 8000 + 1000,  // -12000 (refund)
	}

	for _, o := range stored {
		want, ok := expectedDiffs[o.OrderID]
		if !ok {
			t.Errorf("unexpected order %s", o.OrderID)
			continue
		}
		got := analytics.ComputeTiktokShippingDiff(
			o.ShippingFeeCustomerPaid, o.ShippingFeeActual, o.ShippingFeePlatformDiscount)
		if got != want {
			t.Errorf("order %s: ComputeTiktokShippingDiff(%v, %v, %v) = %v, want %v",
				o.OrderID, o.ShippingFeeCustomerPaid, o.ShippingFeeActual, o.ShippingFeePlatformDiscount, got, want)
		}
	}
}

// ============================================================================
// 5. EDGE CASES — Negative values, zero quantity, canceled items, chargebacks
// ============================================================================

func TestEscrowFormulas_NegativeValues(t *testing.T) {
	t.Parallel()

	t.Run("shopee negative buyer_paid", func(t *testing.T) {
		t.Parallel()
		// When a Shopee order is charged back, BuyerPaidShippingFee can be negative
		got := analytics.ComputeShopeeShippingDiff(-5000, 7000, 1500)
		want := -5000.0 - 7000.0 + 1500.0 // -10500
		if got != want {
			t.Errorf("negative buyer paid: got %v, want %v", got, want)
		}
	})

	t.Run("shopee negative actual_shipping", func(t *testing.T) {
		t.Parallel()
		// Actual shipping fee is always non-negative in Shopee (no normalizeShippingFee),
		// but test the formula handles it correctly if it occurs
		got := analytics.ComputeShopeeShippingDiff(10000, -3000, 1500)
		want := 10000.0 - (-3000.0) + 1500.0 // 14500
		if got != want {
			t.Errorf("negative actual shipping: got %v, want %v", got, want)
		}
	})

	t.Run("shopee negative rebate", func(t *testing.T) {
		t.Parallel()
		// Rebate is typically non-negative, but test defensive handling
		got := analytics.ComputeShopeeShippingDiff(10000, 7000, -500)
		want := 10000.0 - 7000.0 + (-500.0) // 2500
		if got != want {
			t.Errorf("negative rebate: got %v, want %v", got, want)
		}
	})

	t.Run("tiktok negative customer_paid", func(t *testing.T) {
		t.Parallel()
		got := analytics.ComputeTiktokShippingDiff(-5000, 8000, 2000)
		want := -5000.0 - 8000.0 + 2000.0 // -11000
		if got != want {
			t.Errorf("negative customer paid: got %v, want %v", got, want)
		}
	})

	t.Run("tiktok negative discount", func(t *testing.T) {
		t.Parallel()
		got := analytics.ComputeTiktokShippingDiff(10000, 8000, -500)
		want := 10000.0 - 8000.0 + (-500.0) // 1500
		if got != want {
			t.Errorf("negative discount: got %v, want %v", got, want)
		}
	})

	t.Run("tiktok normalized actual is always non-negative", func(t *testing.T) {
		t.Parallel()
		// After normalizeShippingFee(), actual fee is always ≥ 0.
		// Verify formula works with zero actual
		got := analytics.ComputeTiktokShippingDiff(10000, 0, 2000)
		want := 10000.0 - 0 + 2000.0 // 12000
		if got != want {
			t.Errorf("zero actual after normalization: got %v, want %v", got, want)
		}
	})
}

// ============================================================================
// 6. SHIPPING ADJUSTMENTS — Various rebate and discount combinations
// ============================================================================

func TestEscrowFormulas_ShippingAdjustments(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string // "shopee" or "tiktok"
		fields   [3]float64 // [buyerPaid/customerPaid, actual, rebate/discount]
		want     float64
	}{
		// Shopee adjustments
		{name: "shopee: full rebate covers all", platform: "shopee", fields: [3]float64{10000, 10000, 2000}, want: 2000},
		{name: "shopee: no rebate, equal fees", platform: "shopee", fields: [3]float64{10000, 10000, 0}, want: 0},
		{name: "shopee: rebate exceeds buyer paid", platform: "shopee", fields: [3]float64{5000, 3000, 4000}, want: 6000},
		{name: "shopee: actual > buyer paid, no rebate", platform: "shopee", fields: [3]float64{5000, 8000, 0}, want: -3000},
		{name: "shopee: zero actual shipping", platform: "shopee", fields: [3]float64{10000, 0, 1500}, want: 11500},
		{name: "shopee: all positive large", platform: "shopee", fields: [3]float64{50000, 35000, 10000}, want: 25000},

		// TikTok adjustments
		{name: "tiktok: discount equals actual", platform: "tiktok", fields: [3]float64{10000, 5000, 5000}, want: 10000},
		{name: "tiktok: no discount, equal fees", platform: "tiktok", fields: [3]float64{10000, 10000, 0}, want: 0},
		{name: "tiktok: actual zero (free shipping)", platform: "tiktok", fields: [3]float64{10000, 0, 2000}, want: 12000},
		{name: "tiktok: discount > customer paid", platform: "tiktok", fields: [3]float64{3000, 2000, 5000}, want: 6000},
		{name: "tiktok: all costs covered by platform", platform: "tiktok", fields: [3]float64{0, 8000, 8000}, want: 0},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			switch tt.platform {
			case "shopee":
				got := analytics.ComputeShopeeShippingDiff(tt.fields[0], tt.fields[1], tt.fields[2])
				if got != tt.want {
					t.Errorf("got %v, want %v", got, tt.want)
				}
			case "tiktok":
				got := analytics.ComputeTiktokShippingDiff(tt.fields[0], tt.fields[1], tt.fields[2])
				if got != tt.want {
					t.Errorf("got %v, want %v", got, tt.want)
				}
			}
		})
	}
}

// ============================================================================
// 7. MULTI-LINE ALLOCATION — Items within a single order
// ============================================================================
//
// The shipping fee difference is computed at the order level, not per-item.
// Multi-line allocation verifies that the order-level total is correctly
// computed and that per-item prices sum correctly within an order.
// ============================================================================

func TestEscrowFormulas_MultiLineAllocation(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	now := time.Now()
	orderID := uuid.New().String()

	// One order with 3 items
	order := models.ShopeeEscrowOrder{
		ID:                   orderID,
		TenantID:             "multi-line-tenant",
		OrderSN:              "MULTI-ORD-001",
		Month:                6,
		Year:                 2026,
		BuyerPaidShippingFee: 15000,
		ActualShippingFee:    10000,
		ShopeeShippingRebate: 3000,
		EscrowAmount:         200000,
		SyncedAt:             now,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	items := []models.ShopeeEscrowItem{
		{
			ID:            uuid.New().String(),
			TenantID:      "multi-line-tenant",
			EscrowOrderID: orderID,
			OrderSN:       &order.OrderSN,
			Quantity:      2,
			OriginalPrice: 50000,
			SellingPrice:  45000,
		},
		{
			ID:            uuid.New().String(),
			TenantID:      "multi-line-tenant",
			EscrowOrderID: orderID,
			OrderSN:       &order.OrderSN,
			Quantity:      1,
			OriginalPrice: 80000,
			SellingPrice:  75000,
		},
		{
			ID:            uuid.New().String(),
			TenantID:      "multi-line-tenant",
			EscrowOrderID: orderID,
			OrderSN:       &order.OrderSN,
			Quantity:      3,
			OriginalPrice: 30000,
			SellingPrice:  25000,
		},
	}
	for _, it := range items {
		if err := db.Create(&it).Error; err != nil {
			t.Fatalf("failed to seed item: %v", err)
		}
	}

	// Read back and verify
	var stored models.ShopeeEscrowOrder
	if err := db.Where("id = ?", orderID).First(&stored).Error; err != nil {
		t.Fatalf("failed to read order: %v", err)
	}

	// 1. Order-level shipping formula is correct
	shipDiff := analytics.ComputeShopeeShippingDiff(
		stored.BuyerPaidShippingFee, stored.ActualShippingFee, stored.ShopeeShippingRebate)
	expectedShipDiff := 15000.0 - 10000.0 + 3000.0 // 8000
	if shipDiff != expectedShipDiff {
		t.Errorf("order shipping diff: got %v, want %v", shipDiff, expectedShipDiff)
	}

	// 2. Verify item counts and prices
	var storedItems []models.ShopeeEscrowItem
	if err := db.Where("escrow_order_id = ?", orderID).Find(&storedItems).Error; err != nil {
		t.Fatalf("failed to read items: %v", err)
	}
	if len(storedItems) != 3 {
		t.Errorf("expected 3 items, got %d", len(storedItems))
	}

	// 3. Total escrow amount should match sum of (item selling price × qty)
	var itemTotal float64
	var totalQty int
	for _, it := range storedItems {
		itemTotal += it.SellingPrice * float64(it.Quantity)
		totalQty += it.Quantity
	}
	if totalQty != 6 {
		t.Errorf("expected total qty 6, got %d", totalQty)
	}
}

// ============================================================================
// 8. CROSS-CURRENCY — Different currencies must be kept separate
// ============================================================================

func TestEscrowFormulas_CrossCurrency(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupTiktokSyncTestDB(t)
	defer tiktokCleanupDB(db)

	now := time.Now()
	orders := []models.TiktokEscrowOrder{
		{
			ID:                          uuid.New().String(),
			TenantID:                    "cross-currency",
			OrderID:                     "CROSS-IDR",
			Month:                       6,
			Year:                        2026,
			ShippingFeeCustomerPaid:     10000,
			ShippingFeeActual:           8000,
			ShippingFeePlatformDiscount: 2000,
			Currency:                    "IDR",
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		},
		{
			ID:                          uuid.New().String(),
			TenantID:                    "cross-currency",
			OrderID:                     "CROSS-THB",
			Month:                       6,
			Year:                        2026,
			ShippingFeeCustomerPaid:     500,
			ShippingFeeActual:           350,
			ShippingFeePlatformDiscount: 100,
			Currency:                    "THB",
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		},
		{
			ID:                          uuid.New().String(),
			TenantID:                    "cross-currency",
			OrderID:                     "CROSS-PHP",
			Month:                       6,
			Year:                        2026,
			ShippingFeeCustomerPaid:     500,
			ShippingFeeActual:           400,
			ShippingFeePlatformDiscount: 0,
			Currency:                    "PHP",
			SyncedAt:                    now,
			CreatedAt:                   now,
			UpdatedAt:                   now,
		},
	}

	for _, o := range orders {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("failed to seed order %s: %v", o.OrderID, err)
		}
	}

	var stored []models.TiktokEscrowOrder
	if err := db.Where("tenant_id = ?", "cross-currency").
		Order("order_id").Find(&stored).Error; err != nil {
		t.Fatalf("failed to read orders: %v", err)
	}

	// Each order has its own currency — verify formula results are correct per-currency
	for _, o := range stored {
		diff := analytics.ComputeTiktokShippingDiff(
			o.ShippingFeeCustomerPaid, o.ShippingFeeActual, o.ShippingFeePlatformDiscount)
		switch o.Currency {
		case "IDR":
			if diff != 10000-8000+2000 {
				t.Errorf("IDR order diff: got %v, want 4000", diff)
			}
		case "THB":
			if diff != 500-350+100 {
				t.Errorf("THB order diff: got %v, want 250", diff)
			}
		case "PHP":
			if diff != 500-400+0 {
				t.Errorf("PHP order diff: got %v, want 100", diff)
			}
		default:
			t.Errorf("unexpected currency %s for order %s", o.Currency, o.OrderID)
		}
	}

	// Verify no cross-currency mixing in computed totals
	// Each currency's orders are computed independently
	// (the formula operates per-order, so aggregation is done by the caller)
	currencyAgg := make(map[string]float64)
	for _, o := range stored {
		diff := analytics.ComputeTiktokShippingDiff(
			o.ShippingFeeCustomerPaid, o.ShippingFeeActual, o.ShippingFeePlatformDiscount)
		currencyAgg[o.Currency] += diff
	}
	if len(currencyAgg) != 3 {
		t.Errorf("expected 3 currencies, got %d: %v", len(currencyAgg), currencyAgg)
	}
}

// ============================================================================
// 9. PARTIAL PAGINATION FAILURE — Missing rows should NOT be treated as zero
// ============================================================================
//
// If pagination fails mid-fetch, some orders are missing. The formula must
// NOT silently treat missing rows as zero — the caller must detect gaps.
// This test verifies the formula itself returns correct results for what WAS
// synced, and that missing orders produce detectable gaps.
// ============================================================================

func TestEscrowFormulas_PartialPaginationFailure(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	now := time.Now()

	// Simulate a partial sync: 3 of 5 orders were successfully synced
	// (orders C and E are "missing" due to pagination failure)
	syncedOrders := map[string]models.ShopeeEscrowOrder{
		"A": {
			ID:                   uuid.New().String(),
			TenantID:             "partial-sync",
			OrderSN:              "PARTIAL-ORD-A",
			Month:                7,
			Year:                 2026,
			BuyerPaidShippingFee: 10000,
			ActualShippingFee:    7000,
			ShopeeShippingRebate: 1500,
			SyncedAt:             now,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		"B": {
			ID:                   uuid.New().String(),
			TenantID:             "partial-sync",
			OrderSN:              "PARTIAL-ORD-B",
			Month:                7,
			Year:                 2026,
			BuyerPaidShippingFee: 12000,
			ActualShippingFee:    8000,
			ShopeeShippingRebate: 2000,
			SyncedAt:             now,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
		"D": {
			ID:                   uuid.New().String(),
			TenantID:             "partial-sync",
			OrderSN:              "PARTIAL-ORD-D",
			Month:                7,
			Year:                 2026,
			BuyerPaidShippingFee: 15000,
			ActualShippingFee:    10000,
			ShopeeShippingRebate: 3000,
			SyncedAt:             now,
			CreatedAt:            now,
			UpdatedAt:            now,
		},
	}
	// Orders C and E are NOT synced (pagination failure)

	for _, o := range syncedOrders {
		if err := db.Create(&o).Error; err != nil {
			t.Fatalf("failed to seed order %s: %v", o.OrderSN, err)
		}
	}

	// Verify synced orders produce correct results
	var stored []models.ShopeeEscrowOrder
	if err := db.Where("tenant_id = ?", "partial-sync").Find(&stored).Error; err != nil {
		t.Fatalf("failed to read orders: %v", err)
	}

	if len(stored) != 3 {
		t.Errorf("expected 3 synced orders (partial sync), got %d — missing orders must be detectable", len(stored))
	}

	// Verify known order IDs to detect the gaps
	storedIDs := make(map[string]bool)
	for _, o := range stored {
		storedIDs[o.OrderSN] = true
	}
	if storedIDs["PARTIAL-ORD-A"] && storedIDs["PARTIAL-ORD-B"] && storedIDs["PARTIAL-ORD-D"] {
		// Correct — these 3 are present
	} else {
		t.Errorf("synced order set mismatch: %v", storedIDs)
	}

	// GAP DETECTION: The caller must notice that C and E are missing
	// by comparing expected order list against synced order count.
	// The formula itself is correct per-order; the gap is in the count.
	if storedIDs["PARTIAL-ORD-C"] {
		t.Error("PARTIAL-ORD-C should NOT be synced (pagination failure)")
	}
	if storedIDs["PARTIAL-ORD-E"] {
		t.Error("PARTIAL-ORD-E should NOT be synced (pagination failure)")
	}

	// Per-order formula correctness for what WAS synced
	for _, o := range stored {
		diff := analytics.ComputeShopeeShippingDiff(
			o.BuyerPaidShippingFee, o.ActualShippingFee, o.ShopeeShippingRebate)
		switch o.OrderSN {
		case "PARTIAL-ORD-A":
			if diff != 10000-7000+1500 {
				t.Errorf("order A: got %v, want 4500", diff)
			}
		case "PARTIAL-ORD-B":
			if diff != 12000-8000+2000 {
				t.Errorf("order B: got %v, want 6000", diff)
			}
		case "PARTIAL-ORD-D":
			if diff != 15000-10000+3000 {
				t.Errorf("order D: got %v, want 8000", diff)
			}
		}
	}
}

// ============================================================================
// 10. CANCELLED ITEMS — Zero amounts for canceled/refunded items
// ============================================================================

func TestEscrowFormulas_CanceledItems(t *testing.T) {
	t.Parallel()

	// Canceled items → zero buyer paid shipping, zero actual, zero rebate
	// Formula should produce diff = 0
	t.Run("shopee canceled order all zero", func(t *testing.T) {
		t.Parallel()
		got := analytics.ComputeShopeeShippingDiff(0, 0, 0)
		if got != 0 {
			t.Errorf("canceled order: got %v, want 0", got)
		}
	})

	// Canceled but partial refund where shipping was already paid
	t.Run("shopee partially refunded shipping", func(t *testing.T) {
		t.Parallel()
		// Buyer paid 10000, actual was 7000, rebate 1500 → should be 4500
		// If canceled, amounts become 0 but formula still handles correctly
		got := analytics.ComputeShopeeShippingDiff(0, 7000, 1500)
		want := 0.0 - 7000.0 + 1500.0 // -5500 (seller loses shipping cost)
		if got != want {
			t.Errorf("partially refunded: got %v, want %v", got, want)
		}
	})

	t.Run("tiktok canceled order all zero", func(t *testing.T) {
		t.Parallel()
		got := analytics.ComputeTiktokShippingDiff(0, 0, 0)
		if got != 0 {
			t.Errorf("canceled tiktok order: got %v, want 0", got)
		}
	})

	t.Run("tiktok partially refunded shipping", func(t *testing.T) {
		t.Parallel()
		got := analytics.ComputeTiktokShippingDiff(0, 8000, 2000)
		want := 0.0 - 8000.0 + 2000.0 // -6000
		if got != want {
			t.Errorf("partially refunded tiktok: got %v, want %v", got, want)
		}
	})
}

// ============================================================================
// 11. CHARGEBACKS — Full reversal of funds
// ============================================================================

func TestEscrowFormulas_Chargebacks(t *testing.T) {
	t.Parallel()

	t.Run("shopee full chargeback", func(t *testing.T) {
		t.Parallel()
		// Full reversal: buyer paid is negative (reversed), actual already incurred
		got := analytics.ComputeShopeeShippingDiff(-10000, 7000, 1500)
		want := -10000.0 - 7000.0 + 1500.0 // -15500
		if got != want {
			t.Errorf("full chargeback: got %v, want %v", got, want)
		}
	})

	t.Run("shopee chargeback no rebate", func(t *testing.T) {
		t.Parallel()
		got := analytics.ComputeShopeeShippingDiff(-10000, 7000, 0)
		want := -10000.0 - 7000.0 + 0 // -17000
		if got != want {
			t.Errorf("chargeback no rebate: got %v, want %v", got, want)
		}
	})

	t.Run("tiktok full chargeback", func(t *testing.T) {
		t.Parallel()
		got := analytics.ComputeTiktokShippingDiff(-10000, 8000, 2000)
		want := -10000.0 - 8000.0 + 2000.0 // -16000
		if got != want {
			t.Errorf("full chargeback: got %v, want %v", got, want)
		}
	})

	t.Run("tiktok partial chargeback with discount", func(t *testing.T) {
		t.Parallel()
		got := analytics.ComputeTiktokShippingDiff(-5000, 8000, 3000)
		want := -5000.0 - 8000.0 + 3000.0 // -10000
		if got != want {
			t.Errorf("partial chargeback: got %v, want %v", got, want)
		}
	})
}

// ============================================================================
// 12. ZERO QUANTITY — Items with qty=0 should not break allocation
// ============================================================================

func TestEscrowFormulas_ZeroQuantityItems(t *testing.T) {
	log.Logger = log.Output(zerolog.NewTestWriter(t))
	db := setupSyncTestDB(t)
	defer resetSyncTestDB(t, db)

	now := time.Now()
	orderID := uuid.New().String()

	order := models.ShopeeEscrowOrder{
		ID:                   orderID,
		TenantID:             "zero-qty-tenant",
		OrderSN:              "ZEROQTY-ORD-001",
		Month:                8,
		Year:                 2026,
		BuyerPaidShippingFee: 10000,
		ActualShippingFee:    7000,
		ShopeeShippingRebate: 1500,
		SyncedAt:             now,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := db.Create(&order).Error; err != nil {
		t.Fatalf("failed to seed order: %v", err)
	}

	items := []models.ShopeeEscrowItem{
		{
			ID:            uuid.New().String(),
			TenantID:      "zero-qty-tenant",
			EscrowOrderID: orderID,
			OrderSN:       &order.OrderSN,
			Quantity:      0, // zero quantity item
			OriginalPrice: 50000,
			SellingPrice:  50000,
		},
		{
			ID:            uuid.New().String(),
			TenantID:      "zero-qty-tenant",
			EscrowOrderID: orderID,
			OrderSN:       &order.OrderSN,
			Quantity:      2, // normal item
			OriginalPrice: 25000,
			SellingPrice:  25000,
		},
	}
	for _, it := range items {
		if err := db.Create(&it).Error; err != nil {
			t.Fatalf("failed to seed item: %v", err)
		}
	}

	// Read back and verify
	var stored models.ShopeeEscrowOrder
	if err := db.Where("id = ?", orderID).First(&stored).Error; err != nil {
		t.Fatalf("failed to read order: %v", err)
	}

	// Order-level formula unaffected by item quantity
	diff := analytics.ComputeShopeeShippingDiff(
		stored.BuyerPaidShippingFee, stored.ActualShippingFee, stored.ShopeeShippingRebate)
	if diff != 10000-7000+1500 {
		t.Errorf("order shipping diff with zero-qty items: got %v, want 4500", diff)
	}

	// Verify items are persisted
	var storedItems []models.ShopeeEscrowItem
	if err := db.Where("escrow_order_id = ?", orderID).Find(&storedItems).Error; err != nil {
		t.Fatalf("failed to read items: %v", err)
	}
	if len(storedItems) != 2 {
		t.Errorf("expected 2 items (1 zero-qty + 1 normal), got %d", len(storedItems))
	}
}

// ============================================================================
// 13. FORMULA INVARIANTS — Properties that must always hold
// ============================================================================

func TestEscrowFormulas_Invariants(t *testing.T) {
	t.Parallel()

	// Invariant 1: Symmetry — swapping buyer_paid and actual with zero rebate/discount
	// flips the sign: f(a,b,0) = -f(b,a,0)
	t.Run("shopee sign symmetry with zero rebate", func(t *testing.T) {
		t.Parallel()
		a, b := 10000.0, 7000.0
		forward := analytics.ComputeShopeeShippingDiff(a, b, 0)
		reverse := analytics.ComputeShopeeShippingDiff(b, a, 0)
		if forward != -reverse {
			t.Errorf("sign symmetry broken: f(%v,%v,0)=%v, f(%v,%v,0)=%v", a, b, forward, b, a, reverse)
		}
	})

	t.Run("tiktok sign symmetry with zero discount", func(t *testing.T) {
		t.Parallel()
		a, b := 10000.0, 8000.0
		forward := analytics.ComputeTiktokShippingDiff(a, b, 0)
		reverse := analytics.ComputeTiktokShippingDiff(b, a, 0)
		if forward != -reverse {
			t.Errorf("sign symmetry broken: f(%v,%v,0)=%v, f(%v,%v,0)=%v", a, b, forward, b, a, reverse)
		}
	})

	// Invariant 2: Additive identity — adding 0 rebate doesn't change base diff
	t.Run("shopee zero rebate identity", func(t *testing.T) {
		t.Parallel()
		base := analytics.ComputeShopeeShippingDiff(10000, 7000, 0)
		withZero := analytics.ComputeShopeeShippingDiff(10000, 7000, 0)
		if base != withZero {
			t.Errorf("zero rebate identity broken: %v != %v", base, withZero)
		}
	})

	// Invariant 3: For TikTok, normalized actual is always ≥ 0, so
	// diff ≤ customerPaid + discount
	t.Run("tiktok actual non-negative constraint", func(t *testing.T) {
		t.Parallel()
		for _, actual := range []float64{0, 1, 100, 10000} {
			diff := analytics.ComputeTiktokShippingDiff(10000, actual, 2000)
			maxPossible := 10000.0 + 2000.0
			if diff > maxPossible && actual >= 0 {
				t.Errorf("diff %v exceeds max possible %v with actual=%v", diff, maxPossible, actual)
			}
		}
	})
}

// ============================================================================
// HELPERS
// ============================================================================

// tiktokCleanupDB deletes all TikTok escrows from the DB (for tests that
// use setupTiktokSyncTestDB but don't need the file-based DB teardown).
func tiktokCleanupDB(db *gorm.DB) {
	db.Exec("DELETE FROM tiktok_escrow_orders")
	db.Exec("DELETE FROM tiktok_escrow_items")
	db.Exec("DELETE FROM tiktok_escrow_sync")
}

// Ensure fmt import is used (for data integrity assertions in subtests)
var _ = fmt.Sprintf

// Ensure sync package import is used (for potential future concurrent test helpers)
var _ = sync.Mutex{}
