// Package analytics provides escrow sync service for Shopee
package analytics

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	"gorm.io/gorm"
)

// ShopeeEscrowSyncService handles syncing escrow data from Shopee API
type ShopeeEscrowSyncService struct {
	base     BaseEscrowService
	tenantID string
}

// NewShopeeEscrowSyncService creates a new escrow sync service
func NewShopeeEscrowSyncService(db *gorm.DB, tenantID, dbPath string) *ShopeeEscrowSyncService {
	return &ShopeeEscrowSyncService{
		base:     NewBaseEscrowService(db, tenantID, dbPath),
		tenantID: tenantID,
	}
}

// SyncMonth syncs escrow data for a specific month from Shopee API
func (s *ShopeeEscrowSyncService) SyncMonth(
	ctx context.Context,
	month, year int,
	forceResync bool,
) (*dto.SyncResultDTO, error) {
	log.Printf("[ShopeeEscrowSync] Starting sync for %d-%02d, tenant: %s", year, month, s.tenantID)

	// Check if current month (cannot sync)
	now := time.Now()
	if month == int(now.Month()) && year == now.Year() {
		return nil, fmt.Errorf("cannot sync current month, wait until month ends")
	}

	// Check if already synced
	tables := ShopeeEscrowTables()
	if !forceResync {
		var existing models.ShopeeEscrowSync
		err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			First(&existing).Error
		if err == nil {
			return &dto.SyncResultDTO{
				TotalOrders: existing.TotalOrders,
				Message:     "Already synced. Use forceResync to update.",
			}, nil
		}
	}

	// Get Shopee client
	client, err := s.getShopeeClient()
	if err != nil {
		log.Printf("[ShopeeEscrowSync] ERROR getting shopee client: %v", err)
		return nil, fmt.Errorf("get shopee client: %w", err)
	}

	// Get wallet transactions for the month
	walletTx, err := s.getWalletTransactions(ctx, client, month, year)
	if err != nil {
		log.Printf("[ShopeeEscrowSync] ERROR getting wallet transactions: %v", err)
		return nil, fmt.Errorf("get wallet transactions: %w", err)
	}

	if len(walletTx) == 0 {
		return &dto.SyncResultDTO{
			TotalOrders: 0,
			Message:     "No transactions found for this period",
		}, nil
	}

	log.Printf("[ShopeeEscrowSync] Found %d wallet transactions", len(walletTx))

	// Delete existing data if force resync
	if forceResync {
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return nil, err
		}
	}

	// Extract unique order SNs
	orderSNs := s.extractUniqueOrderSNs(walletTx)
	if len(orderSNs) == 0 {
		return &dto.SyncResultDTO{
			TotalOrders: 0,
			Message:     "No valid orders to process (all transactions were non-order types)",
		}, nil
	}

	// Process in batches
	totalItems, processedOrders := s.processOrderBatches(ctx, client, orderSNs, month, year)

	// Create sync record
	syncRecord := models.ShopeeEscrowSync{
		ID:          uuid.New().String(),
		TenantID:    s.tenantID,
		Month:       month,
		Year:        year,
		TotalOrders: processedOrders,
		SyncedAt:    time.Now(),
	}
	if err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).
		Create(&syncRecord).Error; err != nil {
		return nil, err
	}

	return &dto.SyncResultDTO{
		TotalOrders: processedOrders,
		TotalItems:  totalItems,
		Message:     fmt.Sprintf("Synced %d orders, %d items from Shopee API", processedOrders, totalItems),
	}, nil
}

// getShopeeClient creates Shopee client with tenant credentials
func (s *ShopeeEscrowSyncService) getShopeeClient() (*shopeePkg.Client, error) {
	credService := services.NewCredentialService(s.base.DBPath)
	creds, err := credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		return nil, err
	}

	if creds.PartnerID == 0 || creds.PartnerKey == "" {
		return nil, fmt.Errorf("shopee credentials not configured for tenant")
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)
	return client, nil
}

// extractUniqueOrderSNs extracts unique order SNs from wallet transactions
func (s *ShopeeEscrowSyncService) extractUniqueOrderSNs(walletTx []WalletTx) []string {
	orderSNMap := make(map[string]bool)
	for _, tx := range walletTx {
		if tx.OrderSN != "" {
			orderSNMap[tx.OrderSN] = true
		}
	}

	orderSNs := make([]string, 0, len(orderSNMap))
	for orderSN := range orderSNMap {
		orderSNs = append(orderSNs, orderSN)
	}

	log.Printf("[ShopeeEscrowSync] Unique orders with order_sn: %d", len(orderSNs))
	return orderSNs
}

