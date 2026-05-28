package analytics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ShopeeEscrowAPI defines the Shopee SDK methods needed for escrow sync.
// Exported for test injection.
type ShopeeEscrowAPI interface {
	GetWalletTransactions(req shopeePkg.GetWalletTransactionRequest) (*shopeePkg.WalletTransactionResponse, error)
	GetEscrowDetails(req shopeePkg.GetEscrowDetailsRequest) (*shopeePkg.GetEscrowDetailsResponse, error)
}

// WalletTx holds deduplicated wallet transaction data.
type WalletTx struct {
	OrderSN   string
	Amount    float64
	BuyerName string
}

// ShopeeEscrowSyncService implements jobs.EscrowSyncService for Shopee.
type ShopeeEscrowSyncService struct {
	systemDB  *gorm.DB
	tenantDB  *gorm.DB
	tenantID  string
	apiClient ShopeeEscrowAPI
}

// NewShopeeEscrowSyncService creates a new ShopeeEscrowSyncService.
func NewShopeeEscrowSyncService(systemDB *gorm.DB, tenantDB *gorm.DB, tenantID string) *ShopeeEscrowSyncService {
	return &ShopeeEscrowSyncService{
		systemDB: systemDB,
		tenantDB: tenantDB,
		tenantID: tenantID,
	}
}

// SetShopeeClient injects a ShopeeEscrowAPI client for testing.
func (s *ShopeeEscrowSyncService) SetShopeeClient(client ShopeeEscrowAPI) {
	s.apiClient = client
}

// getClient returns the injected client or errors (no silent fallback).
func (s *ShopeeEscrowSyncService) getClient(_ context.Context) (ShopeeEscrowAPI, error) {
	if s.apiClient != nil {
		return s.apiClient, nil
	}
	return nil, fmt.Errorf("shopee client not configured — inject via SetShopeeClient or configure tenant credentials")
}

// SyncMonthWithProgress implements jobs.EscrowSyncService.
// The Shopee API has a ~15-day wallet query window, so the sync chunks by 14 days,
// filters wallet_order_income + MONEY_IN, dedupes by OrderSN+Amount, then fetches
// escrow details in batches. Orders and items are saved transactionally.
func (s *ShopeeEscrowSyncService) SyncMonthWithProgress(ctx context.Context, month, year int, forceResync bool, onProgress func(processed, total int, message string)) error {
	log.Info().
		Str("tenant_id", s.tenantID).
		Int("month", month).
		Int("year", year).
		Bool("force_resync", forceResync).
		Msg("Shopee escrow sync started")

	onProgress(0, 100, "Starting sync...")

	if err := validateMonthYear(month, year); err != nil {
		onProgress(0, 100, fmt.Sprintf("Validation error: %v", err))
		return fmt.Errorf("validate month/year: %w", err)
	}

	if !forceResync {
		var existing models.ShopeeEscrowSync
		err := s.tenantDB.WithContext(ctx).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			First(&existing).Error
		if err == nil {
			log.Info().Str("tenant_id", s.tenantID).Int("month", month).Int("year", year).
				Int("total_orders", existing.TotalOrders).
				Msg("Shopee escrow already synced, skipping")
			onProgress(100, 100, fmt.Sprintf("Already synced (%d orders)", existing.TotalOrders))
			return nil
		}
	}

	onProgress(5, 100, "Getting Shopee client...")

	client, err := s.getClient(ctx)
	if err != nil {
		onProgress(0, 100, fmt.Sprintf("Client error: %v", err))
		return fmt.Errorf("get shopee client: %w", err)
	}

	onProgress(10, 100, "Fetching wallet transactions...")

	walletTx, err := s.getWalletTransactions(ctx, client, month, year)
	if err != nil {
		onProgress(0, 100, fmt.Sprintf("Wallet fetch error: %v", err))
		return fmt.Errorf("get wallet transactions: %w", err)
	}

	if len(walletTx) == 0 {
		log.Info().Str("tenant_id", s.tenantID).Int("month", month).Int("year", year).
			Msg("No wallet transactions found for period")
		onProgress(100, 100, "No transactions found")
		return nil
	}

	log.Info().Str("tenant_id", s.tenantID).Int("month", month).Int("year", year).
		Int("wallet_tx_count", len(walletTx)).
		Msg("Wallet transactions fetched")

	if forceResync {
		onProgress(15, 100, "Force resync: clearing previous data...")
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			onProgress(0, 100, fmt.Sprintf("Clear error: %v", err))
			return fmt.Errorf("clear previous data: %w", err)
		}
	}

	orderSNs := extractUniqueOrderSNs(walletTx)
	if len(orderSNs) == 0 {
		log.Info().Str("tenant_id", s.tenantID).Int("month", month).Int("year", year).
			Msg("No order SNs found in wallet transactions")
		onProgress(100, 100, "No order SNs found")
		return nil
	}

	log.Info().Str("tenant_id", s.tenantID).Int("month", month).Int("year", year).
		Int("unique_orders", len(orderSNs)).
		Msg("Processing orders")

	onProgress(20, 100, fmt.Sprintf("Processing %d orders in batches...", len(orderSNs)))
	totalItems, processedOrders, failedOrderIDs := processOrderBatches(
		ctx, client, orderSNs, s.tenantDB, s.tenantID, month, year, onProgress,
	)

	var failedIDsStr *string
	if len(failedOrderIDs) > 0 {
		joined := strings.Join(failedOrderIDs, ",")
		failedIDsStr = &joined
	}

	syncRecord := models.ShopeeEscrowSync{
		ID:            uuid.New().String(),
		TenantID:      s.tenantID,
		Month:         month,
		Year:          year,
		TotalOrders:   processedOrders,
		FailedOrders:  len(failedOrderIDs),
		FailedOrderIDs: failedIDsStr,
		SyncedAt:      time.Now(),
	}

	if err := s.tenantDB.WithContext(ctx).Create(&syncRecord).Error; err != nil {
		onProgress(0, 100, fmt.Sprintf("Sync record error: %v", err))
		return fmt.Errorf("save sync record: %w", err)
	}

	msg := fmt.Sprintf("Synced %d orders, %d items", processedOrders, totalItems)
	if len(failedOrderIDs) > 0 {
		msg += fmt.Sprintf(", %d failed", len(failedOrderIDs))
	}
	onProgress(100, 100, msg)

	log.Info().
		Str("tenant_id", s.tenantID).
		Int("month", month).
		Int("year", year).
		Int("total_orders", processedOrders).
		Int("total_items", totalItems).
		Int("failed_orders", len(failedOrderIDs)).
		Msg("Shopee escrow sync completed")
	return nil
}

