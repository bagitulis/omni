package shopee

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"gorm.io/gorm"
)

// OrderSyncService handles syncing orders from Shopee API to database
type OrderSyncService struct {
	client    *shopeePkg.Client
	db        *gorm.DB
	orderRepo *repositories.ShopeeOrderRepository
	tenantID  string
}

// NewOrderSyncService creates a new order sync service with tenant ID
func NewOrderSyncService(client *shopeePkg.Client, db *gorm.DB, tenantID string) *OrderSyncService {
	return &OrderSyncService{
		client:    client,
		db:        db,
		orderRepo: repositories.NewShopeeOrderRepository(db),
		tenantID:  tenantID,
	}
}

// SyncOrders fetches orders from Shopee API and saves to database
func (s *OrderSyncService) SyncOrders(ctx context.Context, daysBack int) (int, error) {
	timeTo := time.Now().Unix()
	timeFrom := time.Now().AddDate(0, 0, -daysBack).Unix()

	// Get order list and handle pagination
	var allOrderSNs []string
	cursor := ""

	for {
		listResp, err := s.client.GetOrderList(timeFrom, timeTo, "create_time", "", cursor)
		if err != nil {
			return 0, err
		}

		if len(listResp.Response.OrderList) == 0 {
			break
		}

		for _, order := range listResp.Response.OrderList {
			allOrderSNs = append(allOrderSNs, order.OrderSN)
		}

		if !listResp.Response.More || listResp.Response.NextCursor == "" {
			break
		}
		cursor = listResp.Response.NextCursor
	}

	if len(allOrderSNs) == 0 {
		return 0, nil
	}

	// Get order details in batches of 50 (Shopee limit)
	count := 0
	for i := 0; i < len(allOrderSNs); i += 50 {
		end := i + 50
		if end > len(allOrderSNs) {
			end = len(allOrderSNs)
		}
		batchSNs := allOrderSNs[i:end]

		detailResp, err := s.client.GetOrderDetail(batchSNs)
		if err != nil {
			return 0, fmt.Errorf("get order detail batch: %w", err)
		}

		// Save to database
		for _, order := range detailResp.Response.OrderList {
			// Get ship by date - use from API or calculate from create_time + days_to_ship
			var shipByDate *int64
			if order.ShipByDate > 0 {
				shipByDate = &order.ShipByDate
			} else if order.DaysToShip > 0 && order.CreateTime > 0 {
				// Calculate: create_time + days_to_ship * 86400 (seconds per day)
				calculated := order.CreateTime + int64(order.DaysToShip*86400)
				shipByDate = &calculated
			}

			// Get shipping carrier - prefer actual carrier, fallback to checkout carrier
			shippingCarrier := order.ShippingCarrier
			if shippingCarrier == "" {
				shippingCarrier = order.CheckoutShippingCarrier
			}

			dbOrder := &models.ShopeeOrder{
				TenantID:        s.tenantID,
				OrderSN:         order.OrderSN,
				OrderStatus:     order.OrderStatus,
				TotalAmount:     &order.TotalAmount,
				Currency:        order.Currency,
				BuyerUsername:   order.BuyerUsername,
				PaymentMethod:   order.PaymentMethod,
				ShippingCarrier: shippingCarrier,
				BuyerMessage:    order.MessageToSeller,
				ShipByDate:      shipByDate,
			}

			if err := s.orderRepo.Upsert(ctx, dbOrder); err == nil {
				count++
			}
		}
	}

	return count, nil
}