// processOrderBatches processes orders in batches
func (s *ShopeeEscrowSyncService) processOrderBatches(
	ctx context.Context,
	client *shopeePkg.Client,
	orderSNs []string,
	month, year int,
) (totalItems, processedOrders int) {
	const batchSize = 20

	for i := 0; i < len(orderSNs); i += batchSize {
		end := i + batchSize
		if end > len(orderSNs) {
			end = len(orderSNs)
		}
		batch := orderSNs[i:end]

		log.Printf("[ShopeeEscrowSync] Processing batch %d: %d orders", (i/batchSize)+1, len(batch))

		itemsCount, err := s.processBatch(ctx, client, batch, month, year)
		if err != nil {
			log.Printf("[ShopeeEscrowSync] Batch error: %v", err)
			continue
		}
		totalItems += itemsCount
		processedOrders += len(batch)

		// Rate limiting
		if end < len(orderSNs) {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return totalItems, processedOrders
}

// processBatch processes a batch of orders
func (s *ShopeeEscrowSyncService) processBatch(
	ctx context.Context,
	client *shopeePkg.Client,
	orderSNs []string,
	month, year int,
) (int, error) {
	escrowResp, err := client.GetEscrowDetails(shopeePkg.GetEscrowDetailsRequest{
		OrderSNList: orderSNs,
	})
	if err != nil {
		return 0, err
	}

	var totalItems int
	for _, order := range escrowResp.GetOrderList() {
		itemsCount, err := s.saveEscrowOrder(ctx, order, month, year)
		if err != nil {
			log.Printf("[ShopeeEscrowSync] Error saving order %s: %v", order.OrderSN, err)
			continue
		}
		totalItems += itemsCount
	}

	return totalItems, nil
}

// deleteMonthData deletes existing data for the month
func (s *ShopeeEscrowSyncService) deleteMonthData(ctx context.Context, month, year int) error {
	tables := ShopeeEscrowTables()
	return s.base.DeleteMonthDataGeneric(
		ctx, tables.OrderTable, tables.ItemTable, tables.SyncTable, month, year,
	)
}

// SyncMonthWithProgress syncs escrow data for a specific month with progress callback
// This method is designed for background job execution and supports cancellation
func (s *ShopeeEscrowSyncService) SyncMonthWithProgress(
	ctx context.Context,
	month, year int,
	forceResync bool,
	onProgress ProgressCallback,
) (*SyncResultWithProgress, error) {
	log.Printf("[ShopeeEscrowSync] Starting sync with progress for %d-%02d, tenant: %s", year, month, s.tenantID)

	// Report initial progress
	if onProgress != nil {
		onProgress(0, 0, 0, "Initializing sync...")
	}

	// Check if current month (cannot sync)
	now := time.Now()
	if month == int(now.Month()) && year == now.Year() {
		return nil, fmt.Errorf("cannot sync current month, wait until month ends")
	}

	// Check if already synced
	tables := ShopeeEscrowTables()
	if !forceResync {
		var existing models.ShopeeEscrowSync
		err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			First(&existing).Error
		if err == nil {
			return &SyncResultWithProgress{
				TotalOrders:     existing.TotalOrders,
				ProcessedOrders: existing.TotalOrders,
				Message:         "Already synced. Use forceResync to update.",
			}, nil
		}
	}

	// Check cancellation
	select {
	case <-ctx.Done():
		return &SyncResultWithProgress{Cancelled: true, Message: "Cancelled before starting"}, ctx.Err()
	default:
	}

	// Report progress: getting credentials
	if onProgress != nil {
		onProgress(5, 0, 0, "Getting Shopee credentials...")
	}

	// Get Shopee client
	client, err := s.getShopeeClient()
	if err != nil {
		log.Printf("[ShopeeEscrowSync] ERROR getting shopee client: %v", err)
		return nil, fmt.Errorf("get shopee client: %w", err)
	}

	// Report progress: fetching transactions
	if onProgress != nil {
		onProgress(10, 0, 0, "Fetching wallet transactions...")
	}

	// Get wallet transactions for the month
	walletTx, err := s.getWalletTransactions(ctx, client, month, year)
	if err != nil {
		log.Printf("[ShopeeEscrowSync] ERROR getting wallet transactions: %v", err)
		return nil, fmt.Errorf("get wallet transactions: %w", err)
	}

	if len(walletTx) == 0 {
		return &SyncResultWithProgress{
			TotalOrders: 0,
			Message:     "No transactions found for this period",
		}, nil
	}

	log.Printf("[ShopeeEscrowSync] Found %d wallet transactions", len(walletTx))

	// Delete existing data if force resync
	if forceResync {
		if onProgress != nil {
			onProgress(15, 0, 0, "Deleting existing data...")
		}
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return nil, err
		}
	}

	// Extract unique order SNs
	orderSNs := s.extractUniqueOrderSNs(walletTx)
	if len(orderSNs) == 0 {
		return &SyncResultWithProgress{
			TotalOrders: 0,
			Message:     "No valid orders to process (all transactions were non-order types)",
		}, nil
	}

	// Report progress: starting batch processing
	if onProgress != nil {
		onProgress(20, 0, len(orderSNs), fmt.Sprintf("Processing %d orders...", len(orderSNs)))
	}

	// Process in batches with progress
	totalItems, processedOrders, cancelled := s.processOrderBatchesWithProgress(
		ctx, client, orderSNs, month, year, onProgress,
	)

	if cancelled {
		return &SyncResultWithProgress{
			TotalOrders:     len(orderSNs),
			ProcessedOrders: processedOrders,
			TotalItems:      totalItems,
			Cancelled:       true,
			Message:         fmt.Sprintf("Cancelled after processing %d/%d orders", processedOrders, len(orderSNs)),
		}, ctx.Err()
	}

	// Report progress: creating sync record
	if onProgress != nil {
		onProgress(95, processedOrders, len(orderSNs), "Creating sync record...")
	}

	// Create sync record
	syncRecord := models.ShopeeEscrowSync{
		ID:          uuid.New().String(),
		TenantID:    s.tenantID,
		Month:       month,
		Year:        year,
		TotalOrders: processedOrders,
		SyncedAt:    time.Now(),
	}
	if err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).
		Create(&syncRecord).Error; err != nil {
		return nil, err
	}

	// Report completion
	if onProgress != nil {
		onProgress(100, processedOrders, len(orderSNs), "Sync completed")
	}

	return &SyncResultWithProgress{
		TotalOrders:     len(orderSNs),
		ProcessedOrders: processedOrders,
		TotalItems:      totalItems,
		Message:         fmt.Sprintf("Synced %d orders, %d items from Shopee API", processedOrders, totalItems),
	}, nil
}