func (s *ShopeeEscrowSyncService) deleteMonthData(ctx context.Context, month, year int) error {
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Delete(&models.ShopeeEscrowItem{}).Error; err != nil {
		return fmt.Errorf("delete items: %w", err)
	}
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Delete(&models.ShopeeEscrowOrder{}).Error; err != nil {
		return fmt.Errorf("delete orders: %w", err)
	}
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Delete(&models.ShopeeEscrowSync{}).Error; err != nil {
		return fmt.Errorf("delete sync records: %w", err)
	}
	return nil
}

// getWalletTransactions fetches wallet transactions for the given month/year,
// chunking by 14 days to respect the Shopee API query window limit.
func (s *ShopeeEscrowSyncService) getWalletTransactions(
	ctx context.Context,
	client ShopeeEscrowAPI,
	month, year int,
) ([]WalletTx, error) {
	firstDay := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	lastDay := firstDay.AddDate(0, 1, 0).Add(-time.Second)

	var allTx []WalletTx
	processedTx := make(map[string]bool)
	currentStart := firstDay

	for currentStart.Before(lastDay) || currentStart.Equal(lastDay) {
		currentEnd := currentStart.AddDate(0, 0, 14)
		if currentEnd.After(lastDay) {
			currentEnd = lastDay
		}

		chunkTx, err := fetchWalletChunk(ctx, client, currentStart, currentEnd)
		if err != nil {
			log.Warn().Err(err).
				Time("from", currentStart).Time("to", currentEnd).
				Msg("Shopee wallet chunk fetch failed, continuing")
			currentStart = currentEnd.Add(time.Second)
			continue
		}

		for _, tx := range chunkTx {
			key := fmt.Sprintf("%s_%f", tx.OrderSN, tx.Amount)
			if !processedTx[key] {
				processedTx[key] = true
				allTx = append(allTx, tx)
			}
		}

		currentStart = currentEnd.Add(time.Second)
	}

	return allTx, nil
}

