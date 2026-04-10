// Package analytics provides TikTok escrow sync service
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/services"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// TiktokEscrowSyncService handles syncing escrow data from TikTok API
type TiktokEscrowSyncService struct {
	base     BaseEscrowService
	tenantID string
}

// NewTiktokEscrowSyncService creates a new TikTok escrow sync service
func NewTiktokEscrowSyncService(db *gorm.DB, tenantID, dbPath string) *TiktokEscrowSyncService {
	return &TiktokEscrowSyncService{
		base:     NewBaseEscrowService(db, tenantID, dbPath),
		tenantID: tenantID,
	}
}

// SyncMonth syncs escrow data for a specific month from TikTok API
func (s *TiktokEscrowSyncService) SyncMonth(
	ctx context.Context,
	month, year int,
	forceResync bool,
) (*dto.SyncResultDTO, error) {
	log.Info().Msgf("[TiktokEscrowSync] Starting sync for %d-%02d, tenant: %s", year, month, s.tenantID)

	// Check if current month (cannot sync)
	now := time.Now()
	if month == int(now.Month()) && year == now.Year() {
		return nil, fmt.Errorf("cannot sync current month, wait until month ends")
	}

	tables := TiktokEscrowTables()

	// Smart retry: check for previous failures
	if !forceResync {
		existing, retryIDs := s.checkExistingSyncForRetry(ctx, tables, month, year)
		if existing != nil && retryIDs == nil {
			return &dto.SyncResultDTO{
				TotalOrders: existing.TotalOrders,
				Message:     "Already synced. Use forceResync to update.",
			}, nil
		}
		// If retryIDs != nil, fall through to normal sync logic
		// which will handle them via atomic upsert (safe for re-processing)
	}

	// Get TikTok client
	client, err := s.getTiktokClient()
	if err != nil {
		return nil, fmt.Errorf("get tiktok client: %w", err)
	}

	// Fetch completed orders for the month
	orders, err := s.fetchOrdersByMonth(ctx, client, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch orders: %w", err)
	}

	if len(orders) == 0 {
		return &dto.SyncResultDTO{
			TotalOrders: 0,
			Message:     "No completed orders found for this period",
		}, nil
	}

	log.Info().Msgf("[TiktokEscrowSync] Found %d completed orders", len(orders))

	// Delete existing data if force resync
	if forceResync {
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return nil, err
		}
	}

	// Process orders
	totalItems, processedOrders, failedOrders := s.processOrders(ctx, client, orders, month, year)

	// Collect failed order IDs
	failedIDs := s.collectFailedOrderIDsFromAll(ctx, client, orders, month, year)

	// Save sync record with failed tracking
	if err := s.saveSyncRecord(ctx, tables, month, year, processedOrders, failedOrders, failedIDs); err != nil {
		return nil, err
	}

	return &dto.SyncResultDTO{
		TotalOrders:  processedOrders,
		TotalItems:   totalItems,
		FailedOrders: failedOrders,
		Message: fmt.Sprintf("Synced %d/%d orders (%d failed), %d items from TikTok API",
			processedOrders, len(orders), failedOrders, totalItems),
	}, nil
}

// getTiktokClient creates TikTok client with tenant credentials
func (s *TiktokEscrowSyncService) getTiktokClient() (*tiktokPkg.Client, error) {
	credService := services.NewCredentialService(s.base.DBPath)
	creds, err := credService.GetPlatformCredentials(s.tenantID, "tiktok")
	if err != nil {
		return nil, err
	}

	if creds.AppKey == "" || creds.AppSecret == "" {
		return nil, fmt.Errorf("tiktok credentials not configured for tenant")
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)
	return client, nil
}

// fetchOrdersByMonth fetches completed orders for a month
func (s *TiktokEscrowSyncService) fetchOrdersByMonth(
	ctx context.Context,
	client *tiktokPkg.Client,
	month, year int,
) ([]tiktokPkg.TiktokOrder, error) {
	var allOrders []tiktokPkg.TiktokOrder

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	req := tiktokPkg.OrderSearchRequest{
		OrderStatus:  "COMPLETED",
		CreateTimeGe: startDate.Unix(),
		CreateTimeLt: endDate.Unix(),
	}

	pageToken := ""
	pageCount := 0

	for {
		resp, err := client.SearchOrders(req, 100, pageToken)
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("TikTok API error: %d - %s", resp.Code, resp.Message)
		}

		allOrders = append(allOrders, resp.Data.Orders...)
		pageCount++
		log.Info().Msgf("[TiktokEscrowSync] Fetched page %d: %d orders", pageCount, len(resp.Data.Orders))

		pageToken = resp.Data.NextPageToken
		if pageToken == "" {
			break
		}

		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return allOrders, nil
}