// processOrderBatchesWithProgress processes orders in batches with progress callback
func (s *ShopeeEscrowSyncService) processOrderBatchesWithProgress(
	ctx context.Context,
	client *shopeePkg.Client,
	orderSNs []string,
	month, year int,
	onProgress ProgressCallback,
) (totalItems, processedOrders int, cancelled bool) {
	const batchSize = 20
	totalOrders := len(orderSNs)

	for i := 0; i < len(orderSNs); i += batchSize {
		// Check for cancellation at start of each batch
		select {
		case <-ctx.Done():
			log.Printf("[ShopeeEscrowSync] Context cancelled, stopping at batch %d", (i/batchSize)+1)
			return totalItems, processedOrders, true
		default:
		}

		end := i + batchSize
		if end > len(orderSNs) {
			end = len(orderSNs)
		}
		batch := orderSNs[i:end]

		batchNum := (i / batchSize) + 1
		log.Printf("[ShopeeEscrowSync] Processing batch %d: %d orders", batchNum, len(batch))

		// Update progress (20-95% range for batch processing)
		if onProgress != nil {
			percent := 20 + (75 * (i + len(batch)) / totalOrders)
			onProgress(percent, processedOrders, totalOrders,
				fmt.Sprintf("Processing batch %d (%d/%d orders)...", batchNum, i+len(batch), totalOrders))
		}

		itemsCount, err := s.processBatch(ctx, client, batch, month, year)
		if err != nil {
			log.Printf("[ShopeeEscrowSync] Batch error: %v", err)
			continue
		}
		totalItems += itemsCount
		processedOrders += len(batch)

		// Rate limiting
		if end < len(orderSNs) {
			time.Sleep(500 * time.Millisecond)
		}
	}
	return totalItems, processedOrders, false
}