// fetchWalletChunk fetches one paginated chunk of wallet transactions
// with wallet_order_income tab and MONEY_IN flow filter.
func fetchWalletChunk(
	_ context.Context,
	client ShopeeEscrowAPI,
	startDate, endDate time.Time,
) ([]WalletTx, error) {
	var allTx []WalletTx
	pageNo := 1
	pageSize := 100

	for {
		resp, err := client.GetWalletTransactions(shopeePkg.GetWalletTransactionRequest{
			PageNo:             pageNo,
			PageSize:           pageSize,
			StartDate:          startDate.Unix(),
			EndDate:            endDate.Unix(),
			TransactionTabType: "wallet_order_income",
			MoneyFlow:          "MONEY_IN",
		})
		if err != nil {
			return nil, fmt.Errorf("get wallet transactions page %d: %w", pageNo, err)
		}
		if resp.Error != "" {
			return nil, fmt.Errorf("shopee API error: %s - %s", resp.Error, resp.Message)
		}
		if len(resp.Response.TransactionList) == 0 {
			break
		}

		for _, tx := range resp.Response.TransactionList {
			allTx = append(allTx, WalletTx{
				OrderSN:   tx.OrderSN,
				Amount:    tx.Amount,
				BuyerName: tx.BuyerUsername,
			})
		}

		if !resp.Response.More {
			break
		}
		pageNo++
	}

	return allTx, nil
}

func extractUniqueOrderSNs(walletTx []WalletTx) []string {
	orderSNMap := make(map[string]bool)
	for _, tx := range walletTx {
		if tx.OrderSN != "" {
			orderSNMap[tx.OrderSN] = true
		}
	}
	orderSNs := make([]string, 0, len(orderSNMap))
	for sn := range orderSNMap {
		orderSNs = append(orderSNs, sn)
	}
	return orderSNs
}

const escrowBatchSize = 20

func processOrderBatches(
	ctx context.Context,
	client ShopeeEscrowAPI,
	orderSNs []string,
	db *gorm.DB,
	tenantID string,
	month, year int,
	onProgress func(processed, total int, message string),
) (totalItems, processedOrders int, failedOrderIDs []string) {
	total := len(orderSNs)
	for i := 0; i < total; i += escrowBatchSize {
		end := min(i+escrowBatchSize, total)
		batch := orderSNs[i:end]

		if onProgress != nil {
			pct := 20 + (75 * end / total)
			onProgress(pct, end, fmt.Sprintf("Processing batch %d/%d (%d orders)...",
				(i/escrowBatchSize)+1, (total+escrowBatchSize-1)/escrowBatchSize, len(batch)))
		}

		itemsSaved, failed := processBatch(ctx, client, batch, db, tenantID, month, year)
		totalItems += itemsSaved
		processedOrders += (len(batch) - len(failed))
		failedOrderIDs = append(failedOrderIDs, failed...)

		if end < total {
			select {
			case <-ctx.Done():
				return
			case <-time.After(500 * time.Millisecond):
			}
		}
	}
	return
}

func processBatch(
	ctx context.Context,
	client ShopeeEscrowAPI,
	orderSNs []string,
	db *gorm.DB,
	tenantID string,
	month, year int,
) (int, []string) {
	escrowResp, err := client.GetEscrowDetails(shopeePkg.GetEscrowDetailsRequest{
		OrderSNList: orderSNs,
	})
	if err != nil {
		log.Warn().Err(err).Strs("order_sns", orderSNs).
			Msg("Shopee escrow batch fetch failed")
		return 0, orderSNs
	}

	var totalItems int
	var failedIDs []string

	for _, order := range escrowResp.GetOrderList() {
		itemsCount, err := saveEscrowOrder(ctx, db, tenantID, order, month, year)
		if err != nil {
			log.Warn().Err(err).Str("order_sn", order.OrderSN).
				Msg("Shopee escrow save failed")
			failedIDs = append(failedIDs, order.OrderSN)
			continue
		}
		totalItems += itemsCount
	}

	respondedSNs := make(map[string]bool)
	for _, order := range escrowResp.GetOrderList() {
		respondedSNs[order.OrderSN] = true
	}
	for _, sn := range orderSNs {
		if !respondedSNs[sn] {
			failedIDs = append(failedIDs, sn)
		}
	}

	return totalItems, failedIDs
}
