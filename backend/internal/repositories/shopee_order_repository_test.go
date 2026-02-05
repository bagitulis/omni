package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShopeeOrderRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.ShopeeOrder{})
	repo := NewShopeeOrderRepository(db)
	ctx := context.Background()

	// Helper to create test order
	createTestOrder := func(t *testing.T, orderSN string, status string) *models.ShopeeOrder {
		totalAmount := 100000.0
		shipByDate := time.Now().Add(24 * time.Hour).Unix()
		order := &models.ShopeeOrder{
			TenantID:    "test-tenant",
			OrderSN:     orderSN,
			OrderStatus: status,
			TotalAmount: &totalAmount,
			Currency:    "IDR",
			ShipByDate:  &shipByDate,
		}
		err := repo.Create(ctx, order)
		require.NoError(t, err)
		return order
	}

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name    string
			order   *models.ShopeeOrder
			wantErr bool
		}{
			{
				name: "creates order successfully",
				order: &models.ShopeeOrder{
					TenantID:    "test-tenant",
					OrderSN:     "CREATE001",
					OrderStatus: "READY_TO_SHIP",
					Currency:    "IDR",
				},
				wantErr: false,
			},
			{
				name: "creates order with all fields",
				order: &models.ShopeeOrder{
					TenantID:        "test-tenant",
					OrderSN:         "CREATE002",
					OrderStatus:     "SHIPPED",
					TotalAmount:     ptrFloat64(150000),
					Currency:        "IDR",
					BuyerUsername:   "buyer123",
					PaymentMethod:   "COD",
					ShippingCarrier: "JNE",
					TrackingNumber:  "TRK123456",
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.Create(ctx, tt.order)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.NotZero(t, tt.order.ID)
				assert.False(t, tt.order.CreatedAt.IsZero())
			})
		}
	})

	t.Run("FindByOrderSN", func(t *testing.T) {
		order := createTestOrder(t, "FIND001", "READY_TO_SHIP")

		tests := []struct {
			name    string
			orderSN string
			wantErr bool
		}{
			{
				name:    "finds existing order",
				orderSN: order.OrderSN,
				wantErr: false,
			},
			{
				name:    "returns error for non-existent order",
				orderSN: "NONEXISTENT",
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindByOrderSN(ctx, tt.orderSN)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, tt.orderSN, found.OrderSN)
				// Verify countdown is calculated
				assert.NotNil(t, found.Countdown)
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		// Create multiple orders
		createTestOrder(t, "FINDALL001", "READY_TO_SHIP")
		createTestOrder(t, "FINDALL002", "SHIPPED")
		createTestOrder(t, "FINDALL003", "COMPLETED")

		tests := []struct {
			name     string
			page     int
			pageSize int
			wantMin  int
		}{
			{
				name:     "returns paginated results",
				page:     1,
				pageSize: 10,
				wantMin:  3,
			},
			{
				name:     "handles page size limit",
				page:     1,
				pageSize: 2,
				wantMin:  2,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				orders, total, err := repo.FindAll(ctx, tt.page, tt.pageSize)
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, int(total), tt.wantMin)
				assert.LessOrEqual(t, len(orders), tt.pageSize)
				// Verify countdown is calculated for orders with ship_by_date
				for _, order := range orders {
					if order.ShipByDate != nil {
						assert.NotNil(t, order.Countdown)
					}
				}
			})
		}
	})

	t.Run("Update", func(t *testing.T) {
		order := createTestOrder(t, "UPDATE001", "READY_TO_SHIP")

		order.OrderStatus = "SHIPPED"
		order.TrackingNumber = "NEWTRK123"

		err := repo.Update(ctx, order)
		assert.NoError(t, err)

		// Verify update
		found, err := repo.FindByOrderSN(ctx, order.OrderSN)
		assert.NoError(t, err)
		assert.Equal(t, "SHIPPED", found.OrderStatus)
		assert.Equal(t, "NEWTRK123", found.TrackingNumber)
	})

	t.Run("Upsert", func(t *testing.T) {
		t.Run("creates new order", func(t *testing.T) {
			newOrder := &models.ShopeeOrder{
				TenantID:    "test-tenant",
				OrderSN:     "UPSERT001",
				OrderStatus: "UNPAID",
				Currency:    "IDR",
			}
			err := repo.Upsert(ctx, newOrder)
			assert.NoError(t, err)
			assert.NotZero(t, newOrder.ID)

			// Verify created
			found, err := repo.FindByOrderSN(ctx, "UPSERT001")
			assert.NoError(t, err)
			assert.Equal(t, "UNPAID", found.OrderStatus)
		})

		t.Run("updates existing order", func(t *testing.T) {
			order := createTestOrder(t, "UPSERT002", "READY_TO_SHIP")

			// Upsert with updated status
			updateOrder := &models.ShopeeOrder{
				TenantID:       order.TenantID,
				OrderSN:        order.OrderSN,
				OrderStatus:    "SHIPPED",
				TrackingNumber: "UPSERTTRK",
			}
			err := repo.Upsert(ctx, updateOrder)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindByOrderSN(ctx, order.OrderSN)
			assert.NoError(t, err)
			assert.Equal(t, "SHIPPED", found.OrderStatus)
			assert.Equal(t, "UPSERTTRK", found.TrackingNumber)
		})
	})
}

// Helper functions
func ptrFloat64(v float64) *float64 {
	return &v
}
