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

func TestLazadaOrderRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.LazadaOrder{})
	repo := NewLazadaOrderRepository(db)
	ctx := context.Background()

	// Helper to create test order
	createTestOrder := func(t *testing.T, orderSN string, status string) *models.LazadaOrder {
		totalAmount := 150000.0
		shipByDate := time.Now().Add(24 * time.Hour).Unix()
		order := &models.LazadaOrder{
			TenantID:    "test-tenant",
			OrderSN:     orderSN,
			OrderStatus: status,
			TotalAmount: &totalAmount,
			Currency:    "IDR",
			ShipByDate:  &shipByDate,
		}
		err := db.Create(order).Error
		require.NoError(t, err)
		return order
	}

	t.Run("FindByOrderID", func(t *testing.T) {
		order := createTestOrder(t, "LZD-FIND001", "pending")

		tests := []struct {
			name    string
			orderID string
			wantErr bool
		}{
			{
				name:    "finds existing order",
				orderID: order.OrderSN,
				wantErr: false,
			},
			{
				name:    "returns error for non-existent order",
				orderID: "NONEXISTENT",
				wantErr: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindByOrderID(ctx, tt.orderID)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				require.NotNil(t, found)
				assert.Equal(t, tt.orderID, found.OrderSN)
				// Verify countdown is calculated
				assert.NotNil(t, found.Countdown)
			})
		}
	})

	t.Run("FindAll", func(t *testing.T) {
		// Create multiple orders
		createTestOrder(t, "LZD-FINDALL001", "pending")
		createTestOrder(t, "LZD-FINDALL002", "shipped")
		createTestOrder(t, "LZD-FINDALL003", "delivered")

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

	t.Run("Upsert", func(t *testing.T) {
		t.Run("creates new order", func(t *testing.T) {
			newOrder := &models.LazadaOrder{
				TenantID:    "test-tenant",
				OrderSN:     "LZD-UPSERT001",
				OrderStatus: "unpaid",
				Currency:    "IDR",
			}
			err := repo.Upsert(ctx, newOrder)
			assert.NoError(t, err)
			assert.NotZero(t, newOrder.ID)

			// Verify created
			found, err := repo.FindByOrderID(ctx, "LZD-UPSERT001")
			assert.NoError(t, err)
			assert.Equal(t, "unpaid", found.OrderStatus)
		})

		t.Run("updates existing order", func(t *testing.T) {
			order := createTestOrder(t, "LZD-UPSERT002", "pending")

			// Upsert with updated status
			updateOrder := &models.LazadaOrder{
				TenantID:       order.TenantID,
				OrderSN:        order.OrderSN,
				OrderStatus:    "shipped",
				TrackingNumber: "LZD-TRK123",
			}
			err := repo.Upsert(ctx, updateOrder)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindByOrderID(ctx, order.OrderSN)
			assert.NoError(t, err)
			assert.Equal(t, "shipped", found.OrderStatus)
			assert.Equal(t, "LZD-TRK123", found.TrackingNumber)
		})
	})
}
