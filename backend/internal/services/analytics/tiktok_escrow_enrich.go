// Package analytics provides order enrichment and concurrent processing for TikTok escrow sync
package analytics

import (
	"context"
	"fmt"
	"sync"
	"time"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
)

// fetchOrdersForSettlementWindow fetches completed orders for the target month plus a lookback buffer.
// Lookback = 21 days before month start: covers max TikTok settlement lag (T+14-21 days).
func (s *TiktokEscrowSyncService) fetchOrdersForSettlementWindow(
	ctx context.Context,
	client *tiktokPkg.Client,
	month, year int,
) ([]tiktokPkg.TiktokOrder, error) {
	var allOrders []tiktokPkg.TiktokOrder

	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := monthStart.AddDate(0, 1, 0)
	startDate := monthStart.AddDate(0, 0, -21)

	req := tiktokPkg.OrderSearchRequest{
		OrderStatus:  "COMPLETED",
		CreateTimeGe: startDate.Unix(),
		CreateTimeLt: endDate.Unix(),
	}

	pageToken := ""
	for {
		resp, err := client.SearchOrders(req, 100, pageToken)
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("TikTok API error: %d - %s", resp.Code, resp.Message)
		}

		allOrders = append(allOrders, resp.Data.Orders...)
		pageToken = resp.Data.NextPageToken
		if pageToken == "" {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	return allOrders, nil
}

// processOrdersConcurrent processes orders in parallel using a worker pool.
func (s *TiktokEscrowSyncService) processOrdersConcurrent(
	ctx context.Context,
	client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder,
	month, year int,
) (totalItems, processedOrders, failedOrders int) {
	s.enrichOrdersWithDetails(ctx, client, orders)

	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 5)

	for _, order := range orders {
		wg.Add(1)
		go func(o tiktokPkg.TiktokOrder) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			transaction, err := s.fetchOrderTransaction(ctx, client, o.ID)
			if err != nil {
				log.Warn().Str("order_id", o.ID).Err(err).Msg("[TiktokEscrowSync] Failed to fetch transaction")
				mu.Lock()
				failedOrders++
				mu.Unlock()
				return
			}

			itemsCount, err := s.saveEscrowOrder(ctx, o, transaction, month, year)
			if err != nil {
				log.Warn().Str("order_id", o.ID).Err(err).Msg("[TiktokEscrowSync] Failed to save order")
				mu.Lock()
				failedOrders++
				mu.Unlock()
				return
			}

			if itemsCount > 0 {
				mu.Lock()
				totalItems += itemsCount
				processedOrders++
				mu.Unlock()
			}
		}(order)
	}

	wg.Wait()
	return totalItems, processedOrders, failedOrders
}

// enrichOrdersWithDetails batch-fetches full order details from GetOrderDetail API
// to enrich orders with complete payment_info, recipient address, and line items.
func (s *TiktokEscrowSyncService) enrichOrdersWithDetails(
	ctx context.Context,
	client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder,
) {
	const batchSize = 20
	for i := 0; i < len(orders); i += batchSize {
		end := i + batchSize
		if end > len(orders) {
			end = len(orders)
		}

		ids := make([]string, 0, end-i)
		for _, o := range orders[i:end] {
			ids = append(ids, o.ID)
		}

		detail, err := client.GetOrderDetail(ids)
		if err != nil {
			log.Warn().Err(err).Int("batch", i/batchSize).
				Msg("[TiktokEscrowSync] Failed to get order details batch, skipping enrichment")
			continue
		}

		detailMap := make(map[string]*tiktokPkg.OrderDetailData, len(detail.Data.Orders))
		for j := range detail.Data.Orders {
			detailMap[detail.Data.Orders[j].ID] = &detail.Data.Orders[j]
		}

		for j := i; j < end; j++ {
			d, ok := detailMap[orders[j].ID]
			if !ok {
				continue
			}
			enrichOrderFromDetail(&orders[j], d)
		}

		time.Sleep(200 * time.Millisecond)
	}

	log.Info().Int("total_orders", len(orders)).
		Msg("[TiktokEscrowSync] Order detail enrichment completed")
}

// enrichOrderFromDetail copies detail data into an order struct
func enrichOrderFromDetail(order *tiktokPkg.TiktokOrder, d *tiktokPkg.OrderDetailData) {
	if d.PaymentInfo != nil {
		if order.PaymentInfo.ShippingFee == "" {
			order.PaymentInfo.ShippingFee = d.PaymentInfo.ShippingFee
		}
		if order.PaymentInfo.TotalAmount == "" {
			order.PaymentInfo.TotalAmount = d.PaymentInfo.TotalAmount
		}
		if order.PaymentInfo.SubTotal == "" {
			order.PaymentInfo.SubTotal = d.PaymentInfo.SubTotal
		}
		if order.PaymentInfo.Currency == "" {
			order.PaymentInfo.Currency = d.PaymentInfo.Currency
		}
	}
	if d.RecipientAddress != nil && order.RecipientAddress.Name == "" {
		order.RecipientAddress = *d.RecipientAddress
	}
	if d.Status != "" && order.Status == "" {
		order.Status = d.Status
	}
	if d.CreateTime > 0 && order.CreateTime == 0 {
		order.CreateTime = d.CreateTime
	}
	if len(order.LineItems) == 0 && len(d.LineItems) > 0 {
		for _, li := range d.LineItems {
			order.LineItems = append(order.LineItems, tiktokPkg.TiktokOrderItem{
				ID:            li.ID,
				SkuID:         li.SkuID,
				SkuName:       li.SkuName,
				ProductID:     li.ProductID,
				ProductName:   li.ProductName,
				SellerSku:     li.SellerSku,
				Quantity:      li.Quantity,
				OriginalPrice: li.OriginalPrice,
				SalePrice:     li.SalePrice,
			})
		}
	}
}
