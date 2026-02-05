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

func TestTiktokOrderRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.TiktokOrder{})
	repo := NewTiktokOrderRepository(db)
	ctx := context.Background()

	// Helper to create test order
	createTestOrder := func(t *testing.T, orderSN string, status string) *models.TiktokOrder {
		totalAmount := 200000.0
		shipByDate := time.Now().Add(48 * time.Hour).Unix()
		order := &models.TiktokOrder{
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
		order := createTestOrder(t, "TT-FIND001", "AWAITING_SHIPMENT")

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
		createTestOrder(t, "TT-FINDALL001", "AWAITING_SHIPMENT")
		createTestOrder(t, "TT-FINDALL002", "IN_TRANSIT")
		createTestOrder(t, "TT-FINDALL003", "DELIVERED")

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
			newOrder := &models.TiktokOrder{
				TenantID:    "test-tenant",
				OrderSN:     "TT-UPSERT001",
				OrderStatus: "UNPAID",
				Currency:    "IDR",
			}
			err := repo.Upsert(ctx, newOrder)
			assert.NoError(t, err)
			assert.NotZero(t, newOrder.ID)

			// Verify created
			found, err := repo.FindByOrderID(ctx, "TT-UPSERT001")
			assert.NoError(t, err)
			assert.Equal(t, "UNPAID", found.OrderStatus)
		})

		t.Run("updates existing order", func(t *testing.T) {
			order := createTestOrder(t, "TT-UPSERT002", "AWAITING_SHIPMENT")

			// Upsert with updated status
			updateOrder := &models.TiktokOrder{
				TenantID:       order.TenantID,
				OrderSN:        order.OrderSN,
				OrderStatus:    "IN_TRANSIT",
				TrackingNumber: "TT-TRK123456",
			}
			err := repo.Upsert(ctx, updateOrder)
			assert.NoError(t, err)

			// Verify updated
			found, err := repo.FindByOrderID(ctx, order.OrderSN)
			assert.NoError(t, err)
			assert.Equal(t, "IN_TRANSIT", found.OrderStatus)
			assert.Equal(t, "TT-TRK123456", found.TrackingNumber)
		})
	})
}