// processOrders processes all orders and returns counts
func (s *TiktokEscrowSyncService) processOrders(
	ctx context.Context,
	client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder,
	month, year int,
) (totalItems, processedOrders, failedOrders int) {
	// Enrich orders with full payment details (shipping fees)
	s.enrichOrdersWithDetails(ctx, client, orders)

	for _, order := range orders {
		orderID := order.ID
		transaction, err := s.fetchOrderTransaction(ctx, client, orderID)
		if err != nil {
			log.Warn().Str("order_id", orderID).Str("tenant_id", s.tenantID).
				Err(err).Msg("[TiktokEscrowSync] Error fetching transaction")
			failedOrders++
			continue
		}

		if transaction == nil || transaction.Data.OrderID == "" {
			rawResp, _ := json.Marshal(transaction)
			log.Warn().Str("order_id", orderID).Str("tenant_id", s.tenantID).
				Str("raw_response", string(rawResp)).
				Msg("[TiktokEscrowSync] No transaction data")
			failedOrders++
			continue
		}

		itemsCount, err := s.saveEscrowOrder(ctx, order, transaction, month, year)
		if err != nil {
			log.Warn().Str("order_id", orderID).Str("tenant_id", s.tenantID).
				Err(err).Msg("[TiktokEscrowSync] Error saving order")
			failedOrders++
			continue
		}

		totalItems += itemsCount
		processedOrders++

		// Rate limiting (reduced from 500ms to avoid timeout on large months)
		time.Sleep(150 * time.Millisecond)
	}
	return totalItems, processedOrders, failedOrders
}

// fetchOrderTransaction fetches transaction details for an order
func (s *TiktokEscrowSyncService) fetchOrderTransaction(
	ctx context.Context,
	client *tiktokPkg.Client,
	orderID string,
) (*tiktokPkg.OrderTransactionResponse, error) {
	// Try v202501 API first
	resp, err := client.GetOrderTransactions(orderID)
	if err != nil {
		log.Warn().Str("order_id", orderID).Str("api_version", "v202501").
			Err(err).Msg("[TiktokEscrowSync] v202501 API failed, trying v202309")
		// Fallback to v202309
		resp, err = client.GetOrderTransactionsV202309(orderID)
		if err != nil {
			return nil, fmt.Errorf("both API versions failed for order %s: %w", orderID, err)
		}
	}

	if resp.Code != 0 {
		// Log the raw API error response for debugging
		rawResp, _ := json.Marshal(resp)
		log.Warn().Str("order_id", orderID).Int("api_code", resp.Code).
			Str("api_message", resp.Message).Str("raw_response", string(rawResp)).
			Msg("[TiktokEscrowSync] TikTok API returned error")
		return nil, fmt.Errorf("TikTok API error: %d - %s", resp.Code, resp.Message)
	}

	return resp, nil
}

// deleteMonthData deletes existing data for the month
func (s *TiktokEscrowSyncService) deleteMonthData(ctx context.Context, month, year int) error {
	tables := TiktokEscrowTables()
	return s.base.DeleteMonthDataGeneric(
		ctx, tables.OrderTable, tables.ItemTable, tables.SyncTable, month, year,
	)
}

// enrichOrdersWithDetails batch-fetches full order details from GetOrderDetail API
// to enrich orders with complete payment_info (shipping fees, total amounts, etc.)
// which are NOT provided by the SearchOrders API.
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

		// Build order_id -> detail map
		detailMap := make(map[string]*tiktokPkg.OrderDetailData, len(detail.Data.Orders))
		for j := range detail.Data.Orders {
			detailMap[detail.Data.Orders[j].ID] = &detail.Data.Orders[j]
		}

		// Enrich orders with payment info and line items
		for j := i; j < end; j++ {
			d, ok := detailMap[orders[j].ID]
			if !ok {
				continue
			}
			if d.PaymentInfo != nil {
				if orders[j].PaymentInfo.ShippingFee == "" {
					orders[j].PaymentInfo.ShippingFee = d.PaymentInfo.ShippingFee
				}
				if orders[j].PaymentInfo.TotalAmount == "" {
					orders[j].PaymentInfo.TotalAmount = d.PaymentInfo.TotalAmount
				}
				if orders[j].PaymentInfo.SubTotal == "" {
					orders[j].PaymentInfo.SubTotal = d.PaymentInfo.SubTotal
				}
				if orders[j].PaymentInfo.Currency == "" {
					orders[j].PaymentInfo.Currency = d.PaymentInfo.Currency
				}
			}
			// Enrich line items if originally empty
			if len(orders[j].LineItems) == 0 && len(d.LineItems) > 0 {
				for _, li := range d.LineItems {
					orders[j].LineItems = append(orders[j].LineItems, tiktokPkg.TiktokOrderItem{
						ID:               li.ID,
						SkuID:            li.SkuID,
						SkuName:          li.SkuName,
						ProductID:        li.ProductID,
						ProductName:      li.ProductName,
						SellerSku:        li.SellerSku,
						Quantity:         li.Quantity,
						OriginalPrice:    li.OriginalPrice,
						SalePrice:        li.SalePrice,
					})
				}
			}
		}

		time.Sleep(200 * time.Millisecond) // rate limiting
	}

	log.Info().Int("total_orders", len(orders)).
		Msg("[TiktokEscrowSync] Order detail enrichment completed")
}
