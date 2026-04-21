// Package analytics provides TikTok escrow sync service
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"sync"
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

	// === PHASE 1: Statement-First Approach (accurate, matches Shopee pattern) ===
	// Fetch statement IDs for target month, then collect order settlements from each statement.
	// This is authoritative: settlement date comes from the statement, not order create/update time.
	if result := s.trySyncStatementFirst(ctx, client, tables, month, year, forceResync); result != nil {
		return result, nil
	}

	// === PHASE 2: Order-Centric Fallback (for sellers where statement tx API is unavailable) ===
	log.Info().Int("month", month).Int("year", year).
		Msg("[TiktokEscrowSync] Statement-first yielded no data, falling back to order-centric sync")

	// Fetch completed orders for the window (Target month + 60 days prior)
	orders, err := s.fetchOrdersForSettlementWindow(ctx, client, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch orders: %w", err)
	}

	if len(orders) == 0 {
		return &dto.SyncResultDTO{
			TotalOrders: 0,
			Message:     "No completed orders found for this period",
		}, nil
	}

	log.Info().Msgf("[TiktokEscrowSync] Found %d candidate orders in sync window", len(orders))

	// Delete existing data if force resync
	if forceResync {
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return nil, err
		}
	}

	// Process orders with high concurrency and settlement-based filtering
	totalItems, processedOrders, failedOrders := s.processOrdersConcurrent(ctx, client, orders, month, year)

	failedIDs := []string{}
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

// trySyncStatementFirst attempts the statement-based sync approach.
// Returns a non-nil result if statements exist and orders were found.
// Returns nil to signal the caller to fall back to order-centric sync.
func (s *TiktokEscrowSyncService) trySyncStatementFirst(
	ctx context.Context,
	client *tiktokPkg.Client,
	tables EscrowSyncTables,
	month, year int,
	forceResync bool,
) *dto.SyncResultDTO {
	// Step 1: List statement IDs for target month
	stmtIDs, err := s.fetchStatementIDs(ctx, client, month, year)
	if err != nil {
		log.Warn().Err(err).Msg("[TiktokEscrowSync] Statement ID fetch failed, will fallback")
		return nil
	}
	if len(stmtIDs) == 0 {
		log.Info().Int("month", month).Int("year", year).
			Msg("[TiktokEscrowSync] No statements found for month, will fallback")
		return nil
	}

	// Step 2: Collect order settlements from all statements
	orderData := s.collectOrdersFromStatements(ctx, client, stmtIDs)
	if len(orderData) == 0 {
		log.Warn().Int("stmts", len(stmtIDs)).
			Msg("[TiktokEscrowSync] Statements exist but no order data extracted (tx API may be unsupported), will fallback")
		return nil
	}

	log.Info().Int("orders", len(orderData)).
		Msg("[TiktokEscrowSync] Statement-first: orders with settlement found")

	// Step 3: Batch fetch full order details (20 per batch)
	orderIDs := orderDataKeys(orderData)
	orders := s.buildOrdersFromIDs(ctx, client, orderIDs)

	// Step 4: Delete existing data if force resync
	if forceResync {
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			log.Warn().Err(err).Msg("[TiktokEscrowSync] Failed to delete month data")
		}
	}

	// Step 5: Concurrent save with SKU-level enrichment
	totalItems, processed, failed := s.processStatementOrders(ctx, client, orders, orderData, month, year)

	if err := s.saveSyncRecord(ctx, tables, month, year, processed, failed, nil); err != nil {
		log.Warn().Err(err).Msg("[TiktokEscrowSync] Failed to save sync record")
	}

	return &dto.SyncResultDTO{
		TotalOrders:  processed,
		TotalItems:   totalItems,
		FailedOrders: failed,
		Message: fmt.Sprintf("[Statement-First] Synced %d orders (%d failed), %d items for %d/%02d",
			processed, failed, totalItems, year, month),
	}
}

// processStatementOrders concurrently saves orders sourced from statement API.
// Each order is enriched with SKU-level transactions before saving.
func (s *TiktokEscrowSyncService) processStatementOrders(
	ctx context.Context,
	client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder,
	orderData map[string]*OrderStatementData,
	month, year int,
) (totalItems, processed, failed int) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 5)

	for _, order := range orders {
		sd := orderData[order.ID]
		if sd == nil {
			continue
		}
		wg.Add(1)
		go func(o tiktokPkg.TiktokOrder, settlement *OrderStatementData) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Fetch SKU-level data for price analysis (best-effort)
			skuTx := s.fetchSkuTransactions(ctx, client, o.ID)

			items, err := s.saveOrderFromStatement(ctx, o, settlement, skuTx, month, year)
			if err != nil {
				log.Warn().Str("order_id", o.ID).Err(err).
					Msg("[TiktokEscrowSync] Failed to save statement order")
				mu.Lock()
				failed++
				mu.Unlock()
				return
			}
			if items > 0 {
				mu.Lock()
				totalItems += items
				processed++
				mu.Unlock()
			}
		}(order, sd)
	}

	wg.Wait()
	log.Info().Int("processed", processed).Int("failed", failed).Int("items", totalItems).
		Msg("[TiktokEscrowSync] Statement orders saved")
	return
}

