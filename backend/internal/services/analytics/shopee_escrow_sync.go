// Package analytics provides escrow sync service for Shopee
package analytics

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
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
	log.Info().Msgf("[ShopeeEscrowSync] Starting sync for %d-%02d, tenant: %s", year, month, s.tenantID)

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
		log.Info().Msgf("[ShopeeEscrowSync] ERROR getting shopee client: %v", err)
		return nil, fmt.Errorf("get shopee client: %w", err)
	}

	// Get wallet transactions for the month
	walletTx, err := s.getWalletTransactions(ctx, client, month, year)
	if err != nil {
		log.Info().Msgf("[ShopeeEscrowSync] ERROR getting wallet transactions: %v", err)
		return nil, fmt.Errorf("get wallet transactions: %w", err)
	}

	if len(walletTx) == 0 {
		return &dto.SyncResultDTO{
			TotalOrders: 0,
			Message:     "No transactions found for this period",
		}, nil
	}

	log.Info().Msgf("[ShopeeEscrowSync] Found %d wallet transactions", len(walletTx))

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

	log.Info().Msgf("[ShopeeEscrowSync] Unique orders with order_sn: %d", len(orderSNs))
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

		log.Info().Msgf("[ShopeeEscrowSync] Processing batch %d: %d orders", (i/batchSize)+1, len(batch))

		itemsCount, err := s.processBatch(ctx, client, batch, month, year)
		if err != nil {
			log.Info().Msgf("[ShopeeEscrowSync] Batch error: %v", err)
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
			log.Info().Msgf("[ShopeeEscrowSync] Error saving order %s: %v", order.OrderSN, err)
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