// fetchSkuTransactions fetches SKU-level transaction data for price analysis.
// Tries v202501 first (preferred), falls back to v202309 for older orders.
func (s *TiktokEscrowSyncService) fetchSkuTransactions(
	ctx context.Context,
	client *tiktokPkg.Client,
	orderID string,
) []tiktokPkg.SkuTransaction {
	resp, err := client.GetOrderTransactions(orderID)
	if err == nil && resp.Code == 0 && len(resp.Data.SkuTransactions) > 0 {
		return resp.Data.SkuTransactions
	}
	// Fallback to v202309
	resp, err = client.GetOrderTransactionsV202309(orderID)
	if err == nil && resp.Code == 0 {
		return resp.Data.SkuTransactions
	}
	return nil
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

// fetchOrdersForSettlementWindow fetches completed orders for the target month plus a lookback buffer.
// Lookback = 21 days before month start: covers max TikTok settlement lag (T+14-21 days).
// Orders created earlier than this cannot settle in the target month.
// The Opsi A guard in saveEscrowOrder provides an additional safety check.
func (s *TiktokEscrowSyncService) fetchOrdersForSettlementWindow(
	ctx context.Context,
	client *tiktokPkg.Client,
	month, year int,
) ([]tiktokPkg.TiktokOrder, error) {
	var allOrders []tiktokPkg.TiktokOrder

	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := monthStart.AddDate(0, 1, 0)
	// Look back 21 days from month start (covers max settlement lag T+21)
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
// It filters orders so only those settled in the target month are saved.
func (s *TiktokEscrowSyncService) processOrdersConcurrent(
	ctx context.Context,
	client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder,
	month, year int,
) (totalItems, processedOrders, failedOrders int) {
	// 1. Batch enrich with full details (Buyer Name, Recipient Address, etc.)
	s.enrichOrdersWithDetails(ctx, client, orders)

	// 2. Parallel fetch transaction data and save if settled in target month
	var wg sync.WaitGroup
	var mu sync.Mutex
	sem := make(chan struct{}, 5) // Concurrency limit (reduced from 10 to avoid TikTok rate limits)

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

			// Save to DB (The save routine will check if SettlementTime matches month/year)
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

// fetchOrderTransaction fetches transaction details for an order.
// Uses v202309 first for statement_transactions (settlement timing),
// then supplements with v202501 for sku_transactions (SKU-level detail).
func (s *TiktokEscrowSyncService) fetchOrderTransaction(
	ctx context.Context,
	client *tiktokPkg.Client,
	orderID string,
) (*tiktokPkg.OrderTransactionResponse, error) {
	// Try v202309 first — reliably returns statement_transactions with settlement timing
	resp, err := client.GetOrderTransactionsV202309(orderID)
	if err != nil || resp.Code != 0 {
		log.Warn().Str("order_id", orderID).Str("api_version", "v202309").
			Err(err).Msg("[TiktokEscrowSync] v202309 API failed, trying v202501")
		resp, err = client.GetOrderTransactions(orderID)
		if err != nil {
			return nil, fmt.Errorf("both API versions failed for order %s: %w", orderID, err)
		}
	}

	if resp.Code != 0 {
		rawResp, _ := json.Marshal(resp)
		log.Warn().Str("order_id", orderID).Int("api_code", resp.Code).
			Str("api_message", resp.Message).Str("raw_response", string(rawResp)).
			Msg("[TiktokEscrowSync] TikTok API returned error")
		return nil, fmt.Errorf("TikTok API error: %d - %s", resp.Code, resp.Message)
	}

	// If v202309 has no sku_transactions, supplement with v202501 for SKU-level detail
	if len(resp.Data.SkuTransactions) == 0 {
		v2Resp, v2Err := client.GetOrderTransactions(orderID)
		if v2Err == nil && v2Resp.Code == 0 && len(v2Resp.Data.SkuTransactions) > 0 {
			resp.Data.SkuTransactions = v2Resp.Data.SkuTransactions
		}
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
			// Enrich recipient address (for buyer name)
			if d.RecipientAddress != nil && orders[j].RecipientAddress.Name == "" {
				orders[j].RecipientAddress = *d.RecipientAddress
			}
			// Enrich order status and create time
			if d.Status != "" && orders[j].Status == "" {
				orders[j].Status = d.Status
			}
			if d.CreateTime > 0 && orders[j].CreateTime == 0 {
				orders[j].CreateTime = d.CreateTime
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
